package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Generate TLS certificates using ``openssl``.
func GenerateCertificatePackage() error {
	// Ensure the ``ssl`` directory exists to receive the keys
	if !DirExists(filepath.Join(GetCwdFromExe(), "ssl")) {
		err := os.MkdirAll(filepath.Join(GetCwdFromExe(), "ssl"), os.ModePerm)
		if err != nil {
			log.Fatalf("failed to make the `ssl` directory")
		}
		fmt.Println("[+] Successfully made the `ssl` directory")
	}

	// Generate SSL key and Diffie-Helman parameters files (if ``openssl`` can be found in $PATH)
	makeKeypair := true
	makeDHParam := true
	if CheckPath("openssl") {
		fmt.Println("[*] Found `openssl` in PATH, so proceeding with keypair generation")
		// Check if any certificate files exist already
		if FileExists(filepath.Join(GetCwdFromExe(), "ssl", "ghostwriter.crt")) ||
			FileExists(filepath.Join(GetCwdFromExe(), "ssl", "ghostwriter.key")) {
			overwriteCrt := AskConfirm(
				"[*] One or both of the `ghostwriter.crt` or `ghostwriter.key` files already " +
					"exist in the `ssl` directory. Do you want to overwrite them?",
			)
			if overwriteCrt {
				fmt.Println("[*] Generating new `ghostwriter.crt` and `ghostwriter.key` files")
			} else {
				fmt.Println("[*] Skipping `ghostwriter.crt` and `ghostwriter.key` generation")
				makeKeypair = false
			}
		}
		// Generate the certificate files with ``openssl``
		if makeKeypair {
			generateCertificates()
		}
		// Check if a DH parameters file already exists
		if FileExists(filepath.Join(GetCwdFromExe(), "ssl", "dhparam.pem")) {
			overwriteKey := AskConfirm(
				"[*] The `dhparam.pem` file already exists in the `ssl` directory. " +
					"Do you want to overwrite it?",
			)
			if overwriteKey {
				fmt.Println("[*] Generating a new `dhparam.pem` file (this could take a few minutes)")
			} else {
				fmt.Println("[*] Skipping `dhparam.pem` generation")
				makeDHParam = false
			}
		}
		// Generate the DH parameters file with ``openssl``
		if makeDHParam {
			generateDHParam("2048")
		}
	} else {
		fmt.Println("[!] Did not find `openssl` in the PATH")
	}

	return nil
}

// Generate the TLS certificates using ``openssl``.
func generateCertificates() {
	name := "openssl"
	args := []string{
		"req", "-new", "-newkey", "rsa:4096", "-days", "365", "-nodes", "-x509",
		"-subj", "/C=/ST=/L=/O=Ghostwriter/CN=ghostwriter.local",
		"-keyout", "ssl/ghostwriter.key", "-out", "ssl/ghostwriter.crt",
	}
	RunCmd(name, args)
}

// Generate a Diffie-Helman parameters file of the specified ``size`` using ``openssl``.
// The ``size`` should be a string, such as "2048" or "4096". See reference to determine
// desired size: https://wiki.mozilla.org/Security/Server_Side_TLS
func generateDHParam(size string) {
	name := "openssl"
	args := []string{"dhparam", "-out", "ssl/dhparam.pem", size}
	RunCmd(name, args)
}
