package generate

import (
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/script"
	"go.starlark.net/starlark"
)

func (gen *Generate) generate(b *book.Book, sc string) (*book.Book, error) {

	bbuilder := book.NewBookBuilder()
	addf := script.GetBuilderFunction(bbuilder, "acct", "ccy", "cacct")
	globals := map[string]starlark.Value{
		"data":  script.ConvertBookToStarlark(b),
		"add":   addf,
		"error": script.Errorf,
	}

	err := script.RunStarlark(script.Print, globals, sc)
	if err != nil {
		return nil, err
	}
	return bbuilder.Build(), nil
}
