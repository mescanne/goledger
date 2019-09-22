package cmd

import (
	"fmt"
	"github.com/mescanne/goledger/cmd/accounts"
	"github.com/mescanne/goledger/cmd/app"
	"github.com/mescanne/goledger/cmd/currencies"
	"github.com/mescanne/goledger/cmd/importer"
	"github.com/mescanne/goledger/cmd/register"
	"github.com/mescanne/goledger/cmd/reports"
	"github.com/mescanne/goledger/cmd/utils"
	"github.com/spf13/cobra"
)

//
// Configuration object
//
type Config struct {
	app.App
	TransactionCmd reports.TransactionReport
	RegisterCmd    register.RegisterReport
	ImportDefs     []importer.ImportDef
}

//
// Execute command line program
//
func Execute() error {
	app := &Config{
		App: app.DefaultApp,
	}

	// Load configuration
	if err := load("goledger", app); err != nil {
		return fmt.Errorf("failed loading config: %s", err)
	}

	// Load core application
	appCmd := app.LoadCommand()

	// Load config help
	appCmd.AddCommand(&cobra.Command{
		Use:               "config",
		Short:             "Configuration file",
		Long:              ALL_CONFIG,
		DisableAutoGenTag: true,
	})

	// Add sub-commands
	accounts.Add(appCmd, &app.App)
	reports.Add(appCmd, &app.App, &app.TransactionCmd)
	register.Add(appCmd, &app.App, &app.RegisterCmd)
	importer.Add(appCmd, &app.App, app.ImportDefs)
	currencies.Add(appCmd, &app.App)
	utils.AddShell(appCmd)
	utils.AddDocs(appCmd)

	// Run core app
	if err := appCmd.Execute(); err != nil {
		return err
	}

	return nil
}
