# Symbiosis Cloud CLI
--------------------

Easily manage all symbiosis related resources with our user-friendly CLI tool.

## Installation

### Homebrew
```bash
brew install symbiosis-cloud/tap/cli
```

### Linux

TODO

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