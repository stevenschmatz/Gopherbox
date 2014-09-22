package GBClientNetworkTools

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/stevenschmatz/gopherbox/GBClient/GBClientAESEncryption"
	"github.com/stevenschmatz/gopherbox/GBClient/GBClientWatch"
	"io"
	"io/ioutil"
	"log"
	"os"
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

type (
	// FileMetadata stores the metadata for each file.
	FileMetadata struct {
		Path              string `bson: "path"`
		Is_dir            bool   `bson: "is_dir"`
		Modification_time int64  `bson: "modification_time"`
		Size              int64  `bson: "size"`
	}
)

var (
	// Files_to_ignore_mac lists that files with the following substrings should be ignored.
	Files_to_ignore_mac []string = []string{".DS_Store", ".sb-", ".function.m.swp", ".localized"}
)

// SendMetadata sends the file metadata to the connection, with a truncated or non-truncated path.
func SendMetadata(conn *tls.Conn, file GBClientWatch.OutputData, operation_const int, truncated bool) error {

	io.WriteString(conn, getBeginMessage("METADATA"))

	// Send operation type
	WriteContentsToConn(conn, "OPERATION", strconv.Itoa(operation_const))

	// Send file path
	if truncated && operation_const != DISCONNECT_CONST {
		wd, err := os.Getwd()
		checkErr(err, conn)
		truncated_path := file.Path[len(wd)+1:]
		WriteContentsToConn(conn, "PATH", truncated_path)
	} else {
		WriteContentsToConn(conn, "PATH", file.Path)
	}

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

	return nil
}

// SendData sends the conn the metadata and encrypted contents of the file, according to the Gopherbox file transfer scheme.
func SendDataEncrypted(tlsConn *tls.Conn, file GBClientWatch.OutputData, operation_const int, private_key []byte) error {
	if CheckToIgnore(file.Path) {
		return nil
	}

	_, err := io.WriteString(tlsConn, getBeginMessage("DATA"))
	if err != nil {
		return err
	}

	err = SendMetadata(tlsConn, file, operation_const, true)
	if err != nil {
		return err
	}

	file_contents_byteslice := []byte{}
	if !file.Is_dir && operation_const != FILE_DELETED_CONST {
		file_contents_byteslice, err = ioutil.ReadFile(file.Path)
		if err != nil {
			return err
		}
	} else {
		file_contents_byteslice = []byte{}
	}

	encrypted_contents, err := GBClientAESEncryption.Encrypt(private_key, file_contents_byteslice)
	if err != nil {
		return err
	}

	WriteContentsToConn(tlsConn, "CONTENTS", string(encrypted_contents))
	io.WriteString(tlsConn, getEndMessage("DATA"))

	return nil
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

// getBeginMessage returns the beginning header of a contents block.
func getBeginMessage(contents_type string) string {
	return "-----BEGIN " + contents_type + "-----"
}

// getEndMessage returns the beginning header of a contents block.
func getEndMessage(contents_type string) string {
	return "-----END " + contents_type + "-----"
}

// checkErr logs if an error occurs and closes the tlsConn.
func checkErr(err error, tlsConn *tls.Conn) {
	if err != nil {
		log.Printf("GBNetworkTools: %s\n", err.Error())
		tlsConn.Close()
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

// getContentsType returns the contents type in the header.
func getContentsType(begin_message string) string {
	return begin_message[11 : len(begin_message)-5]
}

func ClientReceiveSessionInformation(tlsConn *tls.Conn, keyfile_path string) (key []byte, last_accessed int64, err error) {
	_, activation_key, _, err := GetValueFromContentsBlock(tlsConn)
	if err != nil {
		return []byte{}, 0, err
	}

	_, last_accessed_string, _, err := GetValueFromContentsBlock(tlsConn)
	if err != nil {
		return []byte{}, 0, err
	}

	last_accessed_int, err := strconv.Atoi(last_accessed_string)
	if err != nil {
		return []byte{}, 0, err
	}
	last_accessed = int64(last_accessed_int)

	key, err = GBClientAESEncryption.GetSymmetricKey(keyfile_path, 32, activation_key)

	return key, last_accessed, err
}

func ReadDataEncrypted(tlsConn *tls.Conn, private_key []byte) (metadata FileMetadata, contents string, operation_int int, err error) {
	buf := make([]byte, 512)
	n, err := tlsConn.Read(buf)
	if err != nil {
		if err.Error() == "EOF" {
			return FileMetadata{}, "", EOF_CONST, nil
		}
		return FileMetadata{}, "", ERR_CONST, err
	}

	begin_message := string(buf[0:n])
	if begin_message != getBeginMessage("DATA") {
		if begin_message == PING {
			return FileMetadata{}, "", PING_CONST, nil
		}
		error_text := fmt.Sprintf("The beginning DATA string was invalid. It was %s", begin_message)
		return FileMetadata{}, "", 0, errors.New(error_text)
	}

	n, err = tlsConn.Read(buf)
	if err != nil {
		if err.Error() == "EOF" {
			return FileMetadata{}, "", EOF_CONST, nil
		}
		return FileMetadata{}, "", ERR_CONST, err
	}
	begin_metadata_message := string(buf[0:n])
	if begin_metadata_message != getBeginMessage("METADATA") {
		error_text := fmt.Sprintf("The beginning METADATA string was invalid. It was %s", begin_metadata_message)
		return FileMetadata{}, "", ERR_CONST, errors.New(error_text)
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
			return FileMetadata{}, "", EOF_CONST, nil
		}
		return FileMetadata{}, "", ERR_CONST, err
	}

	end_metadata_message := string(buf[0:n])
	if end_metadata_message != getEndMessage("METADATA") {
		error_text := fmt.Sprintf("The ending string was invalid. It was %s", end_metadata_message)
		return FileMetadata{}, "", ERR_CONST, errors.New(error_text)
	}

	encrypted_contents_type, encrypted_contents, n, err := GetValueFromContentsBlock(tlsConn)
	checkErr(err, tlsConn)
	err = checkDataFormatting(encrypted_contents_type, "CONTENTS", n)
	checkErr(err, tlsConn)

	n, err = tlsConn.Read(buf)
	if err != nil {
		if err.Error() == "EOF" {
			return FileMetadata{}, "", EOF_CONST, nil
		}
		return FileMetadata{}, "", ERR_CONST, err
	}

	end_message := string(buf[0:n])
	if end_message != getEndMessage("DATA") {
		error_text := fmt.Sprintf("The ending string was invalid. It was %s", end_message)
		return FileMetadata{}, "", ERR_CONST, errors.New(error_text)
	}

	var is_dir_bool bool
	if is_dir == "1" {
		is_dir_bool = true
	} else if is_dir == "0" {
		is_dir_bool = false
	} else {
		return FileMetadata{}, "", ERR_CONST, errors.New("The value for IS_DIR was not valid.")
	}

	modification_time_int64, err := strconv.ParseInt(modification_time, 10, 64)
	checkErr(err, tlsConn)

	size_int64, err := strconv.ParseInt(size, 10, 64)
	checkErr(err, tlsConn)

	operation_int, err = strconv.Atoi(operation)
	checkErr(err, tlsConn)

	var contents_byteslice []byte

	if (operation_int == FILE_DELETED_CONST) || (operation_int == FILE_ADDED_CONST && is_dir_bool) {
		contents_byteslice = []byte{}
	} else {
		contents_byteslice, err = GBClientAESEncryption.Decrypt(private_key, []byte(encrypted_contents))
		checkErr(err, tlsConn)
	}

	metadata = FileMetadata{path, is_dir_bool, modification_time_int64, size_int64}
	return metadata, string(contents_byteslice), operation_int, nil
}

// checkDataFormatting makes sure that the contents type of a block matches the intended one.
func checkDataFormatting(contents_type, target_contents_type string, n int) error {
	if contents_type != target_contents_type {
		error_string := fmt.Sprintf("%s was not of correct contents type", target_contents_type)
		return errors.New(error_string)
	}
	return nil
}
