package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type HTTP struct {
}

func (http *HTTP) OutToTCP(address string, inConn net.Conn, req *HTTPRequest) (err error) {
	inAddr := inConn.RemoteAddr().String()
	inLocalAddr := inConn.LocalAddr().String()
	//防止死循环
	if http.IsDeadLoop(inLocalAddr, req.Host) {
		inConn.Close()
		err = fmt.Errorf("dead loop detected , %s", req.Host)
		return
	}
	var outConn net.Conn
	//var _outConn interface{}
	tryCount := 0
	maxTryCount := 5
	for {
		outConn, err = net.DialTimeout("tcp", address, time.Duration(3000)*time.Millisecond)
		//outConn, err = utils.ConnectHost(s.Resolve(address), *s.cfg.Timeout)
		tryCount++
		if err == nil || tryCount > maxTryCount {
			break
		} else {
			log.Printf("err:%s,retrying...", err)
			time.Sleep(time.Second * 2)
		}
	}
	if err != nil {
		log.Printf("err:%s", err)
		inConn.Close()
		return
	}

	outAddr := outConn.RemoteAddr().String()
	err = req.HTTPSReply()

	//outLocalAddr := outConn.LocalAddr().String()
	http.IoBind(inConn, outConn, func(err interface{}) {
		log.Printf("conn %s - %s released [%s]", inAddr, outAddr, req.Host)
	}) //
	log.Printf("conn %s - %s connected [%s]", inAddr, outAddr, req.Host)

	return
}

func (http *HTTP) IoBind(dst io.ReadWriteCloser, src io.ReadWriteCloser, callback func(err interface{})) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("bind crashed %s", err)
			}
		}()
		e1 := make(chan interface{}, 1)
		e2 := make(chan interface{}, 1)
		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("bind crashed %s", err)
				}
			}()
			//_, err := io.Copy(dst, src)
			err := http.ioCopy(dst, src)
			e1 <- err
		}()
		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("bind crashed %s", err)
				}
			}()
			//_, err := io.Copy(src, dst)
			err := http.ioCopy(src, dst)
			e2 <- err
		}()
		var err interface{}
		select {
		case err = <-e1:
			//log.Printf("e1")
		case err = <-e2:
			//log.Printf("e2")
		}
		src.Close()
		dst.Close()
		callback(err)
	}()
}

func (http *HTTP) ioCopy(dst io.ReadWriter, src io.ReadWriter) (err error) {
	buf := make([]byte, 32*1024)
	n := 0
	for {
		n, err = src.Read(buf)
		if n > 0 {
			if _, e := dst.Write(buf[0:n]); e != nil {
				return e
			}
		}
		if err != nil {
			return
		}
	}
}

func (http *HTTP) IsDeadLoop(inLocalAddr string, host string) bool {
	inIP, inPort, err := net.SplitHostPort(inLocalAddr)
	if err != nil {
		return false
	}
	outDomain, outPort, err := net.SplitHostPort(host)
	if err != nil {
		return false
	}
	if inPort == outPort {
		var outIPs []net.IP
		outIPs, err = net.LookupIP(outDomain)
		if err == nil {
			for _, ip := range outIPs {
				if ip.String() == inIP {
					return true
				}
			}
		}
		interfaceIPs, err := http.GetAllInterfaceAddr()
		if err == nil {
			for _, localIP := range interfaceIPs {
				for _, outIP := range outIPs {
					if localIP.Equal(outIP) {
						return true
					}
				}
			}
		}
	}
	return false
}

func (http *HTTP) GetAllInterfaceAddr() ([]net.IP, error) {

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	addresses := []net.IP{}
	for _, iface := range ifaces {

		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		// if iface.Flags&net.FlagLoopback != 0 {
		//  continue // loopback interface
		// }
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// if ip == nil || ip.IsLoopback() {
			//  continue
			// }
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			addresses = append(addresses, ip)
		}
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("no address Found, net.InterfaceAddrs: %v", addresses)
	}
	//only need first
	return addresses, nil
}
