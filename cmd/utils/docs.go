package utils

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func AddDocs(root *cobra.Command) {
	ncmd := &cobra.Command{
		Use:               "docs",
		Short:             "Generate manpages",
		Long:              "Generate manpages",
		DisableAutoGenTag: true,
	}

	doctype := ncmd.Flags().StringP("type", "t", "manpage", "Format for documentation (manpage, markdown, yaml, or rest)")
	dir := ncmd.Flags().StringP("dir", "d", ".", "Directory for documentation")
	err := cobra.MarkFlagDirname(ncmd.Flags(), "dir")
	if err != nil {
		panic(fmt.Sprintf("failed marking dir as a dirname"))
	}
	ncmd.PreRunE = cobra.NoArgs
	ncmd.RunE = func(rcmd *cobra.Command, args []string) error {
		if *doctype == "manpage" {
			n := time.Now()
			hdr := &doc.GenManHeader{
				Title:   "title",
				Section: "section",
				Date:    &n,
				Source:  "source",
				Manual:  "manual",
			}
			return doc.GenManTree(root, hdr, *dir)
		} else if *doctype == "markdown" {
			err := doc.GenMarkdownTree(root, *dir)
			if err != nil {
				return err
			}
			for _, c := range root.Commands() {
				if !c.IsAdditionalHelpTopicCommand() {
					continue
				}
				basename := strings.Replace(c.CommandPath(), " ", "_", -1) + ".md"
				filename := filepath.Join(*dir, basename)
				err := ioutil.WriteFile(filename, []byte(c.Long+"\n\n"), os.ModePerm)
				if err != nil {
					return fmt.Errorf("error writing %s: %w", filename, err)
				}
			}
			return nil
		} else if *doctype == "yaml" {
			return doc.GenYamlTree(root, *dir)
		} else if *doctype == "rest" {
			return doc.GenReSTTree(root, *dir)
		}
		return fmt.Errorf("invalid documentation type %s: must be manpage, markdown, yaml, or rest", *doctype)
	}
	root.AddCommand(ncmd)
}
