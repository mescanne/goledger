package importer

import (
	"fmt"
	"github.com/antzucaro/matchr"
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"github.com/mescanne/goledger/cmd/reports"
	"github.com/mescanne/goledger/cmd/utils"
	"github.com/spf13/cobra"
	"os"
)

type ImportDef struct {

	// For custom imports
	Name        string
	Description string

	// Decoding of structure (CSV, JSON, etc)
	utils.CLIConfig

	// Code for conversion
	Code string

	// Defaults for code conversion
	Account        string
	CounterAccount string
	CCY            string

	// Rules to apply after import
	Dedup      bool
	Reclassify bool
}

func (imp *ImportDef) run(app *app.App, rcmd *cobra.Command, args []string) error {

	// Use default CCY (base CCY) if non specified for import
	if imp.CCY == "" {
		imp.CCY = app.BaseCCY
	}

	// Create the importer
	importer, err := NewBookImporterByConfig(&imp.CLIConfig)
	if err != nil {
		return fmt.Errorf("invalid import configuration for '%v': %w", &imp.CLIConfig, err)
	}

	// Load main book
	main, err := app.LoadBook()
	if err != nil {
		return fmt.Errorf("error loading book: %s", err)
	}

	// Track previous books
	prevbooks := make([]*book.Book, 0, len(args))

	// Iterate the import files
	for _, arg := range args {

		// Get the reader
		r := os.Stdin
		if arg != "-" {
			r, err = os.Open(arg)
			if err != nil {
				return fmt.Errorf("error opening %s: %w", arg, err)
			}
			defer r.Close()
		}

		// Load the records
		imports, err := importer(r)
		if err != nil {
			return fmt.Errorf("error importing data: %w", err)
		}

		// Convert the records
		b, err := imp.processData(imports, arg, imp.Code)
		if err != nil {
			return fmt.Errorf("error processing data: %w", err)
		}

		// Deduplication if needed (including previous books)
		if imp.Dedup {
			b.RemoveDuplicatesOf(main)
		}

		// Always deduplicate multiple files
		for _, prev := range prevbooks {
			b.RemoveDuplicatesOf(prev)
		}

		// Reclassification if needed
		if imp.Reclassify {
			b.ReclassifyByAccount(main, imp.Account, func(a, b string) float64 {
				d := matchr.JaroWinkler(a, b, true)
				if d > 0.5 {
					return d
				} else {
					return 0.0
				}
			})
		}

		// Use decimals of main book
		bp := app.NewBookPrinter(main.GetCCYDecimals())

		// Dump report ledger-style
		if err := reports.ShowLedger(bp, b.Transactions()); err != nil {
			return err
		}

		// Add to previous books
		prevbooks = append(prevbooks, b)
	}

	return nil
}

func (imp *ImportDef) add(name string, app *app.App) *cobra.Command {
	ncmd := &cobra.Command{
		Use:               name,
		Short:             imp.Description,
		Long:              "Import transactions",
		Args:              cobra.MinimumNArgs(1),
		DisableAutoGenTag: true,
	}
	ncmd.Flags().BoolVarP(&imp.Dedup, "dedup", "d", imp.Dedup, "deduplicate transactions based on payee and date")
	ncmd.Flags().BoolVarP(&imp.Reclassify, "reclassify", "r", imp.Reclassify, "reclassify the counteraccount based on previous transactions")
	if imp.Code == "" {
		ncmd.Flags().StringVar(&imp.Code, "code", imp.Code, "code for import or file:<file> for external code (see help code)")
		cobra.MarkFlagRequired(ncmd.Flags(), "code")
	}
	if imp.ConfigType == "" {
		ncmd.Flags().Var(&imp.CLIConfig, "format", "format of input (see help format)")
		cobra.MarkFlagRequired(ncmd.Flags(), "format")
	}
	ncmd.Flags().StringVarP(&imp.Account, "acct", "a", imp.Account, "account for imported postings")
	if imp.Account == "" {
		cobra.MarkFlagRequired(ncmd.Flags(), "acct")
	}
	ncmd.Flags().StringVarP(&imp.CounterAccount, "cacct", "c", imp.CounterAccount, "counteraccount for new transactions")
	if imp.CounterAccount == "" {
		cobra.MarkFlagRequired(ncmd.Flags(), "cacct")
	}
	ncmd.RunE = func(rcmd *cobra.Command, args []string) error {
		return imp.run(app, rcmd, args)
	}

	return ncmd
}

func Add(root *cobra.Command, app *app.App, config map[string]*ImportDef) {

	// Default -- where everything must be configured
	dftl := &ImportDef{
		Description: "Import transactions",
		Dedup:       true,
		Reclassify:  true,
	}
	ncmd := dftl.add("import", app)
	root.AddCommand(ncmd)

	// User customisation -- each customised for each source
	for name, def := range config {
		ncmd.AddCommand(def.add(name, app))
	}

	// Add in the help
	root.AddCommand(&cobra.Command{
		Use:               "format",
		Short:             "Import format help",
		Long:              ImportUsage,
		DisableAutoGenTag: true,
	})
}
