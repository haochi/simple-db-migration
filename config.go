package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type ConfigJson struct {
	Type   string       `json:"type"`
	Deltas string       `json:"deltas"`
	Sqlite SqliteConfig `json:"sqlite"`
}

type SqliteConfig struct {
	File string `json:"file"`
}

func parseConfigJson(fileName string) (*ConfigJson, error) {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var configJson ConfigJson
	err = json.Unmarshal(byteValue, &configJson)
	if err != nil {
		return nil, err
	}

	return &configJson, nil
}
