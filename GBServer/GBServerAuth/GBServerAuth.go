package GBServerAuth

import (
	"crypto/tls"
	"fmt"
	"github.com/howeyc/gopass"
	"github.com/stevenschmatz/gopherbox/GBServer/GBServerDatabase"
	"github.com/stevenschmatz/gopherbox/GBServer/GBServerNetworkTools"
	"gopkg.in/mgo.v2"
	"os"
	"os/exec"
)

func AuthenticateClient(tlsConn *tls.Conn) (verified bool, err error) {
	var username, password string
	clearScreen()
	for i := 0; i < 3; i++ {
		fmt.Printf("Enter your username and password. (%v remaining)\n", 3-i)
		fmt.Printf("Username: ")
		fmt.Scan(&username)
		fmt.Printf("Password: ")
		password_bytes := gopass.GetPasswdMasked()
		password = string(password_bytes)
		GBServerNetworkTools.WriteContentsToConn(tlsConn, "USERNAME", username)
		GBServerNetworkTools.WriteContentsToConn(tlsConn, "PASSWORD", password)
		_, verification_string, _, err := GBServerNetworkTools.GetValueFromContentsBlock(tlsConn)
		if err != nil {
			return false, err
		}
		if verification_string == "true" {
			clearScreen()
			fmt.Println("You were verified. Session starting.")
			return true, nil
		} else {
			clearScreen()
			fmt.Println("The username or password was not correct. Please try again.")
			continue
		}
	}
	fmt.Println("You were unsuccessful in your authentication attempts. The connection is now closing.")
	return false, nil
}

func AuthenticateServer(tlsConn *tls.Conn, UserCollection *mgo.Collection) (verified bool, username string, err error) {
	var password string

	_, username, username_length, err := GBServerNetworkTools.GetValueFromContentsBlock(tlsConn)
	if username_length == 0 {
		GBServerNetworkTools.WriteContentsToConn(tlsConn, "ERROR", "The username was of length 0.")
	}
	if err != nil {
		return false, "", err
	}

	_, password, password_length, err := GBServerNetworkTools.GetValueFromContentsBlock(tlsConn)
	if password_length == 0 {
		GBServerNetworkTools.WriteContentsToConn(tlsConn, "ERROR", "The password was of length 0.")
	}
	if err != nil {
		return false, "", err
	}

	validated, err := GBServerDatabase.ValidatePassword(UserCollection, username, password)
	if validated {
		GBServerNetworkTools.WriteContentsToConn(tlsConn, "VERIFICATION", "true")
	} else {
		GBServerNetworkTools.WriteContentsToConn(tlsConn, "VERIFICATION", "false")
	}
	if err != nil {
		return false, "", err
	}
	return validated, username, nil
}

func TryAuthenticationThreeTimes(tlsConn *tls.Conn, UserCollection *mgo.Collection) (verified bool, username string, err error) {
	for i := 0; i < 3; i++ {
		verified, username, _ := AuthenticateServer(tlsConn, UserCollection)
		if verified {
			return true, username, nil
		}
	}
	return false, "", nil
}

func clearScreen() {
	clearScreen := exec.Command("clear")
	clearScreen.Stdout = os.Stdout
	clearScreen.Run()
}
