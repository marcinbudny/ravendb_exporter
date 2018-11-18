package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"strings"
)

func prepareTLSConfig() *tls.Config {
	tlsConfig := &tls.Config{}

	if caCertFile != "" {
		caCertData, err := ioutil.ReadFile(caCertFile)
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCertData)
		tlsConfig.RootCAs = caCertPool
	}

	if useAuth {

		cert, err := loadX509KeyPair(clientCertFile, clientKeyFile, clientKeyPassword)
		if err != nil {
			log.Fatal(err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		tlsConfig.BuildNameToCertificate()
	}

	return tlsConfig
}

func loadX509KeyPair(certFile, keyFile, password string) (cert tls.Certificate, err error) {
	clientCertData, err := ioutil.ReadFile(clientCertFile)
	if err != nil {
		return tls.Certificate{}, err
	}

	clientKeyData, err := ioutil.ReadFile(clientKeyFile)
	if err != nil {
		return tls.Certificate{}, err
	}

	return createX509KeyPair(clientCertData, clientKeyData, clientKeyPassword)
}

// code below is partially extracted from Golang standard library Copyright (c) 2009 The Go Authors
// and modified to support encrypted private keys
// original license can be found in GOLANG-BSD-LICENSE file

func createX509KeyPair(certData, keyData []byte, password string) (cert tls.Certificate, err error) {
	var certBlock *pem.Block
	for {
		certBlock, certData = pem.Decode(certData)
		if certBlock == nil {
			break
		}
		if certBlock.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, certBlock.Bytes)
		}
	}

	if len(cert.Certificate) == 0 {
		err = errors.New("Failed to parse certificate PEM data")
		return
	}

	var keyBlock *pem.Block
	for {
		keyBlock, keyData = pem.Decode(keyData)
		if keyBlock == nil {
			err = errors.New("Failed to parse key PEM data")
			return
		}
		if x509.IsEncryptedPEMBlock(keyBlock) {

			if password == "" {
				err = errors.New("Private key is password protected, but password was not specified")
				return
			}

			decryptedKeyBlock, decryptErr := x509.DecryptPEMBlock(keyBlock, []byte(password))
			if decryptErr != nil {
				err = decryptErr
				return
			}
			keyBlock.Bytes = decryptedKeyBlock
			break
		}
		if keyBlock.Type == "PRIVATE KEY" || strings.HasSuffix(keyBlock.Type, " PRIVATE KEY") {
			break
		}
	}

	cert.PrivateKey, err = parsePrivateKey(keyBlock.Bytes)

	return
}

func parsePrivateKey(der []byte) (crypto.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return key, nil
		default:
			return nil, errors.New("Found unknown private key type in PKCS#8 wrapping")
		}
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}

	return nil, errors.New("Failed to parse private key")
}
