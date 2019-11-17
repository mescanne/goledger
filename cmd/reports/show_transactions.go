package reports

import (
	"encoding/json"
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"github.com/mescanne/goledger/cmd/utils"
	"unicode/utf8"
)

func ShowLedger(b *app.BookPrinter, trans []book.Transaction) error {
	for _, posts := range trans {
		tnote := posts.GetTransactionNote()
		if tnote != "" {
			tnote = "  ; " + tnote
		}
		b.Printf("%s %s%s\n", posts.GetDate(), posts.GetPayee(), tnote)
		for _, p := range posts {
			pnote := p.GetPostNote()
			if pnote != "" {
				pnote = "  ; " + pnote
			}
			b.Printf("  %s  %s%s%s\n", p.GetAccount(), p.GetCCY(), b.FormatNumber(p.GetCCY(), p.GetAmount()), pnote)
		}
		b.Printf("\n")
	}
	return nil
}

func ShowJsonLedger(b *app.BookPrinter, trans []book.Transaction, pretty bool) error {
	var ob []byte
	var err error
	if pretty {
		ob, err = json.MarshalIndent(&trans, "", "    ")
	} else {
		ob, err = json.Marshal(&trans)
	}
	if err != nil {
		return err
	}

	if _, err = b.Write(ob); err != nil {
		return err
	}

	return nil
}

func showHTMLReport(b *app.BookPrinter, posts book.Transaction) error {
	//
	// Header
	//
	b.Printf("<div class=\"header\">\n")
	b.Printf("  <div class=\"date\">%s</div>\n", posts.GetDate())
	payee := posts.GetPayee()
	if payee != "" {
		b.Printf("  <div class=\"payee\">%s</div>\n", payee)
	}
	b.Printf("</div>\n")

	//
	// Posts
	//
	b.Printf("<div class=\"posts\">\n")
	lastLevel := 0
	for _, v := range posts {
		thisLevel := v.GetAccountLevel()
		diff := "sameindent"
		if thisLevel > lastLevel {
			diff = "moreindent"
		} else if thisLevel < lastLevel {
			diff = "lessindent"
		}

		b.Printf("  <div class=\"post indent%d %s\">", thisLevel, diff)
		b.Printf("    <div class=\"account\">%s</div>\n", v.GetAccountTerm())
		b.Printf("    <div class=\"amount\">%s%s</div>\n", b.FormatSymbol(v.GetCCY()), b.FormatNumber(v.GetCCY(), v.GetAmount()))
		b.Printf("  </div>\n")

		lastLevel = thisLevel
	}
	b.Printf("</div>\n")

	return nil
}

func ShowHTMLTransactions(b *app.BookPrinter, trans []book.Transaction, HTMLCSS string) error {

	if HTMLCSS == "" {
		HTMLCSS = styleSheet
	} else {
		var err error
		HTMLCSS, err = utils.GetFileOrStr(HTMLCSS)
		if err != nil {
			return err
		}
	}

	b.Printf("<html><head><style>\n%s\n</style></head><body>\n", HTMLCSS)
	b.Printf("<div class=\"reports\">\n")

	for _, posts := range trans {
		b.Printf("<div class=\"report\">\n")
		showHTMLReport(b, posts)
		b.Printf("</div>\n")
	}

	b.Printf("</div>\n")
	b.Printf("</body></html>\n")

	return nil
}

func ShowTransactions(b *app.BookPrinter, trans []book.Transaction) error {

	for _, posts := range trans {

		// Show header
		payee := posts.GetPayee()
		if payee != "" {
			b.Printf("\n%s - %s\n", posts.GetDate(), posts.GetPayee())
		} else {
			b.Printf("\n%s\n", posts.GetDate())
		}

		// Find max length for account names
		l := posts.MaxAccountTerm(2)

		// Find max length for numbers
		maxlen := 0
		for _, v := range posts {
			rv := b.FormatNumber(v.GetCCY(), v.GetAmount())
			nl := utf8.RuneCountInString(v.GetCCY()) + len(rv) + 1
			if nl > maxlen {
				maxlen = nl
			}
		}

		for _, v := range posts {

			// Print out account levels
			parts := v.GetAccountLevel()
			term := v.GetAccountTerm()
			if parts == 0 {
				if b.Colour() {
					b.Printf("\n")
				}
				b.Printf("%s", b.BlueUL(b.Sprintf("%-*.*s    ", l, l, term)))
			} else {
				for i := 0; i < parts; i++ {
					b.Printf("  ")
				}
				b.Printf("%-*.*s    ", l, l, term)
			}

			// Print out money
			b.Printf("%s\n", b.FormatMoney(v.GetCCY(), v.GetAmount(), maxlen))
		}
	}

	return nil
}

const styleSheet = `
div {
        display: flex;
}
.reports {
        flex-direction: column;
        margin: 10px;
}
.report {
	font-family: sans-serif;
        flex-direction: column;
        padding-bottom: 20px;
	padding-right: 100px;
}
.header {
        flex-direction: row;
        padding-bottom: 10px;
	font-size: large;
}
.date {
}
.payee {
        margin-left: 5px;
}
.post {
        flex-direction: row;
	color: gray;
}
.posts {
        flex-direction: column;
}
.account {
        flex-grow: 1;
}
.amount {
        justify-content: right;
	/* font-family: monospace; */
}
.account {
	flex-grow: 1;
}
.indent0 {
	padding-top: 20px;
	border-bottom: 2px solid #aaa;
	margin-bottom: 5px;
	font-size: large;
	color: black;
	font-weight: bold;
}
.indent0 .account {
	margin-left: 0px;
}
.indent0 .amount {
	font-family: sans-serif;
	margin-right: 0px;
}

.indent1 {
	padding-top: 10px;
}
.indent1 {
	font-style: bold;
	font-size: medium;
}
.indent1 .account {
	margin-left: 0px;
}
.indent1 .amount {
	margin-right: 0px;
}
.indent2 {
	font-size: small;
}
.indent2 .account {
	margin-left: 20px;
}
.indent2 .amount {
	margin-right: 20px;
}
.indent3 {
	font-size: smaller;
}
.indent3 .account {
	margin-left: 40px;
}
.indent3 .amount {
	margin-right: 40px;
}
.indent4 .account {
	margin-left: 60px;
}
.indent4 .amount {
	margin-right: 60px;
}
.indent5 .account {
	margin-left: 80px;
}
.indent5 .amount {
	margin-right: 80px;
}
`
