package flags

import (
	"crypto/tls"

	commonClient "github.com/aerospike/tools-common-go/client"
	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/pflag"
)

//nolint:govet // Padding not a concern for a CLI
type TLSFlags struct {
	TLSProtocols   commonFlags.TLSProtocolsFlag
	TLSRootCAFile  commonFlags.CertFlag
	TLSRootCAPath  commonFlags.CertPathFlag
	TLSCertFile    commonFlags.CertFlag
	TLSKeyFile     commonFlags.CertFlag
	TLSKeyFilePass commonFlags.PasswordFlag
}

func NewTLSFlags() *TLSFlags {
	return &TLSFlags{
		TLSProtocols: commonFlags.NewDefaultTLSProtocolsFlag(),
	}
}

// NewTLSFlagSet returns a new pflag.FlagSet with TLS flags defined. Values
// are stored in the TLSFlags struct.
func (tf *TLSFlags) NewTLSFlagSet(fmtUsage commonFlags.UsageFormatter) *pflag.FlagSet {
	f := &pflag.FlagSet{}

	f.Var(&tf.TLSRootCAFile, "tls-cafile", fmtUsage("The CA used when connecting to AVS."))
	f.Var(&tf.TLSRootCAPath, "tls-capath", fmtUsage("A path containing CAs for connecting to AVS."))
	f.Var(&tf.TLSCertFile, "tls-certfile", fmtUsage("The certificate file for mutual TLS authentication with AVS."))
	f.Var(&tf.TLSKeyFile, "tls-keyfile", fmtUsage("The key file used for mutual TLS authentication with AVS."))
	f.Var(&tf.TLSKeyFilePass, "tls-keyfile-password", fmtUsage("The password used to decrypt the key-file if encrypted."))
	f.Var(&tf.TLSProtocols, "tls-protocols", fmtUsage(
		"Set the TLS protocol selection criteria. This format is the same as"+
			" Apache's SSLProtocol documented at https://httpd.apache.org/docs/current/mod/mod_ssl.html#ssl protocol.",
	))

	return f
}

func (tf *TLSFlags) NewTLSConfig() (*tls.Config, error) {
	rootCA := [][]byte{}

	if len(tf.TLSRootCAFile) != 0 {
		rootCA = append(rootCA, tf.TLSRootCAFile)
	}

	rootCA = append(rootCA, tf.TLSRootCAPath...)

	return commonClient.NewTLSConfig(
		rootCA,
		tf.TLSCertFile,
		tf.TLSKeyFile,
		tf.TLSKeyFilePass,
		0,
		0,
	).NewGoTLSConfig()
}
