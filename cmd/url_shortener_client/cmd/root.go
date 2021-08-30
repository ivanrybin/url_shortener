package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

func NewCLI(address string) *cobra.Command {

	root := &cobra.Command{
		Use:                "url_shortener",
		Short:              "URL Shortener",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
			os.Exit(0)
		},
	}

	root.AddCommand(newCreateCmd(address))
	root.AddCommand(newGetCmd(address))

	return root
}
