/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/prometheus-operator/poctl/internal/log"
	"github.com/spf13/cobra"
)

// stackCmd represents the stack command.
var stackCmd = &cobra.Command{
	Use:   "stack",
	Short: "create a stack of Prometheus Operator resources.",
	Long:  `create a stack of Prometheus Operator resources.`,
	Run: func(_ *cobra.Command, _ []string) {
		logger, err := log.NewLogger()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		logger.Info("stack command called")
	},
}

func init() {
	createCmd.AddCommand(stackCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stackCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stackCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
