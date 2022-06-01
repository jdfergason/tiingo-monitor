/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Eod struct {
	DateStr       string  `json:"date"`
	Ticker        string  `json:"ticker"`
	CompositeFigi string  `json:"compositeFigi"`
	Open          float32 `json:"open"`
	High          float32 `json:"high"`
	Low           float32 `json:"low"`
	Close         float32 `json:"close"`
	Volume        float32 `json:"volume"`
	Dividend      float32 `json:"divCash"`
	Split         float32 `json:"splitFactor"`
}

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tiingo-monitor",
	Short: "Monitor when tiingo mutual fund prices become available",
	Long: `Periodically download the most recent EOD quote for a mutual
fund and record it's date vs the actual date`,
	Run: func(cmd *cobra.Command, args []string) {
		ticker := viper.GetString("tiingo.ticker")
		now := time.Now()

		client := resty.New()
		var eodList []Eod
		url := fmt.Sprintf("https://api.tiingo.com/tiingo/daily/%s/prices?token=%s", ticker, viper.GetString("tiingo.token"))
		resp, err := client.
			R().
			SetHeader("Accept", "application/json").
			SetResult(&eodList).
			Get(url)
		if err != nil {
			fmt.Println(err.Error())
		}
		if resp.StatusCode() >= 300 {
			os.Exit(1)
		}
		for _, eod := range eodList {
			dt, _ := time.Parse(time.RFC3339, eod.DateStr)
			fmt.Printf("%s,%s,%s\n", ticker, now.Format(time.RFC3339), dt.Format("2006-01-02"))
		}
	},
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tiingo-monitor.yaml)")
	rootCmd.PersistentFlags().StringP("tiingo-token", "t", "<not-set>", "tiingo API key token")
	viper.BindPFlag("tiingo.token", rootCmd.PersistentFlags().Lookup("tiingo-token"))

	rootCmd.PersistentFlags().String("ticker", "VFIAX", "ticker")
	viper.BindPFlag("tiingo.ticker", rootCmd.PersistentFlags().Lookup("ticker"))

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

		// Search config in home directory with name ".import-tiingo" (without extension).
		viper.AddConfigPath("/etc") // path to look for the config file in
		viper.AddConfigPath(fmt.Sprintf("%s/.config", home))
		viper.AddConfigPath(".")

		viper.SetConfigType("toml")
		viper.SetConfigName("tiingo-monitor")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	viper.ReadInConfig()
}
