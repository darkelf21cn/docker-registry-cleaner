package cmd

import (
	"docker-registry-cleaner/cleaner"
	"docker-registry-cleaner/config"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFile string

var cmd = &cobra.Command{
	Use:   "docker-registry-cleaner",
	Short: "docker-registry-cleaner",
	Run: func(cmd *cobra.Command, args []string) {
		conf := loadConfig(configFile)
		c := cleaner.NewCleaner(conf)
		err := c.Run()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	pf := cmd.PersistentFlags()
	pf.StringVarP(&configFile, "config", "c", "config.yaml", "configuration file")
	viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))
}

func Execute() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadConfig(file string) config.Config {
	conf := config.Config{
		RetentionPolicy: config.RetentionPolicy{
			Default: config.DefaultRetentionPolicy{
				TagsToKeep: 10,
				DaysToKeep: 0,
				KeepLatest: true,
			},
		},
	}

	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("read config failed: %s\n", err.Error())
		os.Exit(1)
	}

	if err := viper.Unmarshal(&conf); err != nil {
		fmt.Printf("unmarshal config failed: %s\n", err.Error())
		os.Exit(1)
	}

	if err := conf.Validate(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return conf
}
