/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	go_system_api "github.com/DeltaScratchpad/go-system-api"
	"github.com/DeltaScratchpad/webhook-interface/processing"
	webhook_tracker "github.com/DeltaScratchpad/webhook-interface/webhook-tracker"
	"github.com/spf13/viper"
	"os"

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
		_, _ = os.Stderr.WriteString("Listening for input from stdin.")
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

	var db_url = serverCmd.PersistentFlags().StringP("db-url", "d", "", "Database URL")
	_ = viper.BindPFlag("DB_URL", serverCmd.PersistentFlags().Lookup("db-url"))

	if db_url == nil || *db_url == "" {
		fmt.Println("No DB URL provided, using in-memory storage")
		state = webhook_tracker.NewLocalWebhookState()
	} else {
		fmt.Println("Using DB URL: ", *db_url)
		state = webhook_tracker.NewMySqlState(*db_url)
	}
}

func runStd() {
	if isInputFromPipe() {
		var query go_system_api.ProcessingEvent
		err := json.NewDecoder(os.Stdin).Decode(&query)
		if err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("Error: %s", err))
			return
		}
		defer func() {
			PushToStdOut(&query)
		}()
		processing.ProcessProcessingEvent(&query, state)

	} else {
		_, _ = os.Stderr.WriteString("No input from pipe.")
	}
}

func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}

func PushToStdOut(event *go_system_api.ProcessingEvent) {
	err := json.NewEncoder(os.Stdout).Encode(event)
	if err != nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("Error: %s", err))
		return
	}
}
