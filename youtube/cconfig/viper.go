package cconfig

import (
	"github.com/spf13/viper"
)

func init() {
	var err error
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".") // local
	viper.AddConfigPath("$HOME/host/WatchForYou/WatchForYou_Crawlers/youtube/cconfig") // TODO: delete. dev
	viper.AddConfigPath("/config/") // for future Dockerfile
	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {
		panic("Failed to initialize config storage... And it's a file!!! What?!")
	}
	
	// #default -ing viper TODO: finish defaults where can
	viper.SetDefault("rmqconn", "amqp://watchforyou:watchforyou@localhost:5672/")
	viper.SetDefault("rmqqname", "crawlpc")
}

type ConfigLoader struct {}

func (ConfigLoader) GetString(query string) string {
	return viper.GetString(query)
}

func (ConfigLoader) GetInt(query string) int {
	return viper.GetInt(query)
}

func (ConfigLoader) GetBool(query string) bool {
	return viper.GetBool(query)
}