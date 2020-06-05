package main

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"flag"
	"fmt"
	"github.com/hashicorp/yamux"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
)

var sSession *yamux.Session


// Catches yamux connecting to us
func remoteListener(bindAddress, tlsCert, tlsKey string) {

	log.Println("Waiting for remote reverse connection client: ", bindAddress)

	cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	ln, err := tls.Listen("tcp", bindAddress, &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		log.Fatal("Error: ", err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		log.Println("Remote client connected: ", conn.RemoteAddr())

		// Add connection to yamux
		sSession, err = yamux.Client(conn, nil)
	}
}

// Catches clients and connects to yamux
func socksListener(bindAddress string) {

	log.Println("Waiting for socks client: ", bindAddress)

	ln, err := net.Listen("tcp", bindAddress)
	if err != nil {
		log.Fatal("Error: ",err)
	}

	for {

		conn, err := ln.Accept()
		if err != nil {
			log.Fatal("Error: ",err)
		}

		if sSession == nil {
			log.Printf("Rejecting incoming socks connection (%s), remote session not connected", conn.RemoteAddr())
			conn.Close()
			continue
		}

		log.Println("Got a socks client: ", conn.RemoteAddr())

		stream, err := sSession.Open()
		if err != nil {
			log.Fatal("Error: ",err)
		}

		// connect both of conn and stream

		var endWaiter sync.WaitGroup
		endWaiter.Add(2)

		go func() {
			defer conn.Close()
			defer endWaiter.Done()
			io.Copy(conn, stream)
		}()

		go func() {
			defer stream.Close()
			defer endWaiter.Done()
			io.Copy(stream, conn)
		}()

		go func() {
			endWaiter.Wait()
			log.Printf("Socks connection (%s) ended", conn.RemoteAddr())
		}()

	}
}



func printClientConfigKey(connectBackAddress, tlsCert string) {

	pemBytes, err := ioutil.ReadFile(tlsCert)
	if err != nil {
		log.Print("TLS cert error: ", err)
		log.Print("Have you generated a certificate pair? e.g. " +
			"$ openssl req -x509 -nodes -newkey rsa:4096 -keyout key.pem -out cert.pem -days 90")
		os.Exit(1)
	}
	block, _ := pem.Decode(pemBytes)

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Fatal("TLS cert error: ", err)
	}

	pubkey, _ := x509.MarshalPKIXPublicKey(cert.PublicKey)
	hash := sha256.Sum256(pubkey)

	var addrBytes []byte
	addrBytes = []byte(connectBackAddress)
	clientConfig := hex.EncodeToString(append(hash[:], addrBytes...))


	fmt.Printf("\nThe remote client needs to know where to connect back to, this is encoded in the following run key.\n" +
		"This key only changes if you change the TLS certs or connect back address." +
		"\n\n" +
		"\t%s\n\n" +
		"E.g. execute the remote client as follows:\n\n" +
		"\t./client %s\n\n\n", clientConfig, clientConfig)

}


func waitForCtrlC() { //https://jjasonclark.com/waiting_for_ctrl_c_in_golang/
	var endWaiter sync.WaitGroup
	endWaiter.Add(1)
	var signalChannel chan os.Signal
	signalChannel = make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	go func() {
		<-signalChannel
		endWaiter.Done()
	}()
	endWaiter.Wait()
}

func main() {

	listen := flag.String("remote-listener", "0.0.0.0:9999", "Bind address port for remote listener. E.g. 0.0.0.0:9999")
	socks := flag.String("socks-listener", "127.0.0.1:1080", "Bind address and port for socks listener. E.g. 127.0.0.1:1080")

	tlsCert := flag.String("tls-cert", "cert.pem", "Path to TLS certificate PEM. E.g. certs/cert.pem")
	tlsKey := flag.String("tls-key", "key.pem", "Path to TLS cert key. E.g. certs/key.pem")

	connectBack := flag.String("connect-back-address", "", "Address for the remote client to connect back to. E.g. 127.0.0.1:9999")

	flag.Usage = func() {
		fmt.Println("3gsocks server")
	}
	flag.Parse()

	if *connectBack == "" {
		flag.PrintDefaults()
		fmt.Print("\n--connect-back-address is required **\n\n")
		os.Exit(1)
	}

	printClientConfigKey(*connectBack, *tlsCert)

	log.Println("Starting server...")

	go remoteListener(*listen, *tlsCert, *tlsKey)
	go socksListener(*socks)

	waitForCtrlC()

}

