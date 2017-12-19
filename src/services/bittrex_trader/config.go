package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

func loadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	dest := &Config{}
	err = json.Unmarshal(data, dest)
	return dest, err
}
