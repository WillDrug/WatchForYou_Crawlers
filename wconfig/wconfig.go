package wconfig

import (
	"github.com/spf13/viper"
	"fmt"
	"log"
)

func init() {
	var err error
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".") // local
	viper.AddConfigPath("$HOME/host/Watch4You/WatchForYou_Crawlers/wconfig/") // TODO: delete. dev
	viper.AddConfigPath("/config/") // for future Dockerfile
	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize config storage: %#v", err))
	}
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

func (ConfigLoader) SetDefaults(parms map[string]interface{}) error {
	for k, v := range parms {
		log.Printf("Setting default for %#v", k)
		switch v.(type) {
			case string: viper.SetDefault(k, v)
			case int: viper.SetDefault(k, v)
			case bool: viper.SetDefault(k, v)
			default: log.Printf("Ommitting %#v, unknown type: %T", k, v)
		}
	}
	return nil // TODO: errors?
}