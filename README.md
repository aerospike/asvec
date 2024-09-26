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

## Configuration File
All connection related command-line flags are available 

asvec.yml:
```
default:
  host: 127.0.0.1:5000
  credentials: admin:admin
  tls-cafile: ./ca.crt
  tls-certfile: ./cert.crt
  tls-keyfile: ./key.key
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


