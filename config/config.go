package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type IConfig struct {
	Port      string `json:"port"`
	Mysql     string `json:"mysql"`
	AppId     string `json:"appId"`
	AppSecret string `json:"appSecret"`
}

var Config IConfig

func InitConfig() error {
	// 读取 config.json 文件
	file, err := os.Open("config.json")
	if err != nil {
		fmt.Println("Error opening config file:", err)
		return err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&Config)
	if err != nil {
		fmt.Println("Error decoding config file:", err)
		return err
	}
	return nil
}
