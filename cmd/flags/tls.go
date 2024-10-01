package flags

import (
	"crypto/tls"

	commonClient "github.com/aerospike/tools-common-go/client"
	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/pflag"
)

//nolint:govet // Padding not a concern for a CLI
type TLSFlags struct {
	Protocols        commonFlags.TLSProtocolsFlag
	RootCAFile       commonFlags.CertFlag
	RootCAPath       commonFlags.CertPathFlag
	CertFile         commonFlags.CertFlag
	KeyFile          commonFlags.CertFlag
	KeyFilePass      commonFlags.PasswordFlag
	HostnameOverride string
}

func NewTLSFlags() *TLSFlags {
	return &TLSFlags{
		Protocols: commonFlags.NewDefaultTLSProtocolsFlag(),
	}
}

// newTLSFlagSet returns a new pflag.FlagSet with TLS flags defined. Values
// are stored in the TLSFlags struct.
func (tf *TLSFlags) newTLSFlagSet() *pflag.FlagSet {
	f := &pflag.FlagSet{}

	f.Var(&tf.RootCAFile, TLSCaFile, "The CA used when connecting to AVS.")
	f.Var(&tf.RootCAPath, TLSCaPath, "A path containing CAs for connecting to AVS.")
	f.Var(&tf.CertFile, TLSCertFile, "The certificate file for mutual TLS authentication with AVS.")
	f.Var(&tf.KeyFile, TLSKeyFile, "The key file used for mutual TLS authentication with AVS.")
	f.Var(&tf.KeyFilePass, TLSKeyFilePass, "The password used to decrypt the key-file if encrypted.")
	f.Var(&tf.Protocols, TLSProtocols,
		"Set the TLS protocol selection criteria. This format is the same as"+
			" Apache's SSLProtocol documented at https://httpd.apache.org/docs/current/mod/mod_ssl.html#ssl protocol.",
	)
	f.StringVar(
		&tf.HostnameOverride,
		TLSHostnameOverride,
		"",
		"The hostname to use when validating the server certificate.",
	)

	return f
}

func (tf *TLSFlags) NewTLSConfig() (*tls.Config, error) {
	rootCA := [][]byte{}

	if len(tf.RootCAFile) != 0 {
		rootCA = append(rootCA, tf.RootCAFile)
	}

	rootCA = append(rootCA, tf.RootCAPath...)

	tlsConfig, err := commonClient.NewTLSConfig(
		rootCA,
		tf.CertFile,
		tf.KeyFile,
		tf.KeyFilePass,
		0,
		0,
	).NewGoTLSConfig()

	if err != nil {
		return nil, err
	}

	if tf.HostnameOverride != "" {
		tlsConfig.ServerName = tf.HostnameOverride
	}

	return tlsConfig, nil
}
