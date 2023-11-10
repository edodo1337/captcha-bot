package logic

import (
	"fmt"
	"strings"
)

func CaptchaMessage(len, currentPos, correctPos int, message string, banTimeout int) string {
	var sb strings.Builder

	sb.WriteString(captchaTimeoutText(banTimeout))
	sb.WriteString("Помоги жабе съесть яблоко:\n")

	for i := 0; i < len; i++ {
		if i == currentPos {
			sb.WriteString("🐸")
		} else if i == correctPos {
			sb.WriteString("🍎")
		} else {
			sb.WriteString("🟡")
		}
	}

	return sb.String()
}

func captchaTimeoutText(banTimeout int) string {
	var sb strings.Builder
	minutes := banTimeout / 60
	seconds := banTimeout % 60

	sb.WriteString("У вас ")

	if minutes != 0 {
		minTxt := getMinuteWord(minutes)
		sb.WriteString(fmt.Sprintf("%d %s ", minutes, minTxt))
	}
	if seconds != 0 {
		secondsTxt := getSecondWord(seconds)
		sb.WriteString(fmt.Sprintf("%d %s ", seconds, secondsTxt))
	}

	sb.WriteString("на решение капчи.\n")

	return sb.String()
}

func getMinuteWord(minutes int) string {
	if minutes%10 == 1 && minutes%100 != 11 {
		return "минута"
	}
	if minutes%10 >= 2 && minutes%10 <= 4 && (minutes%100 < 10 || minutes%100 >= 20) {
		return "минуты"
	}
	return "минут"
}

func getSecondWord(seconds int) string {
	if seconds%10 == 1 && seconds%100 != 11 {
		return "секунда"
	}
	if seconds%10 >= 2 && seconds%10 <= 4 && (seconds%100 < 10 || seconds%100 >= 20) {
		return "секунды"
	}
	return "секунд"
}
