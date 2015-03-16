package GBClientAESEncryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
)

var (
	ErrRandomFailure = fmt.Errorf("failed to read enough random data")
	ErrPadding       = fmt.Errorf("failed to pad")
	ErrInvalidIV     = fmt.Errorf("invalid IV")
)

func GenerateRandomByteSlice(size int) (b []byte, err error) {
	// Modified to use the system's strong PRNG (rather than /dev/random)
	b = make([]byte, size)

	n, err := rand.Read(b)
	if err != nil {
		return
	} else if size != n {
		err = ErrRandomFailure
	}
	return
}

func PadBuffer(m []byte) (p []byte, err error) {
	mLen := len(m)

	p = make([]byte, mLen)
	copy(p, m)

	if len(p) != mLen {
		return p, ErrPadding
	}

	padding := aes.BlockSize - mLen%aes.BlockSize

	p = append(p, 0x80)
	for i := 1; i < padding; i++ {
		p = append(p, 0x0)
	}
	return
}

func UnpadBuffer(p []byte) (m []byte, err error) {
	m = p
	var pLen int
	origLen := len(m)

	for pLen = origLen - 1; pLen >= 0; pLen-- {
		if m[pLen] == 0x80 {
			break
		}

		if m[pLen] != 0x0 || (origLen-pLen) > aes.BlockSize {
			err = ErrPadding
			return
		}
	}
	m = m[:pLen]
	return
}

func GenerateIV() (iv []byte, err error) {
	return GenerateRandomByteSlice(aes.BlockSize)
}

func Encrypt(key []byte, msg []byte) (ct []byte, err error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	iv, err := GenerateIV()
	if err != nil {
		return
	}
	padded, err := PadBuffer(msg)
	if err != nil {
		return
	}
	cbc := cipher.NewCBCEncrypter(c, iv)
	cbc.CryptBlocks(padded, padded) // encrypt in-place
	ct = iv
	ct = append(ct, padded...)

	return
}

func Decrypt(key []byte, ct []byte) (msg []byte, err error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	// Copy the ciphertext to prevent it from being modified.
	tmp_ct := make([]byte, len(ct))
	copy(tmp_ct, ct)
	iv := tmp_ct[:aes.BlockSize]
	if len(iv) != aes.BlockSize {
		return msg, ErrInvalidIV
	}
	msg = tmp_ct[aes.BlockSize:]

	cbc := cipher.NewCBCDecrypter(c, iv)
	cbc.CryptBlocks(msg, msg)
	msg, err = UnpadBuffer(msg)
	return
}

func GenerateKeyIfNoneExists(filepath string, size int, activation_key string) (err error) {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {

		f, err := os.Create(filepath)
		if err != nil {
			return nil
		}

		key, _ := GenerateRandomByteSlice(size)
		encrypted_key, err := Encrypt([]byte(activation_key), key)
		_, err = f.Write(encrypted_key)
		if err != nil {
			return nil
		}
	}
	return nil
}

func GetSymmetricKey(filepath string, size int, activation_key string) ([]byte, error) {
	err := GenerateKeyIfNoneExists(filepath, size, activation_key)
	if err != nil {
		return nil, err
	}

	encrypted_key, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	key, err := Decrypt([]byte(activation_key), encrypted_key)
	return key, err
}
