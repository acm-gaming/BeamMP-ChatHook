package main

import (
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"
)

const (
	sharedClientPortStart = 3400
	sharedClientPortEnd   = 3499
)

type udpClient struct {
	conn *net.UDPConn
}

type udpListener struct {
	conn *net.UDPConn
}

func newUDPClient(ip string, port int) (*udpClient, error) {
	remoteIP := net.ParseIP(ip)
	if remoteIP == nil {
		return nil, fmt.Errorf("cannot connect to given address %s:%d", ip, port)
	}

	conn, err := dialSharedClientSocket(&net.UDPAddr{IP: remoteIP, Port: port})
	if err != nil {
		return nil, err
	}

	return &udpClient{conn: conn}, nil
}

func newUDPListener(ip string, port int) (*udpListener, error) {
	listenIP := net.ParseIP(ip)
	if listenIP == nil {
		return nil, fmt.Errorf("cannot bind to given address %s:%d", ip, port)
	}

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: listenIP, Port: port})
	if err != nil {
		return nil, fmt.Errorf("cannot bind to given address %s:%d", ip, port)
	}

	return &udpListener{conn: conn}, nil
}

func dialSharedClientSocket(remoteAddr *net.UDPAddr) (*net.UDPConn, error) {
	for port := sharedClientPortStart; port <= sharedClientPortEnd; port++ {
		conn, err := net.DialUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: port}, remoteAddr)
		if err == nil {
			return conn, nil
		}
	}
	return nil, fmt.Errorf("cannot bind to given address %s:%d", remoteAddr.IP.String(), remoteAddr.Port)
}

func (c *udpClient) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *udpClient) Send(data string) error {
	if c == nil || c.conn == nil {
		return errors.New("cannot send data")
	}
	_, err := c.conn.Write([]byte(data))
	if err != nil {
		return errors.New("cannot send data")
	}
	return nil
}

func (c *udpClient) RecvString() (string, bool, error) {
	if c == nil || c.conn == nil {
		return "", false, errors.New("cannot receive data")
	}

	buffer := make([]byte, 65507)
	if err := c.conn.SetReadDeadline(time.Now().Add(time.Millisecond)); err != nil {
		return "", false, errors.New("cannot receive data")
	}
	readBytes, err := c.conn.Read(buffer)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return "", false, nil
		}
		return "", false, errors.New("cannot receive data")
	}

	return string(buffer[:readBytes]), true, nil
}

func (l *udpListener) Close() error {
	if l == nil || l.conn == nil {
		return nil
	}
	return l.conn.Close()
}

func (l *udpListener) RecvString() (string, string, bool, error) {
	if l == nil || l.conn == nil {
		return "", "", false, errors.New("cannot receive data")
	}

	buffer := make([]byte, 65507)
	if err := l.conn.SetReadDeadline(time.Now().Add(time.Millisecond)); err != nil {
		return "", "", false, errors.New("cannot receive data")
	}
	readBytes, remoteAddr, err := l.conn.ReadFromUDP(buffer)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return "", "", false, nil
		}
		return "", "", false, errors.New("cannot receive data")
	}

	return string(buffer[:readBytes]), remoteAddr.String(), true, nil
}

func (l *udpListener) SendTo(data, ip string, port int) error {
	if l == nil || l.conn == nil {
		return errors.New("cannot send data")
	}

	remoteIP := net.ParseIP(ip)
	if remoteIP == nil {
		return errors.New("cannot send data")
	}

	if _, err := l.conn.WriteToUDP([]byte(data), &net.UDPAddr{IP: remoteIP, Port: port}); err != nil {
		return errors.New("cannot send data")
	}
	return nil
}

var nextHandle atomic.Uint64
