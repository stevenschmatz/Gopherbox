package GBServerInit

import (
	"crypto/rand"
	"crypto/tls"
	"log"
	"net"
)

const (
	PORT = "8000"
)

func InitListener() net.Listener {
	cert, err := tls.LoadX509KeyPair("config/server.pem", "config/server.key")
	if err != nil {
		checkErr(err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}, ClientAuth: tls.RequireAnyClientCert}
	config.Rand = rand.Reader
	service := "0.0.0.0:" + PORT
	listener, err := tls.Listen("tcp", service, &config)
	if err != nil {
		checkErr(err)
	}
	log.Printf("\n======================\nGBServer: Listening...\n======================\n")
	return listener
}

func checkErr(err error) {
	if err != nil {
		log.Fatalf("GBServer: GBServerInit: %s", err)
	}
}
