package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	DB     DBConfig     `yaml:"database"`
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	c := Config{}
	d := yaml.NewDecoder(f)
	if err = d.Decode(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

type ServerConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	LRUSize int    `yaml:"lru_size"`
}

func (c *ServerConfig) HostAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type DBConfig struct {
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`

	ConnTryTime  int `yaml:"conn_try_time"`
	ConnTriesCnt int `yaml:"conn_tries_cnt"`

	MaxIdleConns int `yaml:"max_idle_conns"`
	MaxOpenConns int `yaml:"max_open_conns"`
}

func (c *DBConfig) ConnectURL() string {
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" + strconv.Itoa(c.Port) + "/" + c.Name
}
