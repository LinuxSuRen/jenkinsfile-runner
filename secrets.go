package main

import (
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	bytes2 "bytes"
	"io/ioutil"
	"fmt"
	"syscall"
	"github.com/mitchellh/go-homedir"
	"path/filepath"
)


func decrypt(secretfile string) (string, error) {

	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	secretKeyring := filepath.Join(home, ".gnupg", "secring.gpg")


	secret, err := ioutil.ReadFile(secretsfile)
	if err != nil {
		panic(err)
	}

	// Open the secret keyring
	keyringFileBuffer, err := os.Open(secretKeyring)
	if err != nil {
		return "", err
	}
	defer keyringFileBuffer.Close()
	entityList, err := openpgp.ReadKeyRing(keyringFileBuffer)
	if err != nil {
		return "", err
	}
	fmt.Println("Private key from armored string:", entityList[0].Identities)

	// Decrypt it with the contents of the private key
	md, err := openpgp.ReadMessage(bytes2.NewBuffer(secret), entityList, prompt, nil)
	if err != nil {
		return "", err
	}
	bytes, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

var passphrase []byte

func prompt(keys []openpgp.Key, symmetric bool) ([]byte, error) {
	// TODO add support for gpg-agent
	if passphrase != nil {
		return passphrase, nil
	}

	fmt.Println("Enter gpg passphrase:")
	passphrase, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return passphrase, err
	}
	return passphrase, nil
}