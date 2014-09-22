package GBServerDatabase

import (
	"code.google.com/p/go.crypto/bcrypt"
	"errors"
	"github.com/stevenschmatz/gopherbox/GBServer/GBServerAESEncryption"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

const (
	DEFAULT_STORAGE_BYTES = 1073741824 // 1 gigabyte
	// Implement a slider on website: Pay for what you want past this.
)

type (
	// A container used for internal functions.
	ClientContainer struct {
		Clients []Client `bson: "client"`
	}

	// FileQueue is the information the server needs to send a file to client.
	FileQueue struct {
		Path              string `bson: "path"`
		Is_dir            bool   `bson: "is_dir"`
		Modification_time int64  `bson: "modification_time"`
		Size              int64  `bson: "size"`
		Operation_int     int    `bson: "operation_int"`
	}

	// FileMetadata stores the metadata for each file.
	FileMetadata struct {
		Path              string `bson: "path"`
		Is_dir            bool   `bson: "is_dir"`
		Modification_time int64  `bson: "modification_time"`
		Size              int64  `bson: "size"`
	}

	// Client stores the name, IP address, and usage time of a computer.
	Client struct {
		Public_Key           string      `bson: "public_key"`
		Currently_being_used bool        `bson: "currently_being_used"`
		Last_accessed        int64       `bson: "last_accessed"`
		Files_to_receive     []FileQueue `bson: "files_to_receive"`
	}

	// UserFiles stores all the metadata for a user's files.
	UserFiles struct {
		Username           string         `bson: "username"`
		User_file_metadata []FileMetadata `bson: "user_file_metadata"`
	}

	// User (collection item) specifies the information kept for each user.
	User struct {
		// Auth info
		Username      string `bson: "username"`
		Password_hash string `bson: "password_hash"`
		Email         string `bson: "email"`

		// A key that symmetrically encrypts the user's encryption key.
		// Sent to user over TLS on session start, decrypts their file key.
		Activation_key string `bson: "activation_key"`

		// Usage info
		Clients               []Client `bson: "clients"`
		Current_storage_usage int64    `bson: "current_storage_usage"`
		Max_storage_possible  int64    `bson: "max_storage_possible"`
	}
)

// AddPathToReceive removes a file from the queue.
func RemovePathFromClient(UserCollection *mgo.Collection, username string, RSA_Public_Key string, path string) error {
	err := UserCollection.Update(
		bson.M{"username": username, "clients.public_key": RSA_Public_Key},
		bson.M{"$pull": bson.M{"clients.$.files_to_receive": bson.M{"path": path}}},
	)
	return err
}

// UpdateLastAccessedTime updates the last accessed time of the client.
func UpdateLastAccessedTime(UserCollection *mgo.Collection, username string, public_key string, timeout int64) error {
	err := UserCollection.Update(
		bson.M{"username": username, "clients.public_key": public_key},
		bson.M{"$set": bson.M{"clients.$.last_accessed": time.Now().Unix() - timeout}},
	)
	return err
}

// UpdateCurrentlyBeingUsed sets the current state of the currently_being_used variable to the value given.
func UpdateCurrentlyBeingUsed(UserCollection *mgo.Collection, username string, RSA_Public_Key string, currently_being_used bool) error {
	err := UserCollection.Update(
		bson.M{"username": username, "clients.public_key": RSA_Public_Key},
		bson.M{"$set": bson.M{"clients.$.currently_being_used": currently_being_used}},
	)
	return err
}

// AddPathToAllOtherClients queues the files to be sent to all other clients.
func AddPathToAllOtherClients(UserCollection *mgo.Collection, username string, Source_RSA_Public_Key string, metadata FileMetadata, operation_int int) error {
	clients, err := GetAllClients(UserCollection, username)
	if err != nil {
		return err
	}
	for _, client := range clients {
		if client.Public_Key == Source_RSA_Public_Key {
			continue
		} else {

			err := AddPathToReceive(UserCollection, username, client.Public_Key, metadata, operation_int)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetAllClients returns a list of all clients for a user.
func GetAllClients(UserCollection *mgo.Collection, username string) ([]Client, error) {
	result := User{}
	err := UserCollection.Find(bson.M{"username": username}).One(&result)
	if err != nil {
		return []Client{}, err
	}
	return result.Clients, nil
}

// AddPathToReceive queues a file to be sent to the client.
func AddPathToReceive(UserCollection *mgo.Collection, username string, RSA_Public_Key string, metadata FileMetadata, operation_int int) error {
	err := UserCollection.Update(
		bson.M{"username": username, "clients.public_key": RSA_Public_Key},
		bson.M{"$push": bson.M{"clients.$.files_to_receive": bson.M{
			"path":              metadata.Path,
			"is_dir":            metadata.Is_dir,
			"modification_time": metadata.Modification_time,
			"size":              metadata.Size,
			"operation_int":     operation_int,
		}}},
	)
	return err
}

// GetActivationKey returns the activation key for the given account.
func GetActivationKey(UserCollection *mgo.Collection, username string) (activation_key string, err error) {
	if !UsernameExists(UserCollection, username) {
		return "", errors.New("GetActivationKey: The given username does not exist.")
	}

	result := User{}
	err = UserCollection.Find(bson.M{"username": username}).One(&result)
	if err != nil {
		return "", err
	}

	return result.Activation_key, nil
}

// UsernameExists checks if the given username already exists in the collection.
func UsernameExists(UserCollection *mgo.Collection, username string) bool {
	result := User{}
	err := UserCollection.Find(bson.M{"username": username}).One(&result)
	if err == nil {
		return true
	} else {
		return false
	}
}

func InitClient(UserCollection *mgo.Collection, username string, RSA_Public_Key string) (int64, error) {
	exists, err := ClientExists(UserCollection, username, RSA_Public_Key)
	if err != nil {
		return 0, err
	}
	if exists {
		client, err := GetClient(UserCollection, username, RSA_Public_Key)
		if err != nil {
			return client.Last_accessed, err
		}
		last_accessed := client.Last_accessed
		err = UpdateLastAccessedTime(UserCollection, username, RSA_Public_Key, 0)
		if err != nil {
			return 0, err
		}
		err = UpdateCurrentlyBeingUsed(UserCollection, username, RSA_Public_Key, true)
		if err != nil {
			return 0, err
		}
		return last_accessed, nil
	} else {
		current_time := time.Now().Unix()
		client_to_insert := Client{
			RSA_Public_Key,
			true,
			current_time,
			[]FileQueue{},
		}
		err := InsertClient(UserCollection, username, client_to_insert)
		if err != nil {
			return 0, err
		}
		return current_time, nil
	}
}

// ClientExists returns true if the given client has already been used by the user.
func ClientExists(UserCollection *mgo.Collection, username string, RSA_Public_Key string) (bool, error) {
	result := ClientContainer{}

	err := UserCollection.Find(
		bson.M{"username": username},
	).Select(
		bson.M{"clients": bson.M{"$elemMatch": bson.M{"public_key": RSA_Public_Key}}},
	).One(&result)

	if err != nil {
		return false, err
	}

	if len(result.Clients) == 0 {
		return false, nil
	}

	return true, nil
}

// UsernameExists checks if the given email already exists in the collection.
func EmailExists(UserCollection *mgo.Collection, email string) bool {
	result := User{}
	err := UserCollection.Find(bson.M{"email": email}).One(&result)
	if err == nil {
		return true
	} else {
		return false
	}
}

// GetClient returns a specific client for a given RSA public key.
func GetClient(UserCollection *mgo.Collection, username, RSA_Public_Key string) (Client, error) {
	result := ClientContainer{}

	err := UserCollection.Find(
		bson.M{"username": username},
	).Select(
		bson.M{"clients": bson.M{"$elemMatch": bson.M{"public_key": RSA_Public_Key}}},
	).One(&result)
	if err != nil {
		return Client{}, err
	}

	var queried_client = Client{}

	for _, client := range result.Clients {
		if client.Public_Key == RSA_Public_Key {
			queried_client = client
			return queried_client, nil
		}
	}

	return Client{}, nil
}

// InsertClient inserts a client with the given information into the clients list of the user.
func InsertClient(UserCollection *mgo.Collection, username string, client Client) error {
	err := UserCollection.Update(bson.M{"username": username}, bson.M{"$push": bson.M{"clients": &client}})
	return err
}

// ValidatePassword checks if a given cleartext password matches the hash stored in the server.
func ValidatePassword(collection *mgo.Collection, username string, password string) (bool, error) {
	result := User{}
	err := collection.Find(bson.M{"username": username}).One(&result)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(result.Password_hash), []byte(password))
	if err != nil {
		return false, err // Error would be that they don't match
	} else {
		return true, err
	}
}

// SignupUser creates a user with the given username, email, and password hash.
func SignupUser(UserCollection, MetadataCollection *mgo.Collection, username, email, password_hash string) error {
	if UsernameExists(UserCollection, username) {
		return errors.New("ERROR: The given username is already associated with an account.")
	} else if EmailExists(UserCollection, email) {
		return errors.New("ERROR: The given email address is already associated with an account.")
	}

	activation_key, err := GBServerAESEncryption.GenerateRandomByteSlice(32)
	if err != nil {
		return err
	}

	newUser := User{
		username,
		password_hash,
		email,

		string(activation_key),

		[]Client{},
		int64(0),
		DEFAULT_STORAGE_BYTES,
	}

	newUserMetadata := UserFiles{
		username,
		[]FileMetadata{},
	}

	err = UserCollection.Insert(&newUser)
	if err != nil {
		return err
	}

	err = MetadataCollection.Insert(&newUserMetadata)
	if err != nil {
		return err
	}

	return nil
}
