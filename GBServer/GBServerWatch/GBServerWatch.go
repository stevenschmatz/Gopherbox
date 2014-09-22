package GBServerWatch

import (
	"github.com/stevenschmatz/gopherbox/GBServer/GBServerDatabase"
	"gopkg.in/mgo.v2"
)

type (
	//OutputData is used as the communication of data in channels
	OutputData struct {
		Path              string `bson: "path"`
		Is_dir            bool   `bson: "is_dir"`
		Modification_time int64  `bson: "modification_time"`
		Size              int64  `bson: "size"`
	}
)

// CheckForFilesToSendFromServer checks for the files to send from a client, and returns the first one of the commands.
func CheckForFileToSendFromServer(UserCollection *mgo.Collection, username string, RSA_Public_Key string) (GBServerDatabase.FileQueue, error) {
	client, err := GBServerDatabase.GetClient(UserCollection, username, RSA_Public_Key)
	if err != nil {
		return GBServerDatabase.FileQueue{}, err
	}

	if len(client.Files_to_receive) == 0 {
		return GBServerDatabase.FileQueue{}, nil
	}

	return client.Files_to_receive[0], nil
}
