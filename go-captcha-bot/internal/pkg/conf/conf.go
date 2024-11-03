package conf

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v2"
)

// App config
type Config struct {
	Bot struct {
		Token             string        `yaml:"token"`
		BanTimeout        int           `yaml:"ban_timeout" default:"120"`
		CaptchaMsg        string        `yaml:"captcha_message"`
		UserStateTTL      time.Duration `yaml:"user_state_ttl" default:"300"`
		CleanupInterval   time.Duration `yaml:"cleanup_interval" default:"120"`
		MinKickVotesFor   uint          `yaml:"min_kick_votes_for" default:"3"`
		VoteKickTimeout   time.Duration `yaml:"vote_kick_timeout" default:"120"`
		GeminiApiTokens   []string      `yaml:"gemini_api_tokens"`
		GeminiModel       string        `yaml:"gemini_model" default:"gemini-pro"`
		PromptWrap        string        `yaml:"prompt_wrap"`
		Admins            []string      `yaml:"admins"`
		MsgCountThreshold int           `yaml:"msg_count_threshold" default:"5"`
	} `yaml:"bot"`
	Redis struct {
		Url      string `yaml:"url"`
		Password string `yaml:"password"`
		Db       int    `yaml:"db" default:"0"`
	} `yaml:"redis"`
	Logger struct {
		LogFile string `yaml:"log_file" default:"bot.log"`
	}
}

func (c *Config) Validate() error {
	if c.Redis.Url == "" {
		return errors.New("redis_url should not be empty")
	}
	if len(c.Bot.GeminiApiTokens) == 0 {
		return errors.New("gemini_api_tokens should not be empty")
	}
	return nil
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

	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("App configuration error: %s", err))
	}

	return &cfg
}

func (c *Config) BotToken() string {
	return c.Bot.Token
}
