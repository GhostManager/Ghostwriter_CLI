package internal

import (
	"path/filepath"
	"testing"
)

func TestGenerateCertificatePackage(t *testing.T) {
	defer quietTests()()

	t.Log("Testing `GenerateCertificatePackage()` and generating DH parameters can take several minutes...")
	GenerateCertificatePackage()

	// Paths we expect to exist after generating the certificate package
	ssl_dir := filepath.Join(GetCwdFromExe(), "ssl")
	dh_path := filepath.Join(GetCwdFromExe(), "ssl", "dhparam.pem")
	key_path := filepath.Join(GetCwdFromExe(), "ssl", "ghostwriter.key")
	crt_path := filepath.Join(GetCwdFromExe(), "ssl", "ghostwriter.crt")

	// Test if the `ssl` folder exists
	if !DirExists(ssl_dir) {
		t.Error("Could not find the `ssl` directory")
	}

	// Test if all certificate package files exist
	if !FileExists(dh_path) {
		t.Error("Failed to find the generated `dhparam.pem` file")
	}
	if !FileExists(key_path) {
		t.Error("Failed to find the generated `ghostwriter.key` file")
	}
	if !FileExists(crt_path) {
		t.Error("Failed to find the generated `ghostwriter.crt` file")
	}
}
