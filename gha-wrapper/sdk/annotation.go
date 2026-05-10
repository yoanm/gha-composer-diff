package sdk

import (
	"log"
	"strconv"
	"strings"
)

const annotationEOL = "%0A"

func NoticeAnnotation(txt string, file string, title string, line int) {
	log.Print(buildAnnotation("notice", title, txt, file, line))
}

func WarningAnnotation(title string, txt string, file string, line int) {
	log.Print(buildAnnotation("warning", title, txt, file, line))
}

func buildAnnotation(annoHeader string, title string, txt string, file string, line int) string {
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
	builder.WriteString("::" + strings.ReplaceAll(txt, "\n", annotationEOL))

	return builder.String()
}
