package config

import (
	"errors"
	"os"

	"github.com/spf13/viper"
)

//Load config file
func Load(configFile string) error {
	viper.SetConfigType("yaml")
	viper.BindEnv("J2P_USERNAME")
	viper.BindEnv("J2P_PASSWORD")
	viper.SetDefault("datetime_format", "2006-01-02 15:04:05")
	viper.SetDefault("api_datetime_format", "2006-01-02T15:04:05.999999999-0700")
	viper.SetDefault("query_page_size", 4000)
	viper.SetDefault("issues_per_pdf", 2000)

	if len(viper.GetString("J2P_USERNAME")) == 0 {
		return errors.New("environment variable J2P_USERNAME not set")
	}

	if len(viper.GetString("J2P_PASSWORD")) == 0 {
		return errors.New("environment variable J2P_PASSWORD not set")
	}

	file, err := os.Open(configFile)
	if err != nil {
		return err
	}

	err = viper.ReadConfig(file)
	if err != nil {
		return err
	}

	return nil
}
