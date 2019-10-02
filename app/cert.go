package app

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"time"
)

type CertificateAuthority struct {
	Key tls.Certificate
}

const (
	notAfterTime = time.Hour * 365 * 24
	rootUsage    = x509.KeyUsageDigitalSignature |
		x509.KeyUsageContentCommitment |
		x509.KeyUsageKeyEncipherment |
		x509.KeyUsageDataEncipherment |
		x509.KeyUsageKeyAgreement |
		x509.KeyUsageCertSign |
		x509.KeyUsageCRLSign
	hostAddr = "localhost"
)

func createCerts() (result, resultKey []byte) {

	var (
		currentTime time.Time
		cert        x509.Certificate
	)

	currentTime = time.Now()

	// Create default Cert struct
	cert = x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: hostAddr},
		NotBefore:             currentTime,
		NotAfter:              currentTime.Add(notAfterTime),
		KeyUsage:              rootUsage,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
		SignatureAlgorithm:    x509.ECDSAWithSHA512,
	}

	key, _ := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)

	// Create a raw version of cert
	rawResult, _ := x509.CreateCertificate(
		rand.Reader,
		&cert,
		&cert,
		key.Public(),
		key,
	)

	// Create a raw version of key
	rawKey, _ := x509.MarshalECPrivateKey(key)

	// Enode into []byte
	result = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: rawResult,
	})

	resultKey = pem.EncodeToMemory(&pem.Block{
		Type:  "ECDSA PRIVATE KEY",
		Bytes: rawKey,
	})

	return
}

func getKey() (key tls.Certificate, err error) {

	certRaw, keyRaw := createCerts()

	key, _ = tls.X509KeyPair(certRaw, keyRaw)

	ioutil.WriteFile("cert.pem", certRaw, 0400)
	ioutil.WriteFile("key.pem", keyRaw, 0400)

	key.Leaf, err = x509.ParseCertificate(key.Certificate[0])

	return
}
