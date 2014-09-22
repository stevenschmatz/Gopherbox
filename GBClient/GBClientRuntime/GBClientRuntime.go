package GBClientRuntime

// package GBClientRuntime contains the code used by the client binary file.

import (
	"crypto/tls"
	"github.com/stevenschmatz/gopherbox/GBClient/GBClientNetworkTools"
	"io"
	"io/ioutil"
	"os"
)

// HandleDataChangesClient makes or deletes the appropriate files as requested by the server.
func HandleDataChangesClient(metadata GBClientNetworkTools.FileMetadata, contents *string, operation int) {
	switch {
	case operation == GBClientNetworkTools.FILE_ADDED_CONST:
		if metadata.Is_dir {
			os.MkdirAll(metadata.Path, 0700)
		} else {
			ioutil.WriteFile(metadata.Path, []byte(*contents), 0700)
		}
	case operation == GBClientNetworkTools.FILE_MODIFIED_CONST:
		if !metadata.Is_dir {
			os.RemoveAll(metadata.Path)
			ioutil.WriteFile(metadata.Path, []byte(*contents), 0700)
		}
	case operation == GBClientNetworkTools.FILE_DELETED_CONST:
		os.RemoveAll(metadata.Path)
	}
}

// SendPing sends a ping to the server.
func SendPing(tlsConn *tls.Conn) {
	io.WriteString(tlsConn, GBClientNetworkTools.PING)
}
