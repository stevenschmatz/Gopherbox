package GBServerRuntime

import (
	"crypto/tls"
	"github.com/stevenschmatz/gopherbox/GBServer/GBServerDatabase"
	"github.com/stevenschmatz/gopherbox/GBServer/GBServerNetworkTools"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"log"
	"net"
	"os"
)

func InitTLSconn(conn net.Conn) *tls.Conn {
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return nil
	}
	err := tlsConn.Handshake()
	if err != nil {
		log.Fatalf("server: handshake failed: %s", err)
	}

	return tlsConn
}

func HandleDataChangesServer(UserCollection *mgo.Collection, username, Source_RSA_Public_Key string, metadata GBServerDatabase.FileMetadata, contents *string, operation int) {

	pathWithUsername := username + "/" + metadata.Path

	switch {

	// File or directory created
	case operation == GBServerNetworkTools.FILE_ADDED_CONST:

		if metadata.Is_dir {
			err := GBServerDatabase.AddPathToAllOtherClients(UserCollection, username, Source_RSA_Public_Key, metadata, operation)
			checkErr(err)

			log.Printf("User %s\tadded directory\t %s\n", username, pathWithUsername)

			err = os.MkdirAll(pathWithUsername, 0700)
			checkErr(err)

		} else {
			err := GBServerDatabase.AddPathToAllOtherClients(UserCollection, username, Source_RSA_Public_Key, metadata, operation)
			checkErr(err)

			log.Printf("User %s\tadded file\t %s\n", username, pathWithUsername)

			err = ioutil.WriteFile(pathWithUsername, []byte(*contents), 0700)
			checkErr(err)
		}

	// File modified (dir modification does not need to be communicated)
	case operation == GBServerNetworkTools.FILE_MODIFIED_CONST:
		if !metadata.Is_dir {
			err := GBServerDatabase.AddPathToAllOtherClients(UserCollection, username, Source_RSA_Public_Key, metadata, operation)
			checkErr(err)

			log.Printf("User %s\tmodified file\t %s\n", username, pathWithUsername)
			err = os.RemoveAll(pathWithUsername)
			checkErr(err)
			err = ioutil.WriteFile(pathWithUsername, []byte(*contents), 0700)
			checkErr(err)
		}

	// File or directory deleted
	case operation == GBServerNetworkTools.FILE_DELETED_CONST:
		if metadata.Is_dir {
			err := GBServerDatabase.AddPathToAllOtherClients(UserCollection, username, Source_RSA_Public_Key, metadata, operation)
			checkErr(err)

			log.Printf("User %s\tdeleted directory\t %s\n", username, pathWithUsername)

			err = os.RemoveAll(pathWithUsername)
			checkErr(err)

		} else {
			err := GBServerDatabase.AddPathToAllOtherClients(UserCollection, username, Source_RSA_Public_Key, metadata, operation)
			checkErr(err)

			log.Printf("User %s\tdeleted file\t %s\n", username, pathWithUsername)

			err = os.RemoveAll(pathWithUsername)
			checkErr(err)
		}
	}
}

func HandleUserDisconnect(UserCollection *mgo.Collection, username, RSA_Public_Key string, conn *tls.Conn) {
	log.Printf("Disconnect:\tUser '%s' at %s\n", username, conn.RemoteAddr())

	err := GBServerDatabase.UpdateLastAccessedTime(UserCollection, username, string(RSA_Public_Key), 0)
	checkErr(err)
	err = GBServerDatabase.UpdateCurrentlyBeingUsed(UserCollection, username, string(RSA_Public_Key), false)
	checkErr(err)
}

func HandleUserTimeout(UserCollection *mgo.Collection, username, RSA_Public_Key string, conn *tls.Conn, timeout_sec int) {
	log.Printf("Timeout:\tUser '%s' at %s\n", username, conn.RemoteAddr())
	err := GBServerDatabase.UpdateLastAccessedTime(UserCollection, username, string(RSA_Public_Key), 5)
	checkErr(err)
	err = GBServerDatabase.UpdateCurrentlyBeingUsed(UserCollection, username, string(RSA_Public_Key), false)
	checkErr(err)
}

func StringToBool(str string) bool {
	if str == "1" {
		return true
	} else {
		return false
	}
}

func checkErr(err error) {
	if err != nil {
		log.Printf("\n======================\nERROR:\tGBServer: GBServerRuntime: %s\n======================\n", err.Error())
	}
}
