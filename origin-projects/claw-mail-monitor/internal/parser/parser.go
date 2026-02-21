package parser

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	html2markdown "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-message/mail"
)

type ParsedEmail struct {
	From    string
	To      string
	Subject string
	Date    time.Time
	Body    string
	Summary string
	UID     uint32
}

func ParseMessage(msg *imap.Message, body io.Reader) (ParsedEmail, error) {
	parsed := ParsedEmail{}
	if msg != nil {
		parsed.UID = msg.Uid
		if msg.Envelope != nil {
			parsed.Subject = msg.Envelope.Subject
			parsed.Date = msg.Envelope.Date
			parsed.From = formatAddresses(msg.Envelope.From)
			parsed.To = formatAddresses(msg.Envelope.To)
		}
	}

	if body == nil {
		return parsed, nil
	}

	raw, err := io.ReadAll(body)
	if err != nil {
		return parsed, fmt.Errorf("read body failed: %w", err)
	}

	reader := bytes.NewReader(raw)
	mr, err := mail.CreateReader(reader)
	if err != nil {
		parsed.Body = strings.TrimSpace(string(raw))
		parsed.Summary = summarize(parsed.Body)
		return parsed, nil
	}

	var plainBody string
	var htmlBody string
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return parsed, fmt.Errorf("read multipart failed: %w", err)
		}

		header, ok := part.Header.(*mail.InlineHeader)
		if !ok {
			continue
		}
		partType, _, _ := header.ContentType()
		content, _ := io.ReadAll(part.Body)
		text := strings.TrimSpace(string(content))
		switch {
		case strings.Contains(partType, "text/plain"):
			if plainBody == "" {
				plainBody = text
			}
		case strings.Contains(partType, "text/html"):
			if htmlBody == "" {
				htmlBody = text
			}
		}
	}

	if plainBody != "" {
		parsed.Body = plainBody
		parsed.Summary = summarize(plainBody)
		return parsed, nil
	}
	if htmlBody != "" {
		parsed.Body = convertHTML(htmlBody)
		parsed.Summary = summarize(parsed.Body)
		return parsed, nil
	}

	parsed.Body = strings.TrimSpace(string(raw))
	parsed.Summary = summarize(parsed.Body)
	return parsed, nil
}

func formatAddresses(list []*imap.Address) string {
	if len(list) == 0 {
		return ""
	}

	parts := make([]string, 0, len(list))
	for _, addr := range list {
		if addr == nil {
			continue
		}
		email := addr.MailboxName + "@" + addr.HostName
		if addr.PersonalName != "" {
			parts = append(parts, fmt.Sprintf("%s <%s>", addr.PersonalName, email))
		} else {
			parts = append(parts, email)
		}
	}

	return strings.Join(parts, ", ")
}

func convertHTML(content string) string {
	converter := html2markdown.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(content)
	if err != nil {
		return strings.TrimSpace(content)
	}
	return strings.TrimSpace(markdown)
}

func summarize(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return "(空内容)"
	}

	const limit = 300
	runes := []rune(content)
	if len(runes) <= limit {
		return content
	}

	return string(runes[:limit]) + "..."
}
