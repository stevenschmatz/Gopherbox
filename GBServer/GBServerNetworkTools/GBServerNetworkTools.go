package GBServerNetworkTools

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/stevenschmatz/gopherbox/GBServer/GBServerDatabase"
	"github.com/stevenschmatz/gopherbox/GBServer/GBServerWatch"
	"gopkg.in/mgo.v2"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

const (
	FILE_ADDED_CONST = iota
	FILE_MODIFIED_CONST
	FILE_DELETED_CONST
	ERR_CONST
	EOF_CONST
	PING_CONST
	DISCONNECT_CONST

	PING = "." // Sent when pinging the server.
)

var (
	// Files_to_ignore_mac lists that files with the following substrings should be ignored.
	Files_to_ignore_mac []string = []string{".DS_Store", ".sb-", ".function.m.swp", ".localized"}
)

// ReadData reads from the conn and returns the metadata, contents, and intended operation of the data,
// according to the Gopherbox file transfer scheme.
func ReadData(tlsConn *tls.Conn) (metadata GBServerDatabase.FileMetadata, contents string, operation_int int, err error) {
	buf := make([]byte, 512)
	n, err := tlsConn.Read(buf)
	if err != nil {
		if err.Error() == "EOF" {
			return GBServerDatabase.FileMetadata{}, "", EOF_CONST, nil
		}
		return GBServerDatabase.FileMetadata{}, "", ERR_CONST, err
	}

	begin_message := string(buf[0:n])
	if begin_message != getBeginMessage("DATA") {
		if begin_message == PING {
			return GBServerDatabase.FileMetadata{}, "", PING_CONST, nil
		}
		error_text := fmt.Sprintf("The beginning DATA string was invalid. It was %s", begin_message)
		return GBServerDatabase.FileMetadata{}, "", 0, errors.New(error_text)
	}

	n, err = tlsConn.Read(buf)
	if err != nil {
		if err.Error() == "EOF" {
			return GBServerDatabase.FileMetadata{}, "", EOF_CONST, nil
		}
		return GBServerDatabase.FileMetadata{}, "", ERR_CONST, err
	}
	begin_metadata_message := string(buf[0:n])
	if begin_metadata_message != getBeginMessage("METADATA") {
		error_text := fmt.Sprintf("The beginning METADATA string was invalid. It was %s", begin_metadata_message)
		return GBServerDatabase.FileMetadata{}, "", ERR_CONST, errors.New(error_text)
	}

	operation_type, operation, n, err := GetValueFromContentsBlock(tlsConn)
	checkErr(err, tlsConn)
	err = checkDataFormatting(operation_type, "OPERATION", n)
	checkErr(err, tlsConn)

	path_type, path, n, err := GetValueFromContentsBlock(tlsConn)
	checkErr(err, tlsConn)
	err = checkDataFormatting(path_type, "PATH", n)
	checkErr(err, tlsConn)

	size_type, size, n, err := GetValueFromContentsBlock(tlsConn)
	checkErr(err, tlsConn)
	err = checkDataFormatting(size_type, "SIZE", n)
	checkErr(err, tlsConn)

	is_dir_type, is_dir, n, err := GetValueFromContentsBlock(tlsConn)
	checkErr(err, tlsConn)
	err = checkDataFormatting(is_dir_type, "IS_DIR", n)
	checkErr(err, tlsConn)

	modification_time_type, modification_time, n, err := GetValueFromContentsBlock(tlsConn)
	checkErr(err, tlsConn)
	err = checkDataFormatting(modification_time_type, "MODIFICATION_TIME", n)
	checkErr(err, tlsConn)

	n, err = tlsConn.Read(buf)
	if err != nil {
		if err.Error() == "EOF" {
			return GBServerDatabase.FileMetadata{}, "", EOF_CONST, nil
		}
		return GBServerDatabase.FileMetadata{}, "", ERR_CONST, err
	}

	end_metadata_message := string(buf[0:n])
	if end_metadata_message != getEndMessage("METADATA") {
		error_text := fmt.Sprintf("The ending string was invalid. It was %s", end_metadata_message)
		return GBServerDatabase.FileMetadata{}, "", ERR_CONST, errors.New(error_text)
	}

	contents_type, contents, n, err := GetValueFromContentsBlock(tlsConn)
	checkErr(err, tlsConn)
	err = checkDataFormatting(contents_type, "CONTENTS", n)
	checkErr(err, tlsConn)

	n, err = tlsConn.Read(buf)
	if err != nil {
		if err.Error() == "EOF" {
			return GBServerDatabase.FileMetadata{}, "", EOF_CONST, nil
		}
		return GBServerDatabase.FileMetadata{}, "", ERR_CONST, err
	}

	end_message := string(buf[0:n])
	if end_message != getEndMessage("DATA") {
		error_text := fmt.Sprintf("The ending string was invalid. It was %s", end_message)
		return GBServerDatabase.FileMetadata{}, "", ERR_CONST, errors.New(error_text)
	}

	var is_dir_bool bool
	if is_dir == "1" {
		is_dir_bool = true
	} else if is_dir == "0" {
		is_dir_bool = false
	} else {
		return GBServerDatabase.FileMetadata{}, "", ERR_CONST, errors.New("The value for IS_DIR was not valid.")
	}

	modification_time_int64, err := strconv.ParseInt(modification_time, 10, 64)
	checkErr(err, tlsConn)

	size_int64, err := strconv.ParseInt(size, 10, 64)
	checkErr(err, tlsConn)

	operation_int, err = strconv.Atoi(operation)
	checkErr(err, tlsConn)

	metadata = GBServerDatabase.FileMetadata{path, is_dir_bool, modification_time_int64, size_int64}
	return metadata, contents, operation_int, nil
}

// SendData sends the conn the metadata and contents of the file, according to the Gopherbox file transfer scheme.
func SendData(conn *tls.Conn, file GBServerWatch.OutputData, operation_const int) error {
	if CheckToIgnore(file.Path) {
		return nil
	}

	_, err := io.WriteString(conn, getBeginMessage("DATA"))
	if err != nil {
		return err
	}

	io.WriteString(conn, getBeginMessage("METADATA"))

	// Send operation type
	WriteContentsToConn(conn, "OPERATION", strconv.Itoa(operation_const))

	// Send file path
	WriteContentsToConn(conn, "PATH", file.Path)

	// Send file size
	WriteContentsToConn(conn, "SIZE", strconv.FormatInt(file.Size, 10))

	// Send is_dir
	if file.Is_dir {
		WriteContentsToConn(conn, "IS_DIR", "1")
	} else {
		WriteContentsToConn(conn, "IS_DIR", "0")
	}

	// Send modification time
	WriteContentsToConn(conn, "MODIFICATION_TIME", strconv.FormatInt(file.Modification_time, 10))
	io.WriteString(conn, getEndMessage("METADATA"))

	file_contents_byteslice := []byte{}
	if !file.Is_dir && operation_const != FILE_DELETED_CONST {
		file_contents_byteslice, err = ioutil.ReadFile(file.Path)
		if err != nil {
			return err
		}
	} else {
		file_contents_byteslice = []byte{}
	}

	WriteContentsToConn(conn, "CONTENTS", string(file_contents_byteslice))
	io.WriteString(conn, getEndMessage("DATA"))

	return nil
}

// getBeginMessage returns the beginning header of a contents block.
func getBeginMessage(contents_type string) string {
	return "-----BEGIN " + contents_type + "-----"
}

// getContentsType returns the contents type in the header.
func getContentsType(begin_message string) string {
	return begin_message[11 : len(begin_message)-5]
}

// getEndMessage returns the beginning header of a contents block.
func getEndMessage(contents_type string) string {
	return "-----END " + contents_type + "-----"
}

// GetValueFromContentsBlock returns the contents type in the header, as well as the contents contained within.
func GetValueFromContentsBlock(tlsConn *tls.Conn) (contents_type string, contents string, n int, err error) {
	buf := make([]byte, 512)
	n, err = tlsConn.Read(buf)
	if err != nil {
		if err.Error() == "EOF" {
			return "", "", 0, nil
		}
		return "", "", 0, err
	}

	begin_message := string(buf[0:n])
	if !(begin_message[:11] == "-----BEGIN " && begin_message[len(begin_message)-5:] == "-----") {
		error_text := fmt.Sprintf("The beginning string was invalid. It was %s", begin_message)
		return "", "", 0, errors.New(error_text)
	}

	contents_type = getContentsType(begin_message)
	end_message := getEndMessage(contents_type)
	contents = ""

	total_bytes := 0
	for {
		n, err := tlsConn.Read(buf)
		if err != nil && err.Error() != "EOF" {
			return "", "", 0, err
		}
		message := string(buf[0:n])
		if message == end_message {
			break
		}
		total_bytes += n
		contents += message
	}

	return contents_type, contents, total_bytes, nil
}

// checkDataFormatting makes sure that the contents type of a block matches the intended one.
func checkDataFormatting(contents_type, target_contents_type string, n int) error {
	if contents_type != target_contents_type {
		error_string := fmt.Sprintf("%s was not of correct contents type", target_contents_type)
		return errors.New(error_string)
	}
	return nil
}

func InitSessionServer(tlsConn *tls.Conn, UserCollection *mgo.Collection, username string) ([]byte, error) {
	activation_key, err := GBServerDatabase.GetActivationKey(UserCollection, username)
	if err != nil {
		return []byte{}, err
	}
	WriteContentsToConn(tlsConn, "ACTIVATION_KEY", activation_key)

	public_key, err := getPublicKey(tlsConn)
	if err != nil {
		return []byte{}, err
	}

	last_accessed, err := GBServerDatabase.InitClient(UserCollection, username, string(public_key))
	if err != nil {
		return []byte{}, err
	}

	last_accessed_string := strconv.Itoa(int(last_accessed))
	_, err = WriteContentsToConn(tlsConn, "LAST_ACCESSED", last_accessed_string)
	if err != nil {
		return []byte{}, err
	}
	return public_key, nil
}

// checkErr logs if an error occurs and closes the tlsConn.
func checkErr(err error, tlsConn *tls.Conn) {
	if err != nil {
		tlsConn.Close()
		log.Fatalf("GBNetworkTools: %s\n", err.Error())

	}
}

// WriteContentsToConn writes contents to the server in block format, with the content_type header.
func WriteContentsToConn(conn *tls.Conn, content_type string, contents string) (int, error) {
	// Write message to server
	beginMessage := fmt.Sprintf("-----BEGIN %s-----", content_type)
	n, err := io.WriteString(conn, beginMessage)
	if err != nil {
		return n, err
	}
	// log.Printf("client: wrote %q (%d bytes)", beginMessage, n)

	n_contents, err := io.WriteString(conn, contents)
	if err != nil {
		return n_contents, err
	}
	// log.Printf("client: wrote %q (%d bytes)", contents, n_contents)

	endMessage := fmt.Sprintf("-----END %s-----", content_type)
	n, err = io.WriteString(conn, endMessage)
	if err != nil {
		return n, err
	}
	// log.Printf("client: wrote %q (%d bytes)", endMessage, n)

	return n_contents, nil
}

func getPublicKey(tlsConn *tls.Conn) ([]byte, error) {
	state := tlsConn.ConnectionState()
	for _, v := range state.PeerCertificates {
		return x509.MarshalPKIXPublicKey(v.PublicKey)
	}
	return []byte{}, nil
}

// CheckToIgnore checks if the given file should be sent to the Gopherbox filesystem.
// Currently only works with Mac files that should be ignored.
func CheckToIgnore(filePath string) bool {
	for _, files := range Files_to_ignore_mac {
		if strings.Contains(filePath, files) {
			return true
		}
	}
	return false
}
