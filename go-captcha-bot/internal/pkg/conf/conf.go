package conf

import (
	"os"
	"time"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v2"
)

// App config
type Config struct {
	Bot struct {
		Token           string        `yaml:"token"`
		BanTimeout      int           `yaml:"ban_timeout" default:"120"`
		CaptchaMsg      string        `yaml:"captcha_message"`
		UserStateTTL    time.Duration `yaml:"user_state_ttl" default:"300"`
		CleanupInterval time.Duration `yaml:"cleanup_interval" default:"120"`
		MinKickVotesFor uint          `yaml:"min_kick_votes_for" default:"3"`
		VoteKickTimeout time.Duration `yaml:"vote_kick_timeout" default:"120"`
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
