package internal

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/Luzifer/go-dhparam"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const (
	// Bit size for Diffie-Helman parameters
	bitSize = 2048
)

// Callback function for “go-dhparam“.
func dhCallback(r dhparam.GeneratorResult) {
	switch r {
	case dhparam.GeneratorFoundPossiblePrime:
		os.Stdout.WriteString(".")
	case dhparam.GeneratorFirstConfirmation:
		os.Stdout.WriteString("+")
	case dhparam.GeneratorSafePrimeFound:
		os.Stdout.WriteString("*\n")
	}
}

// Generate Diffie-Helman parameters using “go-dhparam“.
func generateDHParam() ([]byte, error) {
	dh, err := dhparam.Generate(bitSize, dhparam.GeneratorTwo, dhCallback)
	if err != nil {
		return nil, err
	}
	bytes, err := dh.ToPEM()
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// Generate the Diffie-Helman parameters and then write the file to disk in the
// output directory (“outputDir“) with the specified “name“.
func writeDHParams(outputDir, name string) error {
	fileName := filepath.Join(outputDir, name+".pem")
	if FileExists(fileName) {
		fmt.Printf("[*] Skipping DH params because %s already exists\n", fileName)
		return nil
	}
	fmt.Println("[*] Generating a new `dhparam.pem` file (this could take a few minutes)")
	b, err := generateDHParam()
	if err != nil {
		return err
	}
	fmt.Printf("[+] Writing DH parameters to %s\n", fileName)
	if err := os.WriteFile(fileName, b, 0644); err != nil {
		return err
	}
	return nil
}

// Check if the SSL certificates are present in the specified “certPath“ and “keyPath“.
func checkCerts(certPath string, keyPath string) error {
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return err
	} else if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return err
	}
	return nil
}

// Generate the TLS certificates and Diffie-Helman parameters file using Go.
func generateCertificates() error {
	certPath := filepath.Join(GetCwdFromExe(), "ssl", "ghostwriter.crt")
	keyPath := filepath.Join(GetCwdFromExe(), "ssl", "ghostwriter.key")
	if checkCerts(certPath, keyPath) == nil {
		fmt.Printf("[!] Found existing certificate files, so new ones will not be generated...\n")
		fmt.Printf("[*] Rename or delete ssl/ghostwriter.key and ssl/ghostwriter.key if you want to replace these keys")
		return nil
	}
	fmt.Printf("[*] Did not find existing TLS/SSL certs for the Nginx container, so generating them now...\n")

	// Generate the ECDSA private key
	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		log.Printf("Failed to generate private key: %s\n", err)
		return err
	}

	// Set dates to today and one year from now
	notBefore := time.Now()
	oneYear := 365 * 24 * time.Hour
	notAfter := notBefore.Add(oneYear)

	// Generate a serial number for the certificate
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Printf("Failed to generate the serial number: %s\n", err)
		return err
	}

	// Template the certificate with necessary values
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Ghostwriter"},
			CommonName:   "nginx",
		},
		SignatureAlgorithm: x509.ECDSAWithSHA384,
		PublicKeyAlgorithm: x509.ECDSA,
		NotBefore:          notBefore,
		NotAfter:           notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create the certificate using our private key and the template
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Printf("Failed to create certificate: %s\n", err)
		return err
	}

	certOut, err := os.Create(certPath)
	if err != nil {
		log.Printf("Failed to open %s for writing: %s\n", certPath, err)
		return err
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("Failed to open %s for writing: %s/n", keyPath, err)
		return err
	}
	marshalKey, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		log.Printf("Unable to marshal ECDSA private key: %v\n", err)
		return err
	}
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: marshalKey})
	keyOut.Close()
	fmt.Printf("[+] Successfully generated new TLS/SSL certificates\n")

	return nil
}

// GenerateCertificatePackage generate TLS certificates and Diffie-Helman parameters file using Go.
func GenerateCertificatePackage() error {
	// Ensure the ``ssl`` directory exists to receive the keys
	sslPath := filepath.Join(GetCwdFromExe(), "ssl")
	if !DirExists(sslPath) {
		err := os.MkdirAll(sslPath, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to make the `ssl` directory")
		}
		fmt.Println("[+] Successfully made the `ssl` directory")
	}

	fmt.Println("[*] Generating new `ghostwriter.crt` and `ghostwriter.key` files")
	certErr := generateCertificates()
	if certErr != nil {
		fmt.Printf("[!] Failed to generate TLS/SSL certificate files: %s\n", certErr)
	}

	dhErr := writeDHParams(sslPath, "dhparam")
	if dhErr != nil {
		fmt.Printf("[!] Failed to generate Diffie-Helman parameters: %s\n", dhErr)
	}

	return nil
}
