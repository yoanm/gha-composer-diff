package ghasdk

import (
	"log"
	"strconv"
	"strings"
)

const annotationEOL = "%0A"

func NoticeAnnotation(body string, file string, title string, line int) {
	log.Print(buildAnnotation("notice", title, body, file, line))
}

func WarningAnnotation(body string, file string, title string, line int) {
	log.Print(buildAnnotation("warning", title, body, file, line))
}

func buildAnnotation(annoHeader string, title string, body string, file string, line int) string {
	// Replace \n by %0A to ensure multiline messages are correctly displayed in GitHub UI
	builder := strings.Builder{}

	parts := []string{}
	if len(file) > 0 {
		parts = append(parts, "file="+file)
	}

	if line > 0 {
		parts = append(parts, "line="+strconv.Itoa(line))
	}

	if len(title) > 0 {
		parts = append(parts, "title="+strings.ReplaceAll(title, "\n", annotationEOL))
	}

	builder.WriteString("::" + annoHeader)

	if len(parts) > 0 {
		builder.WriteString(" " + strings.Join(parts, ","))
	}

	builder.WriteString("::" + strings.ReplaceAll(body, "\n", annotationEOL))

	return builder.String()
}
