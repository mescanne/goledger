// CLI main for goledger command line accounting
package main

import (
	"fmt"
	"github.com/mescanne/goledger/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}
