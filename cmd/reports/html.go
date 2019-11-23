package reports

import (
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"github.com/mescanne/goledger/cmd/utils"
	"math/big"
)

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
		zero := &big.Rat{}
		for _, v := range posts {
			thisLevel := v.GetAccountLevel()
			diff := "sameindent"
			if thisLevel > lastLevel {
				diff = "moreindent"
				b.Printf("  <div class=\"indentbox\">\n")
			} else if thisLevel < lastLevel {
				diff = "lessindent"
				b.Printf("  </div>\n")
			}

			amt := v.GetAmount()
			sign := "neg"
			if amt.Cmp(zero) >= 0 {
				sign = "pos"
			}

			b.Printf("  <div class=\"post indent%d %s\">\n", thisLevel, diff)
			b.Printf("    <div class=\"account\">%s</div>\n", v.GetAccountTerm())
			b.Printf("    <div class=\"amount %s\">%s%s</div>\n", sign, b.FormatSymbol(v.GetCCY()), b.FormatNumber(v.GetCCY(), amt))
			b.Printf("  </div>\n")

			lastLevel = thisLevel
		}
		b.Printf("</div>\n")

		// End of report
		b.Printf("</div>\n")
	}

	// End of reports
	b.Printf("</div>\n")
	b.Printf("</body></html>\n")

	return nil
}

const styleSheet = `
div {
        display: flex;
}

.reports {
        flex-direction: column;
        margin: 5px;
}

.report {
        flex-direction: column;
	font-family: sans-serif;
        padding: 15px;
	background: #e4f2f7;
	border-radius: 5px;
	font-weight: 900;
	font-size: large;
	color: #00009f;
}
.header {
        flex-direction: row;
        padding-bottom: 10px;
}
.payee {
        margin-left: 5px;
}
.post {
        flex-direction: row;
}
.post.sameindent {
	margin-top: 1px;
}
.posts {
        flex-direction: column;
}
.account {
	flex-grow: 1;
}

.indentbox {
	flex-direction: column;

	border-top: 1px solid black;
	padding-top: 15px;
	padding-bottom: 15px;
	padding-left: 15px;

	/* Smaller, less bold (thinner), and lighter (opacity) */
	font-size: 90%;
	font-weight: lighter;
	filter: opacity(0.75) grayscale(0.4);
}
`
