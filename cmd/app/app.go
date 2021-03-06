// Package app provides the root application capabilities for the
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
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"
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
	Ledger  string              // Location of ledger file
	BaseCCY string              // Conversion CCY for reporting
	Verbose bool                // Verbose modw
	Divider string              // Default (normally ":")
	Colour  bool                // Use Ansi Colour
	Macros  map[string][]string // Macros
	All     bool                // Use all accounts, rather than just accounts with a non-zero balance
	Lang    string              // Language for formatting
}

// Default configuration if none specified
var DefaultApp App = App{
	Ledger:  "main.ledger",
	Verbose: false,
	Lang:    "en",
	Divider: ":",
	Colour:  true,
	All:     false,
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

var appCleanup func() error = nil

func initialiseLess(w io.Writer) io.Writer {

	// Already initialised - skip
	if appCleanup != nil {
		return w
	}

	// If stdout is not a terminal, stop now
	if termWidth == -1 {
		return w
	}

	// TODO: Make this configurable from command line
	// TODO: Make this look into PAGER environment variable

	// Find less
	path, err := exec.LookPath("less")
	if err != nil {
		return w
	}

	// Configure less
	lesscmd := exec.Command(path, "-r")
	lesscmd.Stdout = os.Stdout
	lesscmd.Stderr = os.Stderr

	// Create a new stdin pipe
	stdin, err := lesscmd.StdinPipe()
	if err != nil {
		return w
	}

	// Start less
	if err := lesscmd.Start(); err != nil {
		return w
	}

	// Return cleanup
	appCleanup = func() error {
		if err := stdin.Close(); err != nil {
			return err
		}
		if err := lesscmd.Wait(); err != nil {
			return err
		}
		return nil
	}

	return stdin
}

// Load the root application cobra command
func (app *App) LoadCommand() *cobra.Command {

	var appCmd = &cobra.Command{
		Use:                    "goledger",
		Short:                  "goledger text-based account application",
		Long:                   goledger_long,
		BashCompletionFunction: custom_func,
		Version:                version,
		DisableAutoGenTag:      true,
		SilenceUsage:           true,
		SilenceErrors:          true,

		// Prior to the RunE method running, suppress any usage output
		// if there is an error -- at this point all CLI syntax-related
		// errors should be resolved. This is just for runtime errors.
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			//cmd.SilenceUsage = true
			cmd.SilenceErrors = true
		},

		// Cleanup if needed
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if appCleanup == nil {
				return
			}
			if err := appCleanup(); err != nil {
				fmt.Printf("error: %v\n", err)
			}
		},
	}

	appCmd.PersistentFlags().StringVarP(&app.Ledger, "ledger", "l", app.Ledger, "ledger to read")
	appCmd.PersistentFlags().StringVar(&app.BaseCCY, "ccy", app.BaseCCY, "base currency")
	appCmd.PersistentFlags().StringVar(&app.Divider, "divider", app.Divider, "divider for account components for reports")
	appCmd.PersistentFlags().StringVar(&app.Lang, "lang", app.Lang, "language")
	appCmd.PersistentFlags().BoolVar(&app.Verbose, "verbose", app.Verbose, "verbose")
	appCmd.PersistentFlags().BoolVar(&app.Colour, "colour", app.Colour, "colour (ansi) for reports")
	appCmd.PersistentFlags().BoolVar(&app.All, "all", app.All, "all accounts, not just non-zero balance")

	appCmd.AddCommand(&cobra.Command{
		Use:               "ops",
		Short:             "Operations on books",
		Long:              BookOperationUsage,
		DisableAutoGenTag: true,
	})

	appCmd.AddCommand(&cobra.Command{
		Use:               "macros",
		Short:             "Preconfigured macros for operations",
		Long:              mustResolveTemplate("macros", macroTemplate, app.Macros),
		DisableAutoGenTag: true,
	})

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

var macroTemplate = `Preconfigured macros
  {{ range $key, $ops := . }}
  Macro {{ $key }}
  {{- range $op := $ops }}
    {{ $op }}
  {{- end }}
  {{ end }}
`

func mustResolveTemplate(name string, templ string, data interface{}) string {
	t := template.New(name)
	t, err := t.Parse(templ)
	if err != nil {
		panic(fmt.Sprintf("template %s failed compiling, but is essential: %v", name, err))
	}
	var b strings.Builder
	err = t.Execute(&b, data)
	if err != nil {
		panic(fmt.Sprintf("template %s failed executing, but is essential: %v", name, err))
	}
	return b.String()

}
