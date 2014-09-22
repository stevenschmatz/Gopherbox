package GBServerAESEncryption

import (
	"fmt"
	"os"
)

var (
	ErrRandomFailure = fmt.Errorf("failed to read enough random data")
)

func GenerateRandomByteSlice(size int) (b []byte, err error) {
	// Courtesy of Kyle Isom, @gokyle
	devrand, err := os.Open("/dev/random")
	if err != nil {
		return
	}
	defer devrand.Close()

	b = make([]byte, size)
	n, err := devrand.Read(b)
	if err != nil {
		return
	} else if size != n {
		err = ErrRandomFailure
	}
	return
}
