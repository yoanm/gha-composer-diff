package sdk

import (
	"log"
	"strings"
)

const annotationEOL = "%0A"

func NoticeAnnotation(title string, txt string, file string) {
	// Replace \n by %0A to ensure multiline messages are correctly displayed in GitHub UI
	log.Printf("::notice file=%s,line=1,title=%s::%s\n", file, title, strings.ReplaceAll(txt, "\n", annotationEOL))
}

func WarningAnnotation(title string, txt string, file string) {
	// Replace \n by %0A to ensure multiline messages are correctly displayed in GitHub UI
	log.Printf("::warning file=%s,line=1,title=%s::%s\n", file, title, strings.ReplaceAll(txt, "\n", annotationEOL))
}
