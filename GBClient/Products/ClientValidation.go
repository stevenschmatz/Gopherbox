package main

import (
	"crypto/tls"
	"fmt"
	"github.com/stevenschmatz/gopherbox/GBClient/GBClientInit"
	"github.com/stevenschmatz/gopherbox/GBClient/GBClientNetworkTools"
	"log"
	"os"
	"time"
)

var (
	loginInfo string = os.Args[1]
	password  string = os.Args[2]
	tlsConn   *tls.Conn
)

// init generates the keypair and initializes the TLS connection.
func init() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s %s %s %s\n", os.Args[0], "<loginInfo: username/email>", "<password>")
		return
	}

	time.Sleep(100 * time.Millisecond)

	// create client public key and private key for secure session.
	if _, err := os.Stat(GBClientInit.CONFIG_FOLDER_NAME + "/" + GBClientInit.CLIENT_PUBLIC_KEY_FILENAME); os.IsNotExist(err) {
		GBClientInit.GenerateKeypair(loginInfo)

		// sleep so that the keypair has time to be generated properly
		time.Sleep(100 * time.Millisecond)
	}

	tlsConn = GBClientInit.InitConnection()
}

// main sends the server login info, and server sends back validation.
func main() {
	defer tlsConn.Close()

	GBClientNetworkTools.WriteContentsToConn(tlsConn, "IS_ALREADY_VERIFIED", "0")
	GBClientNetworkTools.WriteContentsToConn(tlsConn, "USERNAME", loginInfo)
	GBClientNetworkTools.WriteContentsToConn(tlsConn, "PASSWORD", password)

	_, verified_string, _, err := GBClientNetworkTools.GetValueFromContentsBlock(tlsConn)
	checkErr(err, tlsConn)

	fmt.Print(verified_string)
}

// checkErr exits if there is an error.
func checkErr(err error, conn *tls.Conn) {
	if err != nil {
		conn.Close()
		log.Fatalf("GBClient: GBClientValidation: %s", err)
	}
}
