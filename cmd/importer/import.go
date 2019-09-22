package importer

import (
	"fmt"
	"github.com/antzucaro/matchr"
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"github.com/mescanne/goledger/cmd/reports"
	"github.com/mescanne/goledger/cmd/utils"
	"github.com/spf13/cobra"
	"io"
	"math/big"
	"os"
)

type ImportDef struct {
	Name           string
	Description    string
	Account        string
	CounterAccount string
	CCY            string
	Dedup          bool
	Reclassify     bool
	utils.CLIConfig
}

func (imp *ImportDef) run(app *app.App, rcmd *cobra.Command, args []string) error {

	// Use default CCY (base CCY) if non specified for import
	ccy := imp.CCY
	if ccy == "" {
		ccy = app.BaseCCY
	}

	// Create the importer
	importer, err := NewBookImporterByConfig(&imp.CLIConfig)
	if err != nil {
		return fmt.Errorf("invalid import configuration for '%v': %s", &imp.CLIConfig, err)
	}

	// Get the reader
	r, err := fileArgsToReader(args)
	if err != nil {
		return err
	}

	// Load the records
	imports, err := importer(r)
	if err != nil {
		return fmt.Errorf("error importing: %v", err)
	}

	// Build the book
	bbuilder := book.NewBookBuilder()
	for _, i := range imports {
		bbuilder.NewTransaction(i.Date, i.Payee, "")
		bbuilder.AddPosting(imp.Account, app.BaseCCY, i.Amount, "")
		bbuilder.AddPosting(imp.CounterAccount, app.BaseCCY, big.NewRat(0, 1).Neg(i.Amount), "")
	}
	b := bbuilder.Build()

	// Perform deduplication, re-classification if needed
	if imp.Dedup || imp.Reclassify {
		main, err := app.LoadBook()
		if err != nil {
			return fmt.Errorf("error loading book: %s", err)
		}

		if imp.Dedup {
			b.RemoveDuplicatesOf(main)
		}

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
	}

	bp := app.NewBookPrinter(rcmd.OutOrStdout(), b.GetCCYDecimals())

	// Dump report ledger-style
	return reports.ShowLedger(bp, b.Transactions())
}

func (imp *ImportDef) add(app *app.App) *cobra.Command {
	ncmd := &cobra.Command{
		Use:               imp.Name,
		Short:             imp.Description,
		Long:              "Import transactions", // TODO: Fill this in automatically
		Args:              cobra.MinimumNArgs(1),
		DisableAutoGenTag: true,
	}
	ncmd.Flags().BoolVarP(&imp.Dedup, "dedup", "d", imp.Dedup, "Deduplicate transactions based on payee and date")
	ncmd.Flags().BoolVarP(&imp.Reclassify, "reclassify", "r", imp.Reclassify, "Reclassify the counteraccount based on previous transactions")
	ncmd.Flags().Var(&imp.CLIConfig, "format", "Format of input (see help format)")
	if imp.CLIConfig.ConfigType == "" {
		cobra.MarkFlagRequired(ncmd.Flags(), "format")
	}
	ncmd.Flags().StringVarP(&imp.Account, "acct", "a", imp.Account, "Account for imported postings")
	if imp.Account == "" {
		cobra.MarkFlagRequired(ncmd.Flags(), "acct")
	}
	ncmd.Flags().StringVarP(&imp.CounterAccount, "cacct", "c", imp.CounterAccount, "Counteraccount for new transactions")
	if imp.CounterAccount == "" {
		cobra.MarkFlagRequired(ncmd.Flags(), "cacct")
	}
	ncmd.RunE = func(rcmd *cobra.Command, args []string) error {
		return imp.run(app, rcmd, args)
	}

	return ncmd
}

func Add(root *cobra.Command, app *app.App, config []ImportDef) {
	dftl := &ImportDef{
		Name:        "import",
		Description: "Import transactions",
		Dedup:       true,
		Reclassify:  true,
	}
	ncmd := dftl.add(app)
	root.AddCommand(ncmd)

	for _, def := range config {
		ncmd.AddCommand(def.add(app))
	}

	root.AddCommand(&cobra.Command{
		Use:               "format",
		Short:             "Import format syntax",
		Long:              ImportFormatUsage,
		DisableAutoGenTag: true,
	})

}

func fileArgsToReader(args []string) (io.Reader, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("error missing arguments: no input specified")
	}

	if len(args) == 1 && args[0] == "-" {
		return os.Stdin, nil
	}

	readers := make([]io.Reader, len(args))
	for i, arg := range args {
		r, err := os.Open(arg)
		if err != nil {
			return nil, fmt.Errorf("error opening %s: %w", arg, err)
		}
		readers[i] = r
	}

	// Multi reader
	return io.MultiReader(readers...), nil
}
