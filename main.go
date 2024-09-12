package main

import (
	"io"
	"log"
	"net"
	"os"
)

func handleConnection(src net.Conn, dstAddress string) {
	dst, err := net.Dial("tcp", dstAddress)
	if err != nil {
		log.Printf("Failed to connect to destination: %v", err)
		src.Close()
		return
	}
	defer dst.Close()

	go io.Copy(dst, src)
	io.Copy(src, dst)
}

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s [local address:port] [destination address:port]", os.Args[0])
	}

	listener, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		log.Fatalf("Failed to bind to address: %v", err)
	}
	defer listener.Close()

	log.Printf("Proxy listening on %s, forwarding to %s", os.Args[1], os.Args[2])

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		log.Printf("Accepted connection from %v, starting proxying...", conn.RemoteAddr().String())

		go handleConnection(conn, os.Args[2])
	}
}
