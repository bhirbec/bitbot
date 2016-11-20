package main

import (
	"encoding/json"
	"io/ioutil"
)

type credential struct {
	Key    string
	Secret string
}

func LoadConfig(path string) (map[string]credential, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	dest := map[string]credential{}
	err = json.Unmarshal(data, &dest)
	return dest, err
}
