
# Goledger

## Overview

Goledger is yet-another-implementation of [Plain Text Accounting](https://plaintextaccounting.org/).

After discovering [ledger cli](https://ledger-cli.org) I've begun storing all my finances in text files.
However, there was intense motivation for finding other alternatives:
  * Customising reports for getting the right information was difficult
  * Extending the core ledger CLI is tough (complex C++ code)
  * I wanted a better way!

The key difference you can find in goledger:
 * Smaller and more opinionated feature set. It does far less. It recognises a subset of
   plain text accounting ledger files.
 * Report configuration allows for regex on accounts (re-mapping accounts).
 * Configured by a config file. This is where you customise your default reports.

## Documentation
 * [godoc](https://godoc.org/github.com/mescanne/goledger)
 * [goledger docs](docs/goledger.md)

## Installation

Note: Goledger requires go 1.13.

  go get -u github.com/mescanne/goledger

## Running

The first time you run it there will be a sample configuration file
written into $HOME/.config/goledger/goledger.toml. You can
move this to another location and specify it using GOLEGER_CONF 
environment variable.

This configuration file configures your experience using goledger,
including default arguments (location of ledger, base currency),
macros for making reports, import parsing definitions, registry
accounts for inspection, and more.

It is recommended you use the shell integration as this will use
your configuration file to generate the shell integration.
