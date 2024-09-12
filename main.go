package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"tailscale.com/tsnet"
)

var (
	listeningAddr   = flag.String("l", ":1337", "Listening address")
	destinationAddr = flag.String("dst", "", "Destination address to be forwarded to")
	hostname        = flag.String("hn", "tsnet-revproxy", "Hostname to use on the tailnet")
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
	flag.Parse()

	if *destinationAddr == "" {
		log.Fatalln("destination address not specified, the program will quit.")
	}

	tsnetServer := new(tsnet.Server)
	tsnetServer.Hostname = *hostname
	tsnetServer.Ephemeral = true // ephemeral = remove from tailnet after inactivity period
	defer tsnetServer.Close()

	listener, err := tsnetServer.Listen("tcp", *listeningAddr)
	if err != nil {
		log.Fatalf("Failed to bind to address: %v", err)
	}
	defer listener.Close()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		log.Println("Shutting down gracefully...")
		listener.Close()
		tsnetServer.Close()
		os.Exit(0)
	}()

	log.Printf("Proxy listening on %s, forwarding to %s", *listeningAddr, *destinationAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		log.Printf("Accepted connection from %v, starting proxying...", conn.RemoteAddr().String())

		go handleConnection(conn, *destinationAddr)
	}
}
