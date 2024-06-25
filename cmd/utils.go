package cmd

import (
	"asvec/cmd/flags"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"time"

	avs "github.com/aerospike/avs-client-go"
)

func createClientFromFlags(clientFlags *flags.ClientFlags, connectTimeout time.Duration) (*avs.AdminClient, error) {
	hosts, isLoadBalancer := parseBothHostSeedsFlag(clientFlags.Seeds, clientFlags.Host)

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	tlsConfig, err := clientFlags.NewTLSConfig()
	if err != nil {
		logger.Error("failed to create TLS config", slog.Any("error", err))
		return nil, err
	}

	var password *string
	if len(clientFlags.Password) != 0 {
		strPass := clientFlags.Password.String()
		password = &strPass
	}

	adminClient, err := avs.NewAdminClient(
		ctx, hosts, clientFlags.ListenerName.Val, isLoadBalancer, clientFlags.User.Val, password, tlsConfig, logger,
	)
	if err != nil {
		logger.Error("failed to create AVS client", slog.Any("error", err))
		return nil, err
	}

	return adminClient, nil
}

func newTLSConfig(rootCA [][]byte, cert []byte, key []byte, keyPass []byte, tlsProtoMin int, tlsProtoMax int) (*tls.Config, error) {
	if len(rootCA) == 0 && len(cert) == 0 && len(key) == 0 {
		return nil, nil
	}

	var (
		clientPool []tls.Certificate
		serverPool *x509.CertPool
		err        error
	)

	serverPool = loadCACerts(rootCA)

	if len(cert) > 0 || len(key) > 0 {
		clientPool, err = loadServerCertAndKey(cert, key, keyPass)
		if err != nil {
			return nil, fmt.Errorf("failed to load client authentication certificate and key `%s`", err)
		}
	}

	tlsConfig := &tls.Config{ //nolint:gosec // aerospike default tls version is TLSv1.2
		Certificates:             clientPool,
		RootCAs:                  serverPool,
		InsecureSkipVerify:       false,
		PreferServerCipherSuites: true,
		MinVersion:               uint16(tlsProtoMin),
		MaxVersion:               uint16(tlsProtoMax),
	}

	return tlsConfig, nil
}

// loadCACerts returns CA set of certificates (cert pool)
// reads CA certificate based on the certConfig and adds it to the pool
func loadCACerts(certsBytes [][]byte) *x509.CertPool {
	certificates, err := x509.SystemCertPool()
	if certificates == nil || err != nil {
		certificates = x509.NewCertPool()
	}

	for _, cert := range certsBytes {
		if len(cert) > 0 {
			certificates.AppendCertsFromPEM(cert)
		}
	}

	return certificates
}

// loadServerCertAndKey reads server certificate and associated key file based on certConfig and keyConfig
// returns parsed server certificate
// if the private key is encrypted, it will be decrypted using key file passphrase
func loadServerCertAndKey(certFileBytes, keyFileBytes, keyPassBytes []byte) ([]tls.Certificate, error) {
	var certificates []tls.Certificate

	// Decode PEM data
	keyBlock, _ := pem.Decode(keyFileBytes)

	if keyBlock == nil {
		return nil, fmt.Errorf("failed to decode PEM data for key or certificate")
	}

	// Check and Decrypt the Key Block using passphrase
	if x509.IsEncryptedPEMBlock(keyBlock) { //nolint:staticcheck,lll // This needs to be addressed by aerospike as multiple projects require this functionality
		decryptedDERBytes, err := x509.DecryptPEMBlock(keyBlock, keyPassBytes) //nolint:staticcheck,lll // This needs to be addressed by aerospike as multiple projects require this functionality
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt PEM Block: `%s`", err)
		}

		keyBlock.Bytes = decryptedDERBytes
		keyBlock.Headers = nil
	}

	// Encode PEM data
	keyPEM := pem.EncodeToMemory(keyBlock)

	if keyPEM == nil {
		return nil, fmt.Errorf("failed to encode PEM data for key or certificate")
	}

	cert, err := tls.X509KeyPair(certFileBytes, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to add certificate and key to the pool: `%s`", err)
	}

	certificates = append(certificates, cert)

	return certificates, nil
}

func parseBothHostSeedsFlag(seeds *flags.SeedsSliceFlag, host *flags.HostPortFlag) (avs.HostPortSlice, bool) {
	isLoadBalancer := false
	hosts := avs.HostPortSlice{}

	if len(seeds.Seeds) > 0 {
		logger.Debug("seeds is set")

		hosts = append(hosts, seeds.Seeds...)
	} else {
		logger.Debug("hosts is set")

		isLoadBalancer = true

		hosts = append(hosts, &host.HostPort)
	}

	return hosts, isLoadBalancer
}
