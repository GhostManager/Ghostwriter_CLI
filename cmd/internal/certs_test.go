package internal

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestGenerateCertificatePackage(t *testing.T) {
	defer quietTests()()

	t.Log("Testing `GenerateCertificatePackage()` and generating DH parameters can take several minutes...")
	GenerateCertificatePackage()

	// Paths we expect to exist after generating the certificate package
	sslDir := filepath.Join(GetCwdFromExe(), "ssl")
	dhPath := filepath.Join(GetCwdFromExe(), "ssl", "dhparam.pem")
	keyPath := filepath.Join(GetCwdFromExe(), "ssl", "ghostwriter.key")
	crtPath := filepath.Join(GetCwdFromExe(), "ssl", "ghostwriter.crt")

	// Test if the `ssl` folder exists
	assert.True(t, DirExists(sslDir), "Expected `ssl` folder to exist")

	// Test if all certificate package files exist
	assert.True(t, FileExists(dhPath), "Expected `dhparam.pem` file to exist")
	assert.True(t, FileExists(keyPath), "Expected `ghostwriter.key` file to exist")
	assert.True(t, FileExists(crtPath), "Expected `ghostwriter.crt` file to exist")
}
