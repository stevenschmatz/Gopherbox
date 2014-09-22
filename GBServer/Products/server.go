package main

import (
	"crypto/tls"
	"github.com/stevenschmatz/gopherbox/GBServer/GBServerAuth"
	"github.com/stevenschmatz/gopherbox/GBServer/GBServerDatabase"
	"github.com/stevenschmatz/gopherbox/GBServer/GBServerInit"
	"github.com/stevenschmatz/gopherbox/GBServer/GBServerNetworkTools"
	"github.com/stevenschmatz/gopherbox/GBServer/GBServerRuntime"
	"github.com/stevenschmatz/gopherbox/GBServer/GBServerWatch"
	"gopkg.in/mgo.v2"
	"io"
	"log"
	"net"
	"os"
	"reflect"
	"time"
)

const (
	RUNTIME_DIRECTORY = "/Users/stevenschmatz/Desktop/Gopherbox-Serverside"
	TIMEOUT_SEC       = 5
)

var (
	listener       net.Listener
	UserCollection *mgo.Collection
)

func init() {
	listener = GBServerInit.InitListener()
	err := os.Chdir(RUNTIME_DIRECTORY)
	if err != nil {
		log.Fatalf("GBServer: Server: %s", err.Error())
	}

	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatalf("GBServer: Server: %s", err.Error())
	}

	UserCollection = session.DB("test").C("userTest")
}

func main() {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("GBServer: Accept error: %s", err)
			continue
		}

		go handleClient(conn, UserCollection)
	}
}

func handleClient(conn net.Conn, UserCollection *mgo.Collection) {
	defer func() {
		conn.Close()
	}()
	tlsConn := GBServerRuntime.InitTLSconn(conn)

	// Is it authenticating or is it good
	_, isAlreadyVerified, _, err := GBServerNetworkTools.GetValueFromContentsBlock(tlsConn)
	CheckErr(err, tlsConn)

	var username string
	verified := false

	if !GBServerRuntime.StringToBool(isAlreadyVerified) {
		GBServerAuth.AuthenticateServer(tlsConn, UserCollection)
		tlsConn.Close()
		return
	} else {
		verified, username, _ = GBServerAuth.AuthenticateServer(tlsConn, UserCollection)
		if !verified {
			tlsConn.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Connect:\tUser '%s' at %s\n", username, conn.RemoteAddr())

	RSA_Public_Key, err := GBServerNetworkTools.InitSessionServer(tlsConn, UserCollection, username)
	CheckErr(err, tlsConn)

	p_time := time.Now()

	for {

		result, err := GBServerWatch.CheckForFileToSendFromServer(UserCollection, username, string(RSA_Public_Key))
		CheckErr(err, tlsConn)

		switch {

		// There exists files to send to clients
		case (!reflect.DeepEqual(GBServerDatabase.FileQueue{}, result)) && (err == nil):

			// Send the file
			GBServerNetworkTools.SendData(
				tlsConn,

				GBServerWatch.OutputData{
					Path:              result.Path[len(username+"/"):],
					Is_dir:            result.Is_dir,
					Modification_time: result.Modification_time,
					Size:              result.Size,
				},

				result.Operation_int,
			)

			GBServerDatabase.RemovePathFromClient(UserCollection, username, string(RSA_Public_Key), username+"/"+result.Path)

		default:
			metadata, contents, operation, _ := GBServerNetworkTools.ReadData(tlsConn)
			time.Sleep(100 * time.Millisecond)

			// Check for disconnect
			if operation == GBServerNetworkTools.DISCONNECT_CONST {
				GBServerRuntime.HandleUserDisconnect(UserCollection, username, string(RSA_Public_Key), tlsConn)
				return
			}

			// Check for timeout
			if time.Since(p_time).Seconds() > TIMEOUT_SEC {
				GBServerRuntime.HandleUserTimeout(UserCollection, username, string(RSA_Public_Key), tlsConn, TIMEOUT_SEC)
				return
			}

			// Normal operation
			if operation != GBServerNetworkTools.PING_CONST && operation != GBServerNetworkTools.ERR_CONST {
				GBServerRuntime.HandleDataChangesServer(UserCollection, username, string(RSA_Public_Key), metadata, &contents, operation)
			} else {
				p_time = time.Now()
				io.WriteString(conn, GBServerNetworkTools.PING)
			}
		}
	}

}

func CheckErr(err error, tlsConn *tls.Conn) {
	if err != nil {
		log.Println("GBServer: %s", err.Error())
		tlsConn.Close()
	}
}
