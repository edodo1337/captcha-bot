package conf

import (
	"os"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v2"
)

// App config
type Config struct {
	Bot struct {
		Token      string `yaml:"token"`
		BanTimeout int    `yaml:"ban_timeout" default:"120"`
	} `yaml:"bot"`
}

func New() *Config {
	f, err := os.Open("./config.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var cfg Config

	decoder := yaml.NewDecoder(f)

	err = decoder.Decode(&cfg)
	if err != nil {
		panic(err)
	}

	if err := defaults.Set(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}

func (c *Config) BotToken() string {
	return c.Bot.Token
}
