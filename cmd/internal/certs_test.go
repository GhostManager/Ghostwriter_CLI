package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCertificatePackage(t *testing.T) {
	defer quietTests()()
	tempDir, err := os.MkdirTemp("", "gotest")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	t.Log("Testing `GenerateCertificatePackage()` and generating DH parameters can take several minutes...")
	GenerateCertificatePackage(tempDir)

	// Paths we expect to exist after generating the certificate package
	sslDir := filepath.Join(tempDir, "ssl")
	dhPath := filepath.Join(tempDir, "ssl", "dhparam.pem")
	keyPath := filepath.Join(tempDir, "ssl", "ghostwriter.key")
	crtPath := filepath.Join(tempDir, "ssl", "ghostwriter.crt")

	// Test if the `ssl` folder exists
	assert.True(t, DirExists(sslDir), "Expected `ssl` folder to exist")

	// Test if all certificate package files exist
	assert.True(t, FileExists(dhPath), "Expected `dhparam.pem` file to exist")
	assert.True(t, FileExists(keyPath), "Expected `ghostwriter.key` file to exist")
	assert.True(t, FileExists(crtPath), "Expected `ghostwriter.crt` file to exist")
}
