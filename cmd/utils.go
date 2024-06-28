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

func createClientFromFlags(clientFlags *flags.ClientFlags) (*avs.AdminClient, error) {
	hosts, isLoadBalancer := parseBothHostSeedsFlag(clientFlags.Seeds, clientFlags.Host)

	ctx, cancel := context.WithTimeout(context.Background(), clientFlags.Timeout)
	defer cancel()

	tlsConfig, err := clientFlags.NewTLSConfig()
	if err != nil {
		logger.Error("failed to create TLS config", slog.Any("error", err))
		return nil, err
	}

	var password *string
	if clientFlags.User.Val != nil {
		if len(clientFlags.Password) != 0 {
			strPass := clientFlags.Password.String()
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

	adminClient, err := avs.NewAdminClient(
		ctx, hosts, clientFlags.ListenerName.Val, isLoadBalancer, clientFlags.User.Val, password, tlsConfig, logger,
	)
	if err != nil {
		logger.Error("failed to create AVS client", slog.Any("error", err))
		return nil, err
	}

	return adminClient, nil
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

func nsAndSetString(namespace string, sets []string) string {
	var setStr string

	if len(sets) == 0 {
		setStr = "*"
	} else if len(sets) == 1 {
		setStr = sets[0]
	} else {
		setStr = fmt.Sprintf("%v", sets)
	}

	return fmt.Sprintf("%s.%s", namespace, setStr)
}

func confirm(prompt string) bool {
	var confirm string

	fmt.Print(prompt + " (y/n): ")
	fmt.Scanln(&confirm)

	return strings.ToLower(confirm) == "y"
}
