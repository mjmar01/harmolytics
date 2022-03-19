package mysql

import (
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/go-errors/errors"
	"golang.org/x/crypto/ssh/terminal"
	"os"
)

var db *sql.DB
var profile string

// ConnectDatabase saves the database configuration and tests the connection.
// It returns the encrypted password for later use.
// The 'password' parameter is the with 'key' encrypted password.
// If no password is specified the user will be prompted to enter one
func ConnectDatabase(user, password, host, port, prof string, key []byte) (p string, err error) {
	profile = prof
	if len(user) == 0 {
		fmt.Println("No database user specified")
		os.Exit(1)
	}
	var pwd string
	if len(password) > 0 {
		pwd, err = decrypt(password, key)
		if err != nil {
			return
		}
		p = password
	} else {
		fmt.Printf("\nEnter database password for user %s: ", user)
		pwdData, err := terminal.ReadPassword(0)
		fmt.Print("\n")
		if err != nil {
			return "", errors.Wrap(err, 0)
		}
		pwd = string(pwdData)
		p, err = encrypt(pwdData, key)
		if err != nil {
			return "", err
		}
	}
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/?timeout=5s", user, pwd, host, port)
	db, err = sql.Open("mysql", connectionString)
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	rows, err := db.Query("SELECT VERSION()")
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	rows.Close()
	return
}

func encrypt(clearText []byte, key []byte) (c string, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	nonce := make([]byte, aesGCM.NonceSize())
	ciphertext := aesGCM.Seal(nonce, nonce, clearText, nil)
	c = fmt.Sprintf("%x", ciphertext)
	return
}

func decrypt(encryptedString string, key []byte) (c string, err error) {
	enc, err := hex.DecodeString(encryptedString)
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	c = string(plaintext)
	return
}
