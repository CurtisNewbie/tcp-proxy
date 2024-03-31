package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	Tcp = "tcp"
)

var (
	ConnCount int64 = 0
)

type TcpConnHandler func(conn *net.TCPConn)

func Listen(host string, port int, handler TcpConnHandler) error {
	addr, err := net.ResolveTCPAddr(Tcp, fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP(Tcp, addr)
	if err != nil {
		return fmt.Errorf("failed to listen on tcp %v:%v, %w", host, port, err)
	}
	defer listener.Close()
	fmt.Printf("Server is listening on port %v\n", port)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		go handler(conn)
	}
}

func DialTcp(host string, port int) (*net.TCPConn, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}
	Debugf("Connected to proxy tcp %s:%d", host, port)
	return conn.(*net.TCPConn), nil
}

type Pipe struct {
	client  *net.TCPConn
	proxied *net.TCPConn
	wg      *sync.WaitGroup
	broke   int32
}

func (p *Pipe) Start() {
	p.wg = &sync.WaitGroup{}
	clientAddr := p.client.RemoteAddr().String()
	proxyAddr := p.proxied.RemoteAddr().String()
	p.pipeBetween(p.client, clientAddr, p.proxied, proxyAddr, func(conn *net.TCPConn) {
		// pipe may be blocked on proxied.Read, even when the client has been disconnected
		p.proxied.SetReadDeadline(time.Now().Add(time.Millisecond * 50))
	})
	p.pipeBetween(p.proxied, proxyAddr, p.client, clientAddr, nil)
}

func (p *Pipe) Wait() {
	p.wg.Wait()
}

func (p *Pipe) pipeBetween(dst *net.TCPConn, dstn string, src *net.TCPConn, srcn string, readDeadline func(conn *net.TCPConn)) {
	p.wg.Add(1)
	go func() {
		Logf("pipe %s -> %s started", srcn, dstn)

		defer func() {
			Logf("pipe %s -> %s stopped", srcn, dstn)
			atomic.StoreInt32(&p.broke, 1)
			p.wg.Done()
		}()

		buf := make([]byte, 8192)
		for {
			if atomic.LoadInt32(&p.broke) == 1 {
				return
			}

			if readDeadline != nil {
				readDeadline(src)
			}

			Debugf("pipe %s -> %s running", srcn, dstn)
			n, err := src.Read(buf)
			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					continue
				}
				Logf("failed to read from %s, %v, exit", srcn, err)
				return
			}
			Debugf("%s Read n: %d", srcn, n)

			m, err := dst.Write(buf[:n])
			if err != nil {
				Logf("failed to write to %s, %v, exit", dstn, err)
				return
			}
			Debugf("%s Write n: %d", dstn, m)

			Logf("%s -> %s (%d bytes):\n%s", srcn, dstn, n, string(buf[:n]))
		}
	}()
}

func NewPipe(client *net.TCPConn, proxied *net.TCPConn) *Pipe {
	return &Pipe{
		client:  client,
		proxied: proxied,
	}
}

func NewProxyHandler(proxied *net.TCPConn) TcpConnHandler {
	return func(conn *net.TCPConn) {
		swapped := atomic.CompareAndSwapInt64(&ConnCount, 0, 1)

		if !swapped {
			Logf("Connection already occupied, only supports one connection")
			conn.Close()
			return
		}
		Logf("Accept connection from %v", conn.RemoteAddr().String())

		defer conn.Close()
		defer func() {
			if *debug {
				Debugf("Connection for %v closed", conn.RemoteAddr().String())
			}
		}()
		defer func() { atomic.AddInt64(&ConnCount, -1) }()

		pipe := NewPipe(conn, proxied)
		pipe.Start()
		pipe.Wait()
		Debugf("Pipe closed")
	}
}
