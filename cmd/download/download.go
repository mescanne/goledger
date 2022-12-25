package download

import (
	"github.com/spf13/cobra"
)

type Download struct {
	Monzo    MonzoDownload
	Starling StarlingDownload
	Nordigen NordigenDownload
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
	download.Starling.Add(ncmd)
	download.Nordigen.Add(ncmd)
	root.AddCommand(ncmd)
}
