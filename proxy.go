package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

const (
	ADDR = "0.0.0.0"

	PORT = 9390
)

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ADDR, PORT))
	if err != nil {
		log.Printf("listen error:%s", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept error:%s", err)
			// handle error
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) (req HTTPRequest, http HTTP, err error) {
	buf := make([]byte, 1024*4)
	len := 0
	len, err = conn.Read(buf[:])
	if err != nil {
		if err != io.EOF {
			err = fmt.Errorf("http decoder read err:%s", err)
		}
		conn.Close()
		return
	}

	req = HTTPRequest{
		HeadBuf:    buf[:len],
		connection: &conn,
	}

	index := bytes.IndexByte(req.HeadBuf, '\n')
	if index == -1 {
		err = fmt.Errorf("http decoder data line err:%s", SubStr(string(req.HeadBuf), 0, 50))
		conn.Close()
		return
	}
	fmt.Sscanf(string(req.HeadBuf[:index]), "%s%s", &req.Method, &req.hostOrURL)
	//}
	log.Printf("%s:%s:%s", req.Method, req.Host, req.hostOrURL)
	req.Method = strings.ToUpper(req.Method)
	if req.IsHTTPS() {
		err = req.HTTPS()
	} else {
		err = req.HTTP()
	}
	log.Printf("use proxy : %v, %s", false, req.Host)
	err = http.OutToTCP(req.Host, conn, &req)

	if err != nil {
		conn.Close()
	}

	return
}
