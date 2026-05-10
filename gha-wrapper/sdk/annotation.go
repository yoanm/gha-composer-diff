package sdk

import (
	"log"
	"strings"
)

func NoticeAnnotation(txt string, file string) {
	// Replace \n by %0A to ensure multiline messages are correctly displayed in GitHub UI
	log.Printf("::notice file=%s::%s\n", file, strings.ReplaceAll(txt, "\n", "%0A"))
}

func WarningAnnotation(txt string, file string) {
	// Replace \n by %0A to ensure multiline messages are correctly displayed in GitHub UI
	log.Printf("::warning file=%s::%s\n", file, strings.ReplaceAll(txt, "\n", "%0A"))
}
