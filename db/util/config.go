package util

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DataSource          string        `mapstructure:"DATA_SOURCE"`
	DBDriver            string        `mapstructure:"DB_DRIVER"`
	ServerAddress       string        `mapstructure:"SERVER_ADDRESS"`
	TokenSymmetricKey   string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
}

func LoadConfig(path string) (config Config, err error) { //here return variables can be used as local variables
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env") //json,xml

	viper.AutomaticEnv()       //override config values with the environment variable values
	err = viper.ReadInConfig() //read and load config file from disk or key/value  store
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)

	return //naked return

}
