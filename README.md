[![codecov](https://codecov.io/gh/aerospike/asvec/graph/badge.svg?token=14G1LIEP2Q)](https://codecov.io/gh/aerospike/asvec)

# Aerospike Vector Search CLI Tool

## Overview

The Aerospike Vector Search (AVS) CLI Tool is designed to simplify the
management of AVS deployments. It offers features aimed at enhancing
efficiency and productivity for users getting started with vector search.

## Getting Started

### Prerequisites

Ensure you have an AVS instance up and running for `asvec` to connect to.
Check out the [AVS documentation](https://aerospike.com/docs/vector) for
instructions on getting started.

### Installation

#### Brew

You can install asvec with homebrew.
```shell
brew install aerospike/tools/asvec
```
or
```shell
brew tap aerospike/tools
brew install asvec
```

For more details see our [asvec homebrew formula](https://github.com/aerospike/homebrew-tools/blob/main/asvec.md).

#### Downloads

Download the latest release from [GitHub Releases](https://github.com/aerospike/asvec/releases).

### Basic Usage

To verify the installation and view available commands, execute:
```bash 
asvec --help
```

## Features

> [!NOTE]
> More features are in the works. Don't worry!

- **Data Browsing**: Easily run queries on an index.
- **Index Management**: Listing, creating, and dropping indexes.
- **User Management**: Listing, creating, and dropping users. Revoking and
  granting user's roles.
- **Node visibility**: Listing nodes and important metadata i.e. version, peers,
  etc.
- **Watch Mode**: Continuously monitor command output with automatic refresh using the `--watch` flag.

## Watch Mode

Some commands support a watch mode that continuously refreshes the output at a specified interval. This is useful for monitoring changes in real-time.

### Supported Commands

The following commands support watch mode:

- `node list` (and its alias `node ls`)
- `index list` (and its alias `index ls`)
- `query`

### Usage

To use watch mode, add the `--watch` flag to any supported command:

```bash
asvec node list --watch
```

By default, the command will refresh every 2 seconds. You can change the refresh interval using the `--watch-interval` flag:

```bash
asvec node list --watch --watch-interval 5  # Refresh every 5 seconds
```

Press Ctrl+C to exit watch mode.

## Configuration File
All connection related command-line flags can also be configured using a
configuration file. By default, the configuration file is installed at
`/etc/aerospike/asvec.yml`. Asvec checks for the existence of `asvec.yml` in
both `/etc/aerospike` and the current working directory. If your configuration
file is elsewhere use the `--config-file` flag.

To support multi-cluster scenarios the configuration file requires nesting keys
under the `cluster-name`. By default, when a configuration file is loaded the
`default` cluster name is used. To use a cluster-name other than `default` use
the `--cluster-name` flag.

Example asvec.yml:

```yaml
default:
  # Host address of the Aerospike server.
  # Uncomment and configure the 'host' field as needed.
  host: 127.0.0.1:5000              # Use host when using a load-balancer
  # seeds: 1.1.1.1:5000,2.2.2.2:5000  # Use seeds when not using a load-balancer
  
  # Credentials for authenticating with the Aerospike server.
  # Format: username:password
  credentials: admin:admin

  # TLS Configuration (optional)
  # Uncomment and provide the paths to the respective TLS files if secure communication is required.
  tls-cafile: ./ca.crt        # Path to the CA certificate file.
  tls-certfile: ./cert.crt    # Path to the client certificate file. (mtls)
  tls-keyfile: ./key.key      # Path to the client key file. (mtls)

# Additional cluster configuration example:
# cluster-2:
  # host: 192.168.0.1:5000
  # credentials: foo:bar
  # tls-cafile: ./other/ca.crt
  # tls-certfile: ./other/cert.crt
  # tls-keyfile: ./other/key.key
```

## Issues

If you encounter an issue feel free to open a GitHub issue or discussion.
Otherwise, if you are an enterprise customer, please [contact support](https://aerospike.com/support/)

## Developing
Before pushing your changes run the tests and run the linter.

### Running Tests
- Unit: `make unit`
- Integrations: `make integration`
- Coverage (Unit + Integration): `make coverage`

### Running Linter
`make lint`

### VSCode Setup
#### Debugging
Add the following to your .vscode/launch.json
```
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "main.go",
            "args": [
                "--log-level", "debug", // example: runs `node ls` with debug log level
                "node",
                "ls"
            ],
            "env": {
                "ASVEC_HOST": "localhost:10000"
            }
        }
    ]
}
```

#### Testing
```
{
    "go.testEnvVars": {
        "ASVEC_TEST_SUITES": "0,1,2,3", # Allows the selection of tests suites based on index
        "ASVEC_FAIL_FAST": "false" # Causes tests to fail immediately rather than finish the suites.
    },
    "go.testTags": "unit,integration,integration_large", 
    "go.buildFlags": [
        "-tags=integration,unit,integration_large"
    ],
}
```

## License

This project is licensed under the Apache License. See the
[LICENSE.md](./LICENSE) file for details.