/*
Copyright © 2024 Andrew Averell aaverell@daedev.net
*/
package cmd

import (
	"fmt"
	"os"

	webhook_tracker "github.com/DeltaScratchpad/webhook-interface/webhook-tracker"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var port *uint16
var state webhook_tracker.WebhookState

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "webhook-interface",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.webhook-interface.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	port = serverCmd.PersistentFlags().Uint16P("port", "p", 80, "Port to listen on")
	_ = viper.BindPFlag("PORT", serverCmd.PersistentFlags().Lookup("port"))

	var _ = serverCmd.PersistentFlags().StringP("db-url", "d", "", "Database URL")
	_ = viper.BindPFlag("DB_URL", serverCmd.PersistentFlags().Lookup("db-url"))
	var db_url_env = os.Getenv("DB_URL")
	db_url_lookup := serverCmd.PersistentFlags().Lookup("db-url")
	var db_url string
	if db_url_lookup == nil || db_url_lookup.Value.String() == "" {
		db_url = db_url_env
	} else {
		db_url = db_url_lookup.Value.String()
	}

	if db_url == "" {
		_, _ = os.Stderr.WriteString("No DB URL provided, using in-memory storage")
		state = webhook_tracker.NewLocalWebhookState()
	} else {
		_, _ = os.Stderr.WriteString("Using DB")
		state = webhook_tracker.NewMySqlState(db_url)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".webhook-interface" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".webhook-interface")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
