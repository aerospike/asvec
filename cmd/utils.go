package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"golang.org/x/term"

	avs "github.com/aerospike/avs-client-go"
	"github.com/spf13/viper"
)

func passwordPrompt(prompt string) (string, error) {
	fmt.Print(prompt)

	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}

	fmt.Println()

	return string(bytePassword), nil
}

func createClientFromFlags(clientFlags *flags.ClientFlags) (*avs.Client, error) {
	hosts := parseBothHostSeedsFlag(clientFlags.Seeds, clientFlags.Host)
	isLoadBalancer := isLoadBalancer(clientFlags.Seeds)

	ctx, cancel := context.WithTimeout(context.Background(), clientFlags.Timeout)
	defer cancel()

	tlsConfig, err := clientFlags.NewTLSConfig()
	if err != nil {
		logger.Error("failed to create TLS config", slog.Any("error", err))
		return nil, err
	}

	var password *string

	if clientFlags.AuthCredentials.User.Val != nil {
		if *clientFlags.AuthCredentials.Password.Val != "" {
			strPass := clientFlags.AuthCredentials.Password.String()
			password = &strPass
		} else {
			pass, err := passwordPrompt("Enter Password: ")
			if err != nil {
				logger.Error("failed to read password", slog.Any("error", err))
				return nil, err
			}

			password = &pass
		}
	}

	var creds *avs.UserPassCredentials
	if clientFlags.AuthCredentials.User.Val != nil {
		creds = avs.NewCredentialsFromUserPass(*clientFlags.AuthCredentials.User.Val, *password)
	}

	client, err := avs.NewClient(
		ctx, hosts, clientFlags.ListenerName.Val, isLoadBalancer, creds, tlsConfig, logger,
	)
	if err != nil {
		if strings.Contains(err.Error(), "because it doesn't contain any IP SANs") {
			view.Printf("Hint: Failed to verify because of certificate hostname mismatch.")
			view.Printf("Hint: Either correctly set your certificate SAN or use --%s", flags.TLSHostnameOverride)
		}

		logger.Error("failed to create AVS client", slog.Any("error", err))
		return nil, err
	}

	return client, nil
}
func parseBothHostSeedsFlag(seeds *flags.SeedsSliceFlag, host *flags.HostPortFlag) avs.HostPortSlice {
	hosts := avs.HostPortSlice{}

	if len(seeds.Seeds) > 0 {
		logger.Debug(fmt.Sprintf("--%s is set, load-balancer will be DISABLED and tending WILL occur", flags.Seeds))

		hosts = append(hosts, seeds.Seeds...)
	} else {
		logger.Debug(fmt.Sprintf("--%s is set, load-balancer will be ENABLED and tending WILL NOT occur", flags.Host))

		hosts = append(hosts, &host.HostPort)
	}

	return hosts
}

func isLoadBalancer(seeds *flags.SeedsSliceFlag) bool {
	return len(seeds.Seeds) == 0
}

func nsAndSetString(namespace string, sets *string) string {
	var setStr string

	if sets == nil {
		setStr = "*"
	} else {
		setStr = *sets
	}

	return fmt.Sprintf("%s.%s", namespace, setStr)
}

func confirm(prompt string) bool {
	var confirm string

	fmt.Print(prompt + " (y/n): ")
	fmt.Scanln(&confirm)

	return strings.EqualFold(confirm, "y")
}

func checkSeedsAndHost() error {
	if viper.IsSet(flags.Seeds) && viper.IsSet(flags.Host) {
		return fmt.Errorf("only --%s or --%s allowed", flags.Seeds, flags.Host)
	}

	return nil
}
