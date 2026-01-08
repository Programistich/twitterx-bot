package telegraph

import (
	"regexp"
	"strings"
)

const (
	MaxContentLength = 64 * 1024 // 64KB
	MaxTitleLength   = 256
)

var urlRegex = regexp.MustCompile(`https?://[^\s]+`)

// Converter конвертує текст в Telegraph DOM
type Converter struct{}

// NewConverter створює новий конвертер
func NewConverter() *Converter {
	return &Converter{}
}

// ValidateTitle валідує та очищує заголовок
func (c *Converter) ValidateTitle(title string) (string, error) {
	trimmed := strings.TrimSpace(title)

	if trimmed == "" {
		return "", ErrTitleEmpty
	}

	if len(trimmed) > MaxTitleLength {
		return "", ErrTitleTooLong
	}

	return trimmed, nil
}

// TextToNodes конвертує текст в список Telegraph DOM вузлів
func (c *Converter) TextToNodes(text string) ([]any, error) {
	// Валідація
	if strings.TrimSpace(text) == "" {
		return nil, ErrContentEmpty
	}

	if len(text) > MaxContentLength {
		return nil, ErrContentTooLong
	}

	// Розбиваємо на параграфи (по подвійному переносу)
	paragraphs := strings.Split(text, "\n\n")

	var nodes []any
	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		children := c.processParagraph(para)
		node := map[string]any{
			"tag":      "p",
			"children": children,
		}
		nodes = append(nodes, node)
	}

	if len(nodes) == 0 {
		return nil, ErrContentEmpty
	}

	return nodes, nil
}

// processParagraph обробляє один параграф
func (c *Converter) processParagraph(text string) []any {
	var result []any

	// Розбиваємо на рядки (по одинарному переносу)
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Обробляємо контент рядка (URL -> посилання)
			result = append(result, c.processLine(line)...)
		}

		// Додаємо <br> між рядками (не після останнього)
		if i < len(lines)-1 {
			result = append(result, map[string]any{"tag": "br"})
		}
	}

	return result
}

// processLine обробляє один рядок, знаходить URL
func (c *Converter) processLine(line string) []any {
	var result []any

	matches := urlRegex.FindAllStringIndex(line, -1)
	if len(matches) == 0 {
		// Немає URL, просто текст
		return []any{line}
	}

	lastIndex := 0
	for _, match := range matches {
		// Текст перед URL
		if match[0] > lastIndex {
			textBefore := line[lastIndex:match[0]]
			if textBefore != "" {
				result = append(result, textBefore)
			}
		}

		// URL як посилання
		url := line[match[0]:match[1]]
		link := map[string]any{
			"tag":      "a",
			"attrs":    map[string]string{"href": url},
			"children": []any{url},
		}
		result = append(result, link)

		lastIndex = match[1]
	}

	// Текст після останнього URL
	if lastIndex < len(line) {
		textAfter := line[lastIndex:]
		if textAfter != "" {
			result = append(result, textAfter)
		}
	}

	return result
}
