package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/stevenschmatz/gopherbox/GBServer/GBServerDatabase"
	"gopkg.in/mgo.v2"
	"log"
)

var (
	username = "schmatz"
	password = "test"
)

func main() {
	session, err := mgo.Dial("localhost")
	checkErr(err)

	UserCollection := session.DB("test").C("userTest")
	MetadataCollection := session.DB("test").C("metadataTest")

	password_hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	checkErr(err)

	err = GBServerDatabase.SignupUser(UserCollection, MetadataCollection, username, "steven@schmatz.com", string(password_hash))
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		log.Fatalf("GBServer: SignupUser: %s", err.Error())
	}
}
