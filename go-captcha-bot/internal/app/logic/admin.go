package logic

import (
	"bufio"
	"captcha-bot/internal/pkg/conf"
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"
)

type AdminService struct {
	Config *conf.Config
	Bot    *tele.Bot
	poller *tele.LongPoller
}

func NewAdminService(bot *tele.Bot, poller *tele.LongPoller, config *conf.Config) (*AdminService, error) {
	return &AdminService{
		Config: config,
		Bot:    bot,
		poller: poller,
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
	result, err := readLastNLines(service.Config.Logger.LogFile, linesNum)
	if err != nil {
		log.Println("Log file reading error:", err)
		return "", nil
	}

	return result, nil
}

func readLastNLines(filePath string, linesNum int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if len(lines) < linesNum {
		linesNum = len(lines)
	}

	return strings.Join(lines[len(lines)-linesNum:], "\n"), nil
}

func (service *AdminService) GetLastUpdateId() int {
	return service.poller.LastUpdateID
}
