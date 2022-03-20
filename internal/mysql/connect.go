package mysql

import (
	"database/sql"
	"fmt"
	"github.com/99designs/keyring"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/internal/helper"
	"github.com/mjmar01/harmolytics/internal/log"
	"golang.org/x/crypto/ssh/terminal"
)

var db *sql.DB
var prfl string

const (
	KeyName = "database-password"
)

// ConnectDatabase initiates a database connection and manages the OS keyring to store the password
func ConnectDatabase(user, host, port, profile string) (err error) {
	log.Task(fmt.Sprintf("Connecting to database %s:%s", host, port), log.DebugLevel)
	prfl = profile
	if len(user) == 0 {
		return errors.Errorf("No user specified")
	}

	log.Task("Getting database-password", log.TraceLevel)
	kr, err := keyring.Open(keyring.Config{ServiceName: "harmolytics"})
	if err != nil {
		return errors.Wrap(err, 0)
	}
	keys, err := kr.Keys()
	if err != nil {
		return errors.Wrap(err, 0)
	}
	var pwd string
	if helper.StringInSlice(KeyName, keys) {
		log.Trace("Keyring contains database-password")
		item, err := kr.Get(KeyName)
		if err != nil {
			return errors.Wrap(err, 0)
		}
		pwd = string(item.Data)
	} else {
		log.Trace("Keyring does not contain database-password")
		fmt.Printf("\nEnter database password for user %s: ", user)
		pwdData, err := terminal.ReadPassword(0)
		fmt.Print("\n")
		if err != nil {
			return errors.Wrap(err, 0)
		}
		pwd = string(pwdData)
		log.Trace("Saving password to keyring")
		err = kr.Set(keyring.Item{Key: KeyName, Data: pwdData})
		if err != nil {
			return errors.Wrap(err, 0)
		}
	}
	log.Done()
	log.Trace("Testing connection")
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/?timeout=5s", user, pwd, host, port)
	db, err = sql.Open("mysql", connectionString)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	rows, err := db.Query("SELECT VERSION()")
	if err != nil {
		return errors.Wrap(err, 0)
	}
	rows.Close()
	log.Done()
	return
}

func RemovePassword() (err error) {
	kr, err := keyring.Open(keyring.Config{ServiceName: "harmolytics"})
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = kr.Remove("database-password")
	if err != nil {
		return errors.Wrap(err, 0)
	}
	return
}
