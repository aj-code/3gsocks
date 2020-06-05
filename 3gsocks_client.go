package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	socks5 "github.com/armon/go-socks5"
	"github.com/hashicorp/yamux"
	"log"
	"net"
	"os"
)

var cSession *yamux.Session

func connect(address string, certFingerprint []byte) {
	server, err := socks5.New(&socks5.Config{})
	if err != nil {
		log.Fatal(err)
	}

	conn, err := tls.DialWithDialer(&net.Dialer{}, "tcp", address, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		log.Fatal(err)
	}

	//do cert pinning
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) != 1 {
		log.Fatal("Unexpected number of TLS certs detected, possible MITM")
	}

	peercert := certs[0]
	pubkey, _ := x509.MarshalPKIXPublicKey(peercert.PublicKey)
	hash := sha256.Sum256(pubkey)
	if bytes.Compare(certFingerprint, hash[:]) != 0 {
		log.Fatal("Unexpected TLS cert public key detected, possible server cert change or MITM")
	}

	//setup tcp streams
	cSession, err = yamux.Server(conn, nil)
	if err != nil {
		log.Fatal(err)
	}

	for {
		stream, err := cSession.Accept()
		if err != nil {
			log.Fatal(err)

		}

		//pass off to socks server
		go func() {
			err = server.ServeConn(stream)
			if err != nil {
				log.Print(err)
			}
		}()
	}
}



func main() {

	if len(os.Args) != 2 {
		log.Fatal("Error, missing run key argument")
	}

	runKey := os.Args[1]
	keyBytes, err := hex.DecodeString(runKey)
	if err != nil || len(keyBytes) <= 32 {
		log.Fatal("Error, config argument not valid hex or too short")
	}

	certFingerprint := keyBytes[:32]
	remote := string(keyBytes[32:])

	connect(remote, certFingerprint)

}
