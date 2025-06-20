package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	homedir "github.com/mitchellh/go-homedir"

)
var (
	cfgFile	  string
	userLicense string

	rootCmd = &cobra.Command{
		Use:   "backups",
		Short: "Backups service",
		Long:  `Backups service for managing and processing database backups.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Welcome to the Backups service!")
			fmt.Println("Use 'backups help' for more information.")
		},
	}
)

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("command execution error: %w", err)
	}
	return nil
}


func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "c", "Path to configuration file")
	rootCmd.PersistentFlags().StringP("author", "a", "d00p1", "Author of the service")
	rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "License of the service")
	rootCmd.PersistentFlags().Bool("viper", true, "Use Viper for configuration management")
	viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	viper.BindPFlag("license", rootCmd.PersistentFlags().Lookup("license"))
	viper.BindPFlag("viper", rootCmd.PersistentFlags().Lookup("viper"))
}

func initConfig() {
 	// Don't forget to read config either from cfgFile or from home directory!
 	if cfgFile != "" {
    	// Use config file from the flag.
    	viper.SetConfigFile(cfgFile)
  	} else {
  	  	// Find home directory.
  	  	home, err := homedir.Dir()
  	  	if err != nil {
  	  		fmt.Println("Home directory not found: %w", err)
			os.Exit(2)
  		}

    	// Search config in home directory with name ".cobra" (without extension).
    	viper.AddConfigPath(home)
    	viper.SetConfigName(".cobra")
  	}

  	if err := viper.ReadInConfig(); err != nil {
  	  	fmt.Println("Can't read config:", err)
		os.Exit(2)
  	}
}