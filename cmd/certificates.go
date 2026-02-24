package cmd

import (
	"fmt"

	certs "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// backupCmd represents the backup command
var certificatesCmd = &cobra.Command{
	Use:   "gencert",
	Short: "Create a new SSL/TLS certificate and DH param file for the Nginx web server",
	Long: `Creates a new SSL/TLS certificate and DH params files for the Nginx web server in the ssl/ directory. This
will not create a new certificate if the ssl/ghostwriter.key and ssl/ghostwriter.crt files already exist. Likewise, it
will not generate a new DH params file if the ssl/dhparam.pem file already exist.

Delete, move, or rename the files if you want to generate new ones.`,
	Run: createCertificates,
}

func init() {
	rootCmd.AddCommand(certificatesCmd)
}

func createCertificates(cmd *cobra.Command, args []string) {
	path := certs.GetDockerDirFromMode(mode)
	certErr := certs.GenerateCertificatePackage(path)
	if certErr == nil {
		fmt.Println("[+] Certificate generation complete!")
	}
}
