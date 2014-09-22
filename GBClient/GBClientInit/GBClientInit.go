package GBClientInit

// package GBClientInit contains the code used by the client to initialize the session with the server.

import (
	"crypto/tls"
	"fmt"
	"github.com/stevenschmatz/gopherbox/GBClient/GBClientNetworkTools"
	"github.com/stevenschmatz/gopherbox/GBClient/GBClientWatch"
	"log"
	"os"
	"os/exec"
	"time"
)

const (
	LOCALHOST_IP = "127.0.0.1:8000"
	SERVER_IP    = "54.201.152.68:8000"

	CONFIG_FOLDER_NAME          = "Gopherbox.app/Contents/Resources/config"
	CLIENT_PRIVATE_KEY_FILENAME = "client.pem"
	CLIENT_PUBLIC_KEY_FILENAME  = "client.key"
)

// HandleOSArguments checks that the number of command-line arguments are correct.
func CheckOSArguments() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <directory-path> <username> <password>\n", os.Args[0])
		os.Exit(1)
	}
}

// InitConnection initializes a TLS connection with the server.
func InitConnection() *tls.Conn {
	cert, err := tls.LoadX509KeyPair(CONFIG_FOLDER_NAME+"/"+CLIENT_PRIVATE_KEY_FILENAME, CONFIG_FOLDER_NAME+"/"+CLIENT_PUBLIC_KEY_FILENAME)
	if err != nil {
		log.Fatalf("client: loadkeys: %s", err)
	}

	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	tlsConn, err := tls.Dial("tcp", LOCALHOST_IP, &config)
	if err != nil {
		log.Fatalf("client: dial: %s", err)
	}
	return tlsConn
}

// CheckVerification checks with the server that the user is verified, and exits the binary if it is not.
func CheckVerification(tlsConn *tls.Conn, loginInfo, password string) {
	GBClientNetworkTools.WriteContentsToConn(tlsConn, "IS_ALREADY_VERIFIED", "1")
	GBClientNetworkTools.WriteContentsToConn(tlsConn, "USERNAME", loginInfo)
	GBClientNetworkTools.WriteContentsToConn(tlsConn, "PASSWORD", password)

	_, verified, _, err := GBClientNetworkTools.GetValueFromContentsBlock(tlsConn)
	checkErr(err, tlsConn)
	if verified != "true" {
		time.Sleep(1000 * time.Millisecond)
		tlsConn.Close()
		os.Exit(1)
	}
}

// ReceiveSessionInfoAndChangeDirectory gets the key and last_accessed information from the server, and changes directory to the Gopherbox folder.
func ReceiveSessionInfoAndChangeDirectory(tlsConn *tls.Conn, key *[]byte, last_accessed *int64, directoryPath, loginInfo string) {
	var err error
	*key, *last_accessed, err = GBClientNetworkTools.ClientReceiveSessionInformation(tlsConn, CONFIG_FOLDER_NAME+"/"+loginInfo)
	checkErr(err, tlsConn)
	err = os.Chdir(directoryPath)
	checkErr(err, tlsConn)
}

// InitChannelsAndFilemap creates the channels and fills the fileMap with a map of paths and fileInfo in the Gopherbox folder.
func InitChannelsAndFilemap(fileAddedChan, fileModifiedChan, fileDeletedChan *chan GBClientWatch.OutputData, fileMap *map[string]GBClientWatch.FileData) {
	channel_capacity := 1000 // This represents the amount of items that can be added, modified, or deleted within the time interval
	*fileAddedChan, *fileModifiedChan, *fileDeletedChan = make(chan GBClientWatch.OutputData, channel_capacity), make(chan GBClientWatch.OutputData, channel_capacity), make(chan GBClientWatch.OutputData, channel_capacity)
	*fileMap = make(map[string]GBClientWatch.FileData)
	GBClientWatch.InitFileMap(fileMap)
}

// SendUnsyncedFiles sends unsynchronized files to the server (ie, files modified after last_accessed)
func SendUnsyncedFiles(tlsConn *tls.Conn, last_accessed int64, private_key []byte) {
	files, err := GBClientWatch.CheckForFilesToSendToServerOnInit(last_accessed)
	checkErr(err, tlsConn)

	for _, fileData := range files {
		log.Println(fileData)
		GBClientNetworkTools.SendDataEncrypted(tlsConn, fileData, GBClientNetworkTools.FILE_ADDED_CONST, private_key)
	}
}

// GenerateKeypair runs the OpenSSL command to generate a private and public key for the user.
func GenerateKeypair(loginInfo string) error {

	os.Mkdir(CONFIG_FOLDER_NAME, 0700)

	configString := "/C=US/O=Gopherbox/emailAddress=" + loginInfo
	cmd := exec.Command(
		"openssl",
		"req",
		"-new",
		"-nodes",
		"-x509",
		"-out", CONFIG_FOLDER_NAME+"/"+CLIENT_PRIVATE_KEY_FILENAME,
		"-keyout", CONFIG_FOLDER_NAME+"/"+CLIENT_PUBLIC_KEY_FILENAME,
		"-subj", configString)

	err := cmd.Start()
	if err != nil {
		return err
	}

	return nil
}

// checkErr logs if an error occurs and exits.
func checkErr(err error, tlsConn *tls.Conn) {
	if err != nil {
		tlsConn.Close()
		log.Fatalf("GBClient: GBClientInit: %s\n", err.Error())
	}
}
