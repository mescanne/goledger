package cmd

import (
	"fmt"
	"github.com/mescanne/goledger/cmd/accounts"
	"github.com/mescanne/goledger/cmd/app"
	"github.com/mescanne/goledger/cmd/currencies"
	"github.com/mescanne/goledger/cmd/download"
	"github.com/mescanne/goledger/cmd/export"
	"github.com/mescanne/goledger/cmd/generate"
	"github.com/mescanne/goledger/cmd/importer"
	"github.com/mescanne/goledger/cmd/register"
	"github.com/mescanne/goledger/cmd/reports"
	"github.com/mescanne/goledger/cmd/utils"
	"github.com/mescanne/goledger/cmd/web"
	"github.com/spf13/cobra"
)

//
// Configuration object
//
type Config struct {
	app.App
	Report     reports.TransactionReport
	Register   register.RegisterReport
	ImportDefs map[string]*importer.ImportDef
	Generate   map[string]*generate.Generate
	Download   download.Download
	Web        web.WebConfig
	Export     export.ExportReport
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
	reports.Add(appCmd, &app.App, &app.Report)
	register.Add(appCmd, &app.App, &app.Register)
	importer.Add(appCmd, &app.App, app.ImportDefs)
	generate.Add(appCmd, &app.App, app.Generate)
	currencies.Add(appCmd, &app.App)
	export.Add(appCmd, &app.App, &app.Export)
	download.Add(appCmd, &app.Download)
	utils.AddShell(appCmd)
	utils.AddDocs(appCmd)

	// Add Web
	web.Add(appCmd, &app.Web, &web.WebApp{
		App:      app.App,
		Report:   &app.Report,
		Register: &app.Register,
	})

	// Run core app
	if err := appCmd.Execute(); err != nil {
		return err
	}

	return nil
}
