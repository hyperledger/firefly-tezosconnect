package cmd

import (
	"context"
	"fmt"

	"github.com/hyperledger/firefly-common/pkg/config"
	"github.com/spf13/cobra"
)

func configCommand() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "docs",
		Short: "Prints the config info as markdown",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			initConfig()
			b, err := config.GenerateConfigMarkdown(context.Background(), "", config.GetKnownKeys())
			fmt.Println(string(b))
			return err
		},
	}
	return versionCmd
}
