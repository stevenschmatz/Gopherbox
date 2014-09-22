package GBClientWatch

import (
	"fmt"
	"os"
	"path/filepath"
)

//FileData is used for the values of the hash map for fileMap
type (
	FileData struct {
		ModificationTime int64 `bson: "modificationTime"`
		IsDir            bool  `bson: "isDir"`
		Size             int64 `bson: "size"`
	}

	//OutputData is used as the communication of data in channels
	OutputData struct {
		Path              string `bson: "path"`
		Is_dir            bool   `bson: "is_dir"`
		Modification_time int64  `bson: "modification_time"`
		Size              int64  `bson: "size"`
	}
)

//InitFileMap initiates a fileMap, representing the intial state of the watched directory
func InitFileMap(fMap *map[string]FileData) error {

	fileMap := *fMap

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	err = filepath.Walk("./",
		func(path string, info os.FileInfo, err error) error {
			fullPath := wd + "/" + path
			fileMap[fullPath] = FileData{info.ModTime().Unix(), info.IsDir(), info.Size()}
			return nil
		})
	return err
}

//CheckForChanges detects if a file was added, modified, or deleted, and pushes the info through the channels given as input
func CheckForChanges(previousFileMap map[string]FileData, file_added_chan, file_modified_chan, file_deleted_chan chan OutputData) {

	wd, err := os.Getwd()
	checkErr(err)

	deleteList := make(map[string]bool)
	filesToDelete := make(map[string]bool)

	filepath.Walk("./",
		func(path string, info os.FileInfo, err error) error {

			fullPath := wd + "/" + path

			modTime := info.ModTime().Unix()
			_, filePresent := previousFileMap[fullPath]

			deleteList[fullPath] = true

			switch {

			//Case 1: File was modified
			case filePresent == true && previousFileMap[fullPath].ModificationTime != modTime:
				go func() { file_modified_chan <- OutputData{fullPath, info.IsDir(), info.ModTime().Unix(), info.Size()} }()
				previousFileMap[fullPath] = FileData{modTime, info.IsDir(), info.Size()}

			//Case 2: File was added
			case filePresent == false:
				go func() { file_added_chan <- OutputData{fullPath, info.IsDir(), info.ModTime().Unix(), info.Size()} }()
				previousFileMap[fullPath] = FileData{modTime, info.IsDir(), info.Size()}

			//Case 3: File unchanged
			default:
				return err
			}
			return err
		})

	for k, _ := range previousFileMap {
		filesToDelete[k] = true
	}

	for k, _ := range deleteList {
		delete(filesToDelete, k)
	}

	for k, _ := range filesToDelete {
		func() {
			file_deleted_chan <- OutputData{k, previousFileMap[k].IsDir, previousFileMap[k].ModificationTime, previousFileMap[k].Size}
		}()
		delete(previousFileMap, k)
	}
}

// CheckForFilesToSendToServerOnInit returns a list of files that need to be sent to the server to be updated.
func CheckForFilesToSendToServerOnInit(last_accessed int64) ([]OutputData, error) {
	files_to_send := []OutputData{}

	wd, err := os.Getwd()
	if err != nil {
		return []OutputData{}, err
	}

	err = filepath.Walk("./",
		func(path string, info os.FileInfo, err error) error {
			fullPath := wd + "/" + path
			if info.ModTime().Unix() > last_accessed {
				files_to_send = append(files_to_send, OutputData{fullPath, info.IsDir(), info.ModTime().Unix(), info.Size()})
			}
			return nil
		})
	return files_to_send, err
}

//CheckErr reports an error to os.Stderr, if it exists.
func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
