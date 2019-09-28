package cmd

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func load(name string, app *Config) error {
	files := make([]string, 0, 5)

	writeDefault := false

	nconfig, ok := os.LookupEnv(strings.ToUpper(name) + "_CONF")
	if ok && nconfig != "" {
		files = append(files, nconfig)
		writeDefault = true
	} else {
		if configDir, err := os.UserConfigDir(); err == nil {
			writeDefault = true
			files = append(files, filepath.Join(configDir, name, name+".toml"))
			files = append(files, filepath.Join(configDir, name, name+".cfg"))
		}
		if homeDir, err := os.UserHomeDir(); err == nil {
			files = append(files, filepath.Join(homeDir, name+".toml"))
			files = append(files, filepath.Join(homeDir, name+".cfg"))
			files = append(files, filepath.Join(homeDir, "."+name+".toml"))
			files = append(files, filepath.Join(homeDir, "."+name+".cfg"))
		}
		files = append(files, name+".toml")
		files = append(files, name+".cfg")
	}

	for _, file := range files {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("%s: %w", file, err)
			}
			continue
		}

		md, err := toml.Decode(string(b), app)
		if err != nil {
			return fmt.Errorf("reading toml file %s: %w", file, err)
		}

		for _, k := range md.Undecoded() {
			return fmt.Errorf("extra config in file %s: %s", file, k.String())
		}

		return nil
	}

	if !writeDefault {
		return nil
	}

	// Initialise configuration path (if needed)
	if err := os.MkdirAll(path.Dir(files[0]), 0755); err != nil {
		return fmt.Errorf("creating config path %s: %w\n", path.Dir(files[0]), err)
	}

	// Persistent configuration
	if err := ioutil.WriteFile(files[0], []byte(DEFAULT_CONFIG_FILE), 0644); err != nil {
		return fmt.Errorf("failed writing config file %s: %w\n", files[0], err)
	}

	// Load default config
	if _, err := toml.Decode(DEFAULT_CONFIG_FILE, app); err != nil {
		return fmt.Errorf("Failed unmarshalling from default config: %w\n", err)
	}

	fmt.Printf("Created new configuration file '%s'. (See file or %s help config for more info)\n", files[0], name)

	return nil
}

const config_help = `Configuration File

The configuration file mirrors largely mirrors parameters available
on the command line.

There are two notable exceptions:
  - report.macros
    This section defines macros for the report (set of operations) that
    can be used on the command line arguments or by other macros. This
    is how complex reports can be built up.

  - importdefs.<name>
    Defining a new importdefs <name> will create a sub-command under
    import that has all of the command line parameters and the import
    configuration pre-configured. This allows you to create an import
    definition per CSV file (or other format) that you download.

  - register.accounts
    List of accounts that shell-completion will match if used. This is
    to make it easier to use the CLI.

Starter configuration file:
`

const ALL_CONFIG = config_help +
	"```" +
	DEFAULT_CONFIG_FILE +
	"```"

const DEFAULT_CONFIG_FILE = `

#
# Main defaults
#
#ledger =  "default_ledger_file"
#baseccy = "£"

#
# Defaults for the report command
#
[report]
combineby = "today"
type =      "Ansi"
sum =       true
convert =   true
credit =    "^(Income|Trading|Liability|Equity)(:.*)?$"

[report.macros]
macroA = [
	"book operation A",
	"book operation B",
]

#
# Defaults for the register command
#

[register]
accounts = [
  "Asset:Default",
]
count = -100
asc = true

[importdefs.bankformat]
description = "Bank Format"
configtype = "csv"
account = "Asset:BankDefaultAccount"
counteraccount = "Expense:DefaultExpenseAccount"
dedup = true
reclassify = true

[importdefs.bankformat.params]
ccy = "£"
date = "0"
amount = "2"
payee = "1"

`
