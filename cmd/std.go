/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	go_system_api "github.com/DeltaScratchpad/go-system-api"
	"github.com/DeltaScratchpad/webhook-interface/processing"

	"github.com/spf13/cobra"
)

// stdCmd represents the std command
var stdCmd = &cobra.Command{
	Use:   "std",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, _ = os.Stderr.WriteString("Webhook starting!.\n")
		runStd()
	},
}

func init() {
	rootCmd.AddCommand(stdCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stdCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stdCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func runStd() {
	if isInputFromPipe() {
		var query go_system_api.ProcessingEvent
		decoder := json.NewDecoder(os.Stdin)
		encoder := json.NewEncoder(os.Stdout)
		for {
			err := decoder.Decode(&query)
			if err != nil {
				if err.Error() == "EOF" {
					break
				} else {
					_, _ = os.Stderr.WriteString(fmt.Sprintf("Error when reading stdin: %s\n", err))
				}
			}
			processing.ProcessProcessingEvent(&query, state)
			query.Commands.Step += 1
			err = encoder.Encode(&query)
			if err != nil {
				_, _ = os.Stderr.WriteString(fmt.Sprintf("Error when reading stdout: %s\n", err))
				break
			}
		}

	} else {
		_, _ = os.Stderr.WriteString("No input from pipe.\n")
	}
}

func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}

func PushToStdOut(event *go_system_api.ProcessingEvent) {
	err := json.NewEncoder(os.Stdout).Encode(event)
	if err != nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("Error: %s\n", err))
		return
	}
}
