package download

import (
	"github.com/spf13/cobra"
)

type Download struct {
	Monzo *MonzoDownload
}

func Add(root *cobra.Command, download *Download) {
	ncmd := &cobra.Command{
		Use:               "download",
		Short:             "Download data from institutions",
		Long:              "Download data from institutions",
		DisableAutoGenTag: true,
	}
	ncmd.Args = cobra.NoArgs
	download.Monzo.Add(ncmd)
	root.AddCommand(ncmd)
}
