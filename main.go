// CLI main for goledger command line accounting
package main

import (
	"fmt"
	"github.com/mescanne/goledger/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
