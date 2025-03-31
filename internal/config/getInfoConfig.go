package config

import (
	"awesomeProject3/internal/global"
	"github.com/spf13/viper"
	"log"
)

type InfoConfig struct {
	vp *viper.Viper
}

func NewConfig() (*InfoConfig, error) {
	vp := viper.New()
	vp.SetConfigName("config")
	vp.AddConfigPath("config")
	vp.SetConfigType("yaml")
	err := vp.ReadInConfig()
	if err != nil {
		return nil, err
	}
	return &InfoConfig{vp}, nil
}

func SetupConfig() {
	conf, err := NewConfig()
	if err != nil {
		log.Panic("NewConfig error : ", err)
	}
	err = conf.ReadSection("MistralApiKey", &global.MistralApiKeyConfig)
	if err != nil {
		log.Panic("ReadSection - MistralApiKey error : ", err)
	}
	err = conf.ReadSection("BlockChain", &global.BlockChainConfig)
	if err != nil {
		log.Panic("ReadSection - BlockChain error : ", err)
	}
}

func (config *InfoConfig) ReadSection(k string, v interface{}) error {
	err := config.vp.UnmarshalKey(k, v)
	if err != nil {
		return err
	}
	return nil
}
