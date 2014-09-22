package GBClientTermination

// package GBClientTermination implements the functions used by the client to terminate the binary.

import (
	"crypto/tls"
	"fmt"
	"github.com/stevenschmatz/gopherbox/GBClient/GBClientNetworkTools"
	"github.com/stevenschmatz/gopherbox/GBClient/GBClientWatch"
	"os"
	"time"
)

// CheckForProgramEnd checks for a kill signal from the Cocoa app (if the user quits, for example).
func CheckForProgramEnd(mustExit chan bool) {
	var input string
	fmt.Scanf("%v", &input)
	if input == "kill" {
		mustExit <- true
		go forceExitWithTimeout(5000 * time.Millisecond)
	}
}

// GracefullyExit is the default exit method for the binary.
func GracefullyExit(tlsConn *tls.Conn, key []byte) {

	// Kill command to server
	err := GBClientNetworkTools.SendDataEncrypted(tlsConn, GBClientWatch.OutputData{"kill", true, 1, 1}, GBClientNetworkTools.DISCONNECT_CONST, key)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}

	// Allow time to send request
	time.Sleep(200 * time.Millisecond)

	tlsConn.Close()
	os.Exit(1)
}

// forceExitWithTimeout forces the binary to exit with a timeout. Last resort to ending binary.
func forceExitWithTimeout(timeout time.Duration) {
	time.Sleep(timeout)
	os.Exit(1)
}
