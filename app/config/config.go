package config

import (
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config contains path to configuration file and allows to read it.
type Config struct {
	File string
}

// Read configuration.
func (c *Config) Read() error {
	if c.File != "" {
		// Use config file from the flag.
		viper.SetConfigFile(c.File)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			return err
		}

		// Search config in home directory with name ".c" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".pmm-agent")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	return viper.ReadInConfig()
}

// Flags assigns flags to cmd.
func (c *Config) Flags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&c.File, "config", "", "config File (default is $HOME/.pmm-agent.yaml)")
}
