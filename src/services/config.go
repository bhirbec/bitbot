package services

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Mysql struct {
		Db   string `json:"db"`
		Host string `json:"host"`
		Port string `json:"port"`
		User string `json:"user"`
		Pwd  string `json:"pwd"`
	} `json:"mysql"`
	Bittrex struct {
		Key    string `json:"key"`
		Secret string `json:"secret"`
	} `json:"bittrex"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	dest := &Config{}
	err = json.Unmarshal(data, dest)
	return dest, err
}
