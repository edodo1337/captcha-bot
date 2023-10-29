package logic

import "strings"

func CaptchaMessage(len int, currentPos, correctPos int) string {
	var sb strings.Builder

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
