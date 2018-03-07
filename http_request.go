package main

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

type HTTPRequest struct {
	HeadBuf    []byte
	connection *net.Conn
	Host       string
	Method     string
	URL        string
	hostOrURL  string
}

func (req *HTTPRequest) HTTPSReply() (err error) {
	_, err = fmt.Fprint(*req.connection, "HTTP/1.1 200 Connection established\r\n\r\n")
	return
}

func (req *HTTPRequest) HTTP() (err error) {
	req.URL = req.getHTTPURL()
	var u *url.URL
	u, err = url.Parse(req.URL)
	if err != nil {
		return
	}
	req.Host = u.Host
	req.addPortIfNot()
	return
}

func (req *HTTPRequest) HTTPS() (err error) {
	req.Host = req.hostOrURL
	req.addPortIfNot()
	return
}

func (req *HTTPRequest) addPortIfNot() (newHost string) {
	//newHost = req.Host
	port := "80"
	if req.IsHTTPS() {
		port = "443"
	}
	if (!strings.HasPrefix(req.Host, "[") && strings.Index(req.Host, ":") == -1) || (strings.HasPrefix(req.Host, "[") && strings.HasSuffix(req.Host, "]")) {
		//newHost = req.Host + ":" + port
		//req.headBuf = []byte(strings.Replace(string(req.headBuf), req.Host, newHost, 1))
		req.Host = req.Host + ":" + port
	}
	return
}

func (req *HTTPRequest) IsHTTPS() bool {
	return req.Method == "CONNECT"
}

func (req *HTTPRequest) getHTTPURL() (URL string) {
	if !strings.HasPrefix(req.hostOrURL, "/") {
		return req.hostOrURL
	}
	_host := req.getHeader("host")
	if _host == "" {
		return
	}
	URL = fmt.Sprintf("http://%s%s", _host, req.hostOrURL)
	return
}

func (req *HTTPRequest) getHeader(key string) (val string) {
	key = strings.ToUpper(key)
	lines := strings.Split(string(req.HeadBuf), "\r\n")
	//log.Println(lines)
	for _, line := range lines {
		line := strings.SplitN(strings.Trim(line, "\r\n "), ":", 2)
		if len(line) == 2 {
			k := strings.ToUpper(strings.Trim(line[0], " "))
			v := strings.Trim(line[1], " ")
			if key == k {
				val = v
				return
			}
		}
	}
	return
}
