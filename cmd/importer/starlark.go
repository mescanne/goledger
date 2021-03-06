package importer

import (
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/script"
	"go.starlark.net/starlark"
)

func (imp *ImportDef) processData(idata starlark.Value, file string, sc string) (*book.Book, error) {

	// Build the book
	bbuilder := book.NewBookBuilder()
	addf := script.GetBuilderFunction(bbuilder, imp.Account, imp.CCY, imp.CounterAccount)

	// This dictionary defines the pre-declared environment.
	predeclared := starlark.StringDict{
		"data":  idata,
		"file":  starlark.String(file),
		"add":   addf,
		"error": script.Errorf,
	}

	err := script.RunStarlark(script.Print, predeclared, sc)
	if err != nil {
		return nil, err
	}

	// Build the book -- done!
	return bbuilder.Build(), nil
}
