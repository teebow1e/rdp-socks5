package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/net/proxy"
)

var lAddr = flag.String(`l`, `127.0.0.1:3388`, `Listen address`)

// var socksURI = flag.String(`x`, `socks5://127.0.0.1:1080?timeout=15m`, `Socks URI`)
var proxyAddr = flag.String(`r`, `10.10.10.10:3389`, `Remote address`)
var dialFunc func(string, string) (net.Conn, error)

func handleConn(lconn net.Conn) {
	rconn, err := dialFunc(`tcp`, *proxyAddr)
	if err != nil {
		lconn.Close()
		log.Println(err)
		return
	}
	log.Println(`Connected to`, *proxyAddr)
	defer func() {
		time.Sleep(time.Second)
		rconn.Close()
		lconn.Close()
	}()
	go io.Copy(lconn, rconn)
	io.Copy(rconn, lconn)
	log.Println(`Closed:`, lconn.RemoteAddr().String())
}

func checkSOCKS5Proxy(dialer proxy.Dialer) error {
	testAddr := "8.8.8.8:53"
	conn, err := dialer.Dial("tcp", testAddr)
	if err != nil {
		return fmt.Errorf("proxy connection failed: %w", err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	return nil
}

func main() {
	flag.Parse()
	socks5ProxyAddr := "xxxx:1080"
	username := "hehe"
	password := "cc"

	auth := &proxy.Auth{
		User:     username,
		Password: password,
	}

	proxyDialer, err := proxy.SOCKS5("tcp", socks5ProxyAddr, auth, proxy.Direct)
	if err != nil {
		fmt.Printf("Error creating SOCKS5 dialer: %v\n", err)
		os.Exit(1)
	}

	err = checkSOCKS5Proxy(proxyDialer)
	if err != nil {
		log.Println(err)
	}

	dialFunc = proxyDialer.Dial
	srv, err := net.Listen(`tcp`, *lAddr)
	if err != nil {
		log.Panicln(err)
	}
	// log.Println(`Proxing`, *lAddr, `to`, *proxyAddr, `via`, *socksURI)
	log.Println(`Proxing`, *lAddr, `to`, *proxyAddr, `via hardcoded socks`)
	for {
		conn, err := srv.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println(`Got connection from`, conn.RemoteAddr().String())
		go handleConn(conn)
	}
}
