/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/DeltaScratchpad/webhook-interface/server"
	webhook_tracker "github.com/DeltaScratchpad/webhook-interface/webhook-tracker"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Creating server")
		done := make(chan os.Signal, 1)
		signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		var setPort uint16
		if port == nil {
			setPort = 80
		} else {
			setPort = *port
		}

		server.CreateServer(nil, fmt.Sprintf("%d", setPort), done, state)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	port = serverCmd.PersistentFlags().Uint16P("port", "p", 80, "Port to listen on")
	_ = viper.BindPFlag("PORT", serverCmd.PersistentFlags().Lookup("port"))

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
