# Symbiosis Cloud CLI
--------------------

Easily manage all symbiosis related resources with our user-friendly CLI tool.

## Installation

### Homebrew
```bash
brew install symbiosis-cloud/tap/cli
```

### Linux

```bash
curl -s https://raw.githubusercontent.com/symbiosis-cloud/cli/main/install.sh | sh
```

## Usage

### Using your Symbiosis Cloud login

In order to get started using your Symbiosis Cloud login, simply run:

```bash
sym login
```

A browser window will open and ask you to login to your team. Once you're logged in your CLI client is ready.

### Using an API key
If you do not have a Symbiosis account already, you can head over to https://app.symbiosis.host/signup
to signup for a new account.

Once you are logged in you will need to create an API key with admin permissions.

1. Head over to https://app.symbiosis.host/api-keys to generate a new API key
2. run sym config init

## Functionality

Currently the CLI is in Beta. We only expose a limited number of CLI commands currently but are aiming for a more comprehensive set of features soon.

### Clusters

* Create cluster
* Delete cluster
* List clusters
* Describe cluster
* Get cluster identity

### Node pools

* Create node pool
* Delete node pool
* Describe node pool
* List node pools

### Autocomplete

You can run:

```bash
sym completion --help
```

To get instructions on how to setup autocomplete for your specific shell.