package webhook

import (
	"fmt"
	"strings"

	"github.com/kiry163/claw-mail-monitor/internal/parser"
)

func FormatNotification(mail parser.ParsedEmail) string {
	lines := []string{
		"## \U0001F4E7 新邮件通知",
		"",
		fmt.Sprintf("**发件人：** %s", mail.From),
		fmt.Sprintf("**收件人：** %s", mail.To),
		fmt.Sprintf("**主题：** %s", mail.Subject),
		fmt.Sprintf("**时间：** %s", mail.Date.Format("2006-01-02 15:04:05")),
		"",
		"---",
		"",
		"### 邮件摘要",
		"",
		mail.Summary,
		"",
		"---",
		"*来自 claw-mail-monitor*",
	}

	return strings.Join(lines, "\n")
}
