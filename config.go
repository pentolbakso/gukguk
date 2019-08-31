package main

import (
	"errors"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Entity struct {
	ID   int
	Name string
	Http struct {
		Url string
	}
	Process struct {
		Path string
	}
}

type AppConfig struct {
	LogLevel      string
	CheckInterval int
	Notify        struct {
		Email struct {
			Smtp struct {
				Host     string
				Port     int
				Username string
				Password string
			}
			Sender   string
			Receiver []string
		}
		Telegram struct {
			AccessToken string
			ChannelId   string
		}
	}
	Watch []Entity
}

func (m *AppConfig) Parse(cfgPath string) (*AppConfig, error) {
	// find and read configuration file:
	cfgFile, e := ioutil.ReadFile(cfgPath)
	if e != nil {
		return nil, errors.New("Could not read configuration file! Error: " + e.Error())
	}

	// parse configuration file (YAML synt):
	if e := yaml.UnmarshalStrict(cfgFile, m); e != nil {
		return nil, errors.New("Could not parse configuration file! Error: " + e.Error())
	}

	return m, nil
}
