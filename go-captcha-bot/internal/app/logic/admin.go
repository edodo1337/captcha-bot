package logic

import (
	"bufio"
	"captcha-bot/internal/pkg/conf"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"
)

type AdminService struct {
	Config *conf.Config
	Bot    *tele.Bot
}

func NewAdminService(bot *tele.Bot, config *conf.Config) (*AdminService, error) {
	return &AdminService{
		Config: config,
		Bot:    bot,
	}, nil
}

func (service *AdminService) InAdminWhitelist(username string) bool {
	for _, item := range service.Config.Bot.Admins {
		if username == item {
			return true
		}
	}
	return false
}

func (service *AdminService) TailLogs(ctx context.Context, msg tele.Message) (string, error) {
	username := msg.Sender.Username
	if !service.InAdminWhitelist(username) {
		log.Println("TailLog declined: Not in whitelist")
		return "", nil
	}

	linesNum, err := strconv.Atoi(msg.Payload)
	if err != nil {
		linesNum = DEFAULT_LINES_NUM
	}

	cmd := exec.Command("tail", "-n", fmt.Sprintf("%d", linesNum), conf.New().Logger.LogFile)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(stdout)
	var output strings.Builder
	for scanner.Scan() {
		output.WriteString(scanner.Text() + "\n")
	}

	if err := cmd.Wait(); err != nil {
		return "", err
	}

	return output.String(), nil
}
