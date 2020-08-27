// CLI main for goledger command line accounting
package main

import (
	"fmt"
	"github.com/mescanne/goledger/cmd"
	"os"
)

//go:generate esc -o cmd/web/assets.go -pkg web -prefix static static

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
