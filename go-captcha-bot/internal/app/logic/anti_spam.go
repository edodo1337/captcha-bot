package logic

import (
	"captcha-bot/internal/pkg/conf"
	"context"
	"log"
)

type SpamFilterClient interface {
	// Init(ctx context.Context, cofig *conf.Config) error
	Shutdown() error
	IsSpam(ctx context.Context, text string) (bool, error)
}

type SpamFilterService struct {
	SpamFilterClient SpamFilterClient
	Config           *conf.Config
}

func NewSpamFilterService(sf SpamFilterClient, config *conf.Config) (*SpamFilterService, error) {
	return &SpamFilterService{
		SpamFilterClient: sf,
		Config:           config,
	}, nil
}

func (sfs *SpamFilterService) CheckMessage(ctx context.Context, text string) bool {
	isSpam, err := sfs.SpamFilterClient.IsSpam(ctx, text)
	if err != nil {
		log.Fatal(err.Error())
		return false
	}

	return isSpam
}
