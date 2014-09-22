package main

import (
	"crypto/tls"
	"github.com/stevenschmatz/gopherbox/GBClient/GBClientInit"
	"github.com/stevenschmatz/gopherbox/GBClient/GBClientNetworkTools"
	"github.com/stevenschmatz/gopherbox/GBClient/GBClientRuntime"
	"github.com/stevenschmatz/gopherbox/GBClient/GBClientTermination"
	"github.com/stevenschmatz/gopherbox/GBClient/GBClientWatch"
	"log"
	"os"
	"time"
)

var (
	// mustExit communicates a "kill" command from the Cocoa application to be handled properly.
	mustExit = make(chan bool)

	// The command-line arguments for the binary.
	directoryPath, loginInfo, password = os.Args[1], os.Args[2], os.Args[3]

	// Connection information
	tlsConn       *tls.Conn
	key           []byte
	last_accessed int64

	// The main channels to communicate information to and from the file watcher.
	fileAddedChan, fileModifiedChan, fileDeletedChan chan GBClientWatch.OutputData
	fileMap                                          map[string]GBClientWatch.FileData
)

// init runs the initialization code for the client.
func init() {
	go GBClientTermination.CheckForProgramEnd(mustExit)

	GBClientInit.CheckOSArguments()
	tlsConn = GBClientInit.InitConnection()
	GBClientInit.CheckVerification(tlsConn, loginInfo, password)
	GBClientInit.ReceiveSessionInfoAndChangeDirectory(tlsConn, &key, &last_accessed, directoryPath, loginInfo)
	// GBClientInit.SendUnsyncedFiles(tlsConn, last_accessed, key)

	GBClientInit.InitChannelsAndFilemap(&fileAddedChan, &fileModifiedChan, &fileDeletedChan, &fileMap)
	GBClientRuntime.SendPing(tlsConn)
}

// main contents the contents of the main loop.
func main() {

	defer tlsConn.Close()

	for {
		GBClientWatch.CheckForChanges(fileMap, fileAddedChan, fileModifiedChan, fileDeletedChan)

		select {

		case <-mustExit:
			GBClientTermination.GracefullyExit(tlsConn, key)

		case fileAdded := <-fileAddedChan:
			log.Println(fileAdded)
			err := GBClientNetworkTools.SendDataEncrypted(tlsConn, fileAdded, GBClientNetworkTools.FILE_ADDED_CONST, key)
			checkErr(err, tlsConn)
		case fileModified := <-fileModifiedChan:
			err := GBClientNetworkTools.SendDataEncrypted(tlsConn, fileModified, GBClientNetworkTools.FILE_MODIFIED_CONST, key)
			checkErr(err, tlsConn)
		case fileDeletedChan := <-fileDeletedChan:
			err := GBClientNetworkTools.SendDataEncrypted(tlsConn, fileDeletedChan, GBClientNetworkTools.FILE_DELETED_CONST, key)
			checkErr(err, tlsConn)

		default:
			metadata, contents, operation, err := GBClientNetworkTools.
				ReadDataEncrypted(tlsConn, key)
			time.Sleep(100 * time.Millisecond)
			checkErr(err, tlsConn)
			if operation != GBClientNetworkTools.
				PING_CONST {
				GBClientRuntime.HandleDataChangesClient(metadata, &contents, operation)
			} else {
				GBClientRuntime.SendPing(tlsConn)
			}
		}
	}
}

// checkErr closes a connection and fatally logs the error if the error is not nil.
func checkErr(err error, conn *tls.Conn) {
	if err != nil {
		conn.Close()
		log.Fatalf("GBClient: Client: %s", err.Error())
	}
}
