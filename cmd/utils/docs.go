package utils

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"time"
)

func AddDocs(root *cobra.Command) {
	ncmd := &cobra.Command{
		Use:               "docs",
		Short:             "Generate manpages",
		Long:              "Generate manpages",
		DisableAutoGenTag: true,
	}

	shell := ncmd.Flags().StringP("type", "t", "manpage", "Format for documentation (manpage, markdown, yaml, or rest)")
	dir := ncmd.Flags().StringP("dir", "d", ".", "Directory for documentation")
	err := cobra.MarkFlagDirname(ncmd.Flags(), "dir")
	if err != nil {
		panic(fmt.Sprintf("failed marking dir as a dirname"))
	}
	ncmd.PreRunE = cobra.NoArgs
	ncmd.RunE = func(rcmd *cobra.Command, args []string) error {
		if *shell == "manpage" {
			n := time.Now()
			hdr := &doc.GenManHeader{
				Title:   "title",
				Section: "section",
				Date:    &n,
				Source:  "source",
				Manual:  "manual",
			}
			return doc.GenManTree(root, hdr, *dir)
		} else if *shell == "markdown" {
			return doc.GenMarkdownTree(root, *dir)
		} else if *shell == "yaml" {
			return doc.GenYamlTree(root, *dir)
		} else if *shell == "rest" {
			return doc.GenReSTTree(root, *dir)
		}
		return fmt.Errorf("invalid documentation type %s: must be manpage, markdown, yaml, or rest", *shell)
	}
	root.AddCommand(ncmd)
}
