package storage

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"os"
	"strings"
	"syscall"
)

var (
	isNewKey bool
	key      Key
)

func configure() {

	log.I.Ln("initializing storage")

	configureKey()
	configureDirPath()
	configureFile()
}

func configureKey() {

	log.I.Ln("looking for encryption key")

	var err error

	if viper.GetString(storeKeyFlag) != "" {

		log.I.Ln("found key")

		key.Decode(viper.GetString(storeKeyFlag))

		return
	}

	log.I.Ln("no key found, checking for keyfile")

	if viper.GetString(storeKeyFileFlag) != "" {

		var fileInfo os.FileInfo

		if fileInfo, err = os.Stat(viper.GetString(storeKeyFileFlag)); err != nil {
			startupErrors <- err
			return
		}

		if fileInfo.Mode() != 0600 {
			log.W.Ln("keyfile permissions are too open:", fileInfo.Mode())
			log.W.Ln("It is recommended that you change them to 0600")
		}

		var keyBytes []byte

		if keyBytes, err = os.ReadFile(viper.GetString(storeKeyFileFlag)); err != nil {
			startupErrors <- err
			return
		}

		key.Decode(string(keyBytes))

		return
	}

	if viper.GetBool(storeAskPassFlag) {

		log.I.Ln("prompting user for key")

		var password []byte

		fmt.Print("Enter Encryption Key: ")
		password, err = term.ReadPassword(int(syscall.Stdin))

		key.Decode(string(password))

		return
	}

	log.I.Ln("no keyfile found, generating a new key")

	isNewKey = true

	if key, err = KeyGen(); err != nil {
		startupErrors <- err
		return
	}

	log.W.Ln("")
	log.W.Ln("--------------------------------------------------------")
	log.W.Ln("--")
	log.W.Ln("-- WARNING: The following key will be used to store")
	log.W.Ln("-- your database securely, please ensure that you make")
	log.W.Ln("-- a copy and store it in a secure place before using")
	log.W.Ln("-- this software in a production environment.")
	log.W.Ln("--")
	log.W.Ln("--")
	log.W.Ln("-- Failure to store this key properly will result in")
	log.W.Ln("-- no longer being able to decrypt this database.")
	log.W.Ln("--")
	log.W.Ln("--")
	log.W.Ln("-- It is recommended to use the following to generate")
	log.W.Ln("-- your key:")
	log.W.Ln("--")
	log.W.Ln("-- indra seed keygen")
	log.W.Ln("--")
	log.W.Ln("--  OR")
	log.W.Ln("--")
	log.W.Ln("-- indra seed keygen --keyfile=/path/to/keyfile")
	log.W.Ln("--")
	log.W.Ln("--")
	log.W.Ln("-- YOU HAVE BEEN WARNED!")
	log.W.Ln("--")
	log.W.Ln("-------------------------------------------------------")
	log.W.Ln("-- KEY:", key.Encode(), "--")
	log.W.Ln("-------------------------------------------------------")
	log.W.Ln("")

	viper.Set(storeKeyFlag, key.Encode())
}

func configureDirPath() {

	var err error

	if viper.GetString(storeFilePathFlag) == "" {
		viper.Set(storeFilePathFlag, viper.GetString("data-dir")+"/"+fileName)
	}

	err = os.MkdirAll(
		strings.TrimSuffix(viper.GetString(storeFilePathFlag), "/"+fileName),
		0755,
	)

	if err != nil {
		startupErrors <- err
		return
	}

}

func configureFile() {

	log.I.Ln("using storage db path:")
	log.I.Ln("-", viper.GetString(storeFilePathFlag))

	var err error

	if _, err = os.Stat(viper.GetString(storeFilePathFlag)); err != nil {

		log.I.Ln("none found, creating a new one")

		//file, err := os.OpenFile(viper.GetString(storeFilePathFlag), os.O_WRONLY, 0666)
		//
		//if err != nil && os.IsPermission(err) {
		//	startupErrors <- err
		//	return
		//}
		//
		//file.Close()
		//
		//os.Remove(viper.GetString(storeFilePathFlag))

		return
	}

	if isNewKey {
		startupErrors <- errors.New("new key generated for an existing database. check your configuration.")
		return
	}
}