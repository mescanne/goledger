// This package provides the root application capabilities for the
// goledger commandline interface.
//
// This includes loading the ledger, tools for printing the ledger,
// and common-configuration across all of goledger.
package app

import (
	"fmt"
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/loader"
	"github.com/spf13/cobra"
	"os"
)

const (
	custom_func = `
__goledger_handle_noun() {
    c=$((c+1))
}
`
)

// Configuration for an Application
type App struct {
	Ledger  string // Location of ledger file
	BaseCCY string // Conversion CCY for reporting
	Verbose bool   // Verbose modw
	Divider string // Default (normally ":")
	Colour  bool   // Use Ansi Colour
	Lang    string // Language for formatting
}

// Default configuration if none specified
var DefaultApp App = App{
	Ledger:  "main.ledger",
	Verbose: false,
	Lang:    "en",
	Divider: ":",
	Colour:  true,
}

func init() {
	l, ok := os.LookupEnv("LANG")
	if ok {
		DefaultApp.Lang = l
	}
}

// Load a book from the configured ledger file
func (app *App) LoadBook() (*book.Book, error) {
	bbuilder := book.NewBookBuilder()
	if err := loader.ParseFile(bbuilder, app.Ledger); err != nil {
		return nil, err
	}
	b := bbuilder.Build()
	return b, nil
}

const version = "0.1"

const goledger_long = `goledger is a text-based accounting.

It is in the same spirit as Plain Text Accounting (https://plaintextaccounting.org/)
and ledger cli (https://www.ledger-cli.org/)
`

// Load the root application cobra command
func (app *App) LoadCommand() *cobra.Command {
	var appCmd = &cobra.Command{
		Use:                    "goledger",
		Short:                  "goledger text-based account application",
		Long:                   goledger_long,
		BashCompletionFunction: custom_func,
		Version:                version,
		DisableAutoGenTag:      true,
	}

	appCmd.PersistentFlags().StringVarP(&app.Ledger, "ledger", "l", app.Ledger, "Ledger to read")
	appCmd.PersistentFlags().StringVar(&app.BaseCCY, "ccy", app.BaseCCY, "Base Currency")
	appCmd.PersistentFlags().StringVar(&app.Divider, "divider", app.Divider, "Divider for account components for reports")
	appCmd.PersistentFlags().StringVar(&app.Lang, "lang", app.Lang, "Language")
	appCmd.PersistentFlags().BoolVar(&app.Verbose, "verbose", app.Verbose, "Verbose")
	appCmd.PersistentFlags().BoolVar(&app.Colour, "colour", app.Colour, "Colour (ansi) for reports")

	appCmd.InitDefaultHelpCmd()
	appCmd.InitDefaultHelpFlag()
	appCmd.InitDefaultVersionFlag()

	// This should never happen. Only where ledger isn't a valid flag.
	if err := appCmd.MarkFlagFilename("ledger"); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(2)
	}

	return appCmd
}
