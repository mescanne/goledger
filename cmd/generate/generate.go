package generate

import (
	"fmt"
	"github.com/mescanne/goledger/cmd/app"
	"github.com/mescanne/goledger/cmd/reports"
	"github.com/spf13/cobra"
)

type Generate struct {
	Description string
	Code        string
	Dedup       bool
}

func (gen *Generate) run(app *app.App, rcmd *cobra.Command, args []string) error {

	// Load main book
	main, err := app.LoadBook()
	if err != nil {
		return fmt.Errorf("error loading book: %s", err)
	}

	// Generate the records from the main book
	b, err := gen.generate(main, gen.Code)
	if err != nil {
		return fmt.Errorf("error processing data: %w", err)
	}

	// Deduplication if needed (including previous books)
	if gen.Dedup {
		b.RemoveDuplicatesOf(main)
	}

	// Use decimals of main book
	bp := app.NewBookPrinter(main.GetCCYDecimals())

	// Dump report ledger-style
	if err := reports.ShowLedger(bp, b.Transactions()); err != nil {
		return err
	}

	return nil
}

func (gen *Generate) add(name string, app *app.App) *cobra.Command {
	ncmd := &cobra.Command{
		Use:               name,
		Short:             gen.Description,
		Long:              "Generate transactions",
		Args:              cobra.ExactArgs(0),
		DisableAutoGenTag: true,
	}
	ncmd.Flags().BoolVarP(&gen.Dedup, "dedup", "d", gen.Dedup, "deduplicate transactions based on payee and date")
	if gen.Code == "" {
		ncmd.Flags().StringVar(&gen.Code, "code", gen.Code, "code for generation (see help code)")
		cobra.MarkFlagRequired(ncmd.Flags(), "code")
	}
	ncmd.RunE = func(rcmd *cobra.Command, args []string) error {
		return gen.run(app, rcmd, args)
	}

	return ncmd
}

func Add(root *cobra.Command, app *app.App, config map[string]*Generate) {

	// Default -- where everything must be configured
	dftl := &Generate{
		Description: "Generate transactions",
		Dedup:       true,
	}
	ncmd := dftl.add("generate", app)
	root.AddCommand(ncmd)

	// User customisation -- each customised for each source
	for name, def := range config {
		ncmd.AddCommand(def.add(name, app))
	}
}
