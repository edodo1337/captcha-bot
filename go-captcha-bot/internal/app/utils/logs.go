package utils

import (
	"captcha-bot/internal/pkg/conf"

	"github.com/leemcloughlin/logfile"

	"io"
	"log"
	"os"
)

func InitLogger(config *conf.Config) error {
	logFile, err := logfile.New(
		&logfile.LogFile{
			FileName: config.Logger.LogFile,
			MaxSize:  500 * 1024,
			Flags:    logfile.FileOnly,
		})
	if err != nil {
		log.Fatalf("Failed to create logFile %s: %s\n", config.Logger.LogFile, err)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	log.Println("Logger initialized")
	return nil
}
