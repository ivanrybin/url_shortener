package config

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

func Test_ConfigDecoder(t *testing.T) {
	dbCfg := DBConfig{
		Name:     "db",
		Host:     "host",
		Port:     5432,
		User:     "user",
		Password: "password",
	}
	serverCfg := ServerConfig{
		Port:    12345,
		Host:    "host",
		LRUSize: 10000,
	}
	cfg := Config{
		DB:     dbCfg,
		Server: serverCfg,
	}

	buf := &bytes.Buffer{}
	e := yaml.NewEncoder(buf)
	assert.Nil(t, e.Encode(cfg))

	d := yaml.NewDecoder(buf)
	cfgDecoded := &Config{}
	assert.Nil(t, d.Decode(cfgDecoded))

	assert.Equal(t, cfg, *cfgDecoded, "configs not equal")
}
