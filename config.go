package main

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	DB DBConfig
	WX WXConfig
}

type WXConfig struct {
	URL       string
	AppId     string
	AppSecret string
}
type DBConfig struct {
	Path string
}

func GetConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Warn().Err(err).Msg("Failed to read configuration from disk")

		return nil,fmt.Errorf("fatal error reading config file: %w", err)
	}

	wxUrl := viper.GetString("weixin.url")
	wxAppId := viper.GetString("weixin.app_id")
	wxAppKey := viper.GetString("weixin.app_secret")

	dbPath := viper.GetString("db.path")

	config := Config{
		DB: DBConfig{
			Path: dbPath,
		},
		WX: WXConfig{
			URL:       wxUrl,
			AppId:     wxAppId,
			AppSecret: wxAppKey,
		},
	}
	return &config, nil
}
