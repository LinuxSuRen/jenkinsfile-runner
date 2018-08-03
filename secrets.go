package main

import (
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/openpgp"
	"os"
	"io/ioutil"
	"fmt"
	"path/filepath"
	"bytes"
	"github.com/rickar/props"
	"golang.org/x/crypto/ssh/terminal"
	"syscall"
	"errors"
)


func decrypt(secretfile string) (string, error) {

	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	secretKeyring, err := os.Open(filepath.Join(home, ".gnupg", "secring.gpg"))
	if err != nil {
		panic(err)
	}

	privring, err := openpgp.ReadKeyRing(secretKeyring)
	if err != nil {
		return "", err
	}

	fmt.Println("Enter GPG passphrase")
	passphrase, err := terminal.ReadPassword(int(syscall.Stdin))

	for _, e := range privring {
		e.PrivateKey.Decrypt(passphrase)
		for _, subkey := range e.Subkeys {
			subkey.PrivateKey.Decrypt(passphrase)
		}
	}

	file, err := os.Open(secretsfile)
	if err != nil {
		return "", err
	}

	md, err := openpgp.ReadMessage(file, privring, nil, nil)
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

var passphrase []byte

func promptFunction(keys []openpgp.Key, symmetric bool) ([]byte, error) {

	if passphrase != nil {
		return passphrase, errors.New("Incorrect GPG passphrase")
	}

	fmt.Println("Enter GPG passphrase")
	passphrase, err := terminal.ReadPassword(int(syscall.Stdin))
	return passphrase, err
	
	/*
	conn, err := gpgagent.NewGpgAgentConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	for _, key := range keys {
		cacheId := strings.ToUpper(hex.EncodeToString(key.PublicKey.Fingerprint[:]))
		// TODO: Add prompt, etc.
		request := gpgagent.PassphraseRequest{CacheKey: cacheId}
		passphrase, err := conn.GetPassphrase(&request)
		if err != nil {
			return nil, err
		}
		err = key.PrivateKey.Decrypt([]byte(passphrase))
		if err != nil {
			return nil, err
		}
		return []byte(passphrase), nil
	}
	return nil, fmt.Errorf("Unable to find key")
	*/
}


func propertiesToDockerSecretLayout(properties string, folder string) error {
	p, err := props.Read(bytes.NewBufferString(properties))
	if err != nil {
		return err
	}
	for _,name := range p.Names() {
		writeFile(filepath.Join(folder, name), p.Get(name))
	}
	return nil
}