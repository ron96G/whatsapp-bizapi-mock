package util

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"time"

	"golang.org/x/crypto/pkcs12"
)

/*
	Implementation based on https://shaneutt.com/blog/golang-ca-and-signed-cert-go/

*/

// GenerateServerTLS uses @GenerateCertificate to generate a new CA and server cert/key
// These are then used to construct a TLS config
func GenerateServerTLS() (serverTLSConf *tls.Config, err error) {
	certPEM, certPrivKeyPEM, err := GenerateCertificate()
	if err != nil {
		return
	}

	serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivKeyPEM.Bytes())
	if err != nil {
		return
	}

	serverTLSConf = &tls.Config{
		Certificates: []tls.Certificate{serverCert},
	}
	return
}

// GenerateCertificate generates a new CA and returns a certPEM and privKeyPEM that it signed
func GenerateCertificate() (certPEM *bytes.Buffer, certPrivKeyPEM *bytes.Buffer, err error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization: []string{"WhatsApp Mockserver Fake CA"},
			Country:      []string{"DE"},
			Province:     []string{"NRW"},
			Locality:     []string{"Bonn"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	// generate server certificate

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"WhatsApp Mockserver Fake Certificate"},
			Country:      []string{"DE"},
			Province:     []string{"NRW"},
			Locality:     []string{"Bonn"},
		},
		DNSNames:     []string{""},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	// sign server certificate with ca

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM = new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM = new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	return certPEM, certPrivKeyPEM, nil
}

// FromPEM reads the raw pem file and returns a tls.Certificate
func FromPEM(in io.Reader) (cert tls.Certificate, err error) {
	raw, err := ioutil.ReadAll(in)
	if err != nil {
		return
	}

	for {
		block, rest := pem.Decode(raw)
		if block == nil {
			break
		}

		if block.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, block.Bytes)

		} else {
			cert.PrivateKey, err = parsePrivateKey(block.Bytes)
			if err != nil {
				err = fmt.Errorf("Failure reading private key: %s", err)
				return
			}
		}
		raw = rest
	}

	if len(cert.Certificate) == 0 {
		err = fmt.Errorf("No certificate found")
		return
	} else if cert.PrivateKey == nil {
		err = fmt.Errorf("No private key found")
		return
	}

	return
}

// FromP12 reads the raw, encrypted p21 keystore file and returns a tls.Certificate
func FromP12(in io.Reader, password string) (cert tls.Certificate, err error) {
	raw, err := ioutil.ReadAll(in)
	if err != nil {
		return
	}

	key, certificate, err := pkcs12.Decode(raw, password)
	if err != nil {
		return
	}

	cert.Certificate = append(cert.Certificate, certificate.Raw)
	cert.PrivateKey = key.(crypto.PrivateKey)

	if len(cert.Certificate) == 0 {
		err = fmt.Errorf("No certificate found")
		return
	} else if cert.PrivateKey == nil {
		err = fmt.Errorf("No private key found")
		return
	}

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
			return nil, fmt.Errorf("Found unknown private key type in PKCS#8 wrapping")
		}
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}
	return nil, fmt.Errorf("Failed to parse private key")
}
