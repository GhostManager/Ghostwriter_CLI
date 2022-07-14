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
	ssl_dir := filepath.Join(GetCwdFromExe(), "ssl")
	dh_path := filepath.Join(GetCwdFromExe(), "ssl", "dhparam.pem")
	key_path := filepath.Join(GetCwdFromExe(), "ssl", "ghostwriter.key")
	crt_path := filepath.Join(GetCwdFromExe(), "ssl", "ghostwriter.crt")

	// Test if the `ssl` folder exists
	assert.True(t, DirExists(ssl_dir), "Expected `ssl` folder to exist")

	// Test if all certificate package files exist
	assert.True(t, FileExists(dh_path), "Expected `dhparam.pem` file to exist")
	assert.True(t, FileExists(key_path), "Expected `ghostwriter.key` file to exist")
	assert.True(t, FileExists(crt_path), "Expected `ghostwriter.crt` file to exist")
}
