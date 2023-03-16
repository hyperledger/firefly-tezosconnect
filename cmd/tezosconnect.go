package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tezosconnect",
	Short: "Hyperledger FireFly Connector for Tezos blockchain",
	Long:  ``,
}

var cfgFile string

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "f", "", "config file")
}

func Execute() error {
	return rootCmd.Execute()
}