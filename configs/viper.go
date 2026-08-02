package configs

import (
	"fmt"

	"github.com/spf13/viper"
)

func NewViper(filename, filetype string, paths ...string) (*viper.Viper, error) {
	config := viper.New()

	config.AutomaticEnv()

	config.SetConfigName(filename)
	config.SetConfigType(filetype)

	for _, p := range paths {
		config.AddConfigPath(p)
	}

	err := config.ReadInConfig()
	if err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if !ok {
			return nil, fmt.Errorf("error reading config viper file: %w", err)
		}
	}

	return config, nil
}
