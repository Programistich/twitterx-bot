package telegraph

import (
	"strings"
	"testing"
)

func TestValidateTitle_Valid(t *testing.T) {
	c := NewConverter()

	title, err := c.ValidateTitle("Tweet by @elonmusk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if title != "Tweet by @elonmusk" {
		t.Fatalf("expected 'Tweet by @elonmusk', got '%s'", title)
	}
}

func TestValidateTitle_Empty(t *testing.T) {
	c := NewConverter()

	_, err := c.ValidateTitle("")
	if err != ErrTitleEmpty {
		t.Fatalf("expected ErrTitleEmpty, got %v", err)
	}
}

func TestValidateTitle_TooLong(t *testing.T) {
	c := NewConverter()

	longTitle := strings.Repeat("a", 300) // Більше 256
	_, err := c.ValidateTitle(longTitle)
	if err != ErrTitleTooLong {
		t.Fatalf("expected ErrTitleTooLong, got %v", err)
	}
}

func TestValidateTitle_Trimmed(t *testing.T) {
	c := NewConverter()

	title, err := c.ValidateTitle("  Tweet by @user  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if title != "Tweet by @user" {
		t.Fatalf("expected trimmed title, got '%s'", title)
	}
}

func TestTextToNodes_SimpleParagraph(t *testing.T) {
	c := NewConverter()

	nodes, err := c.TextToNodes("Hello, World!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	// Перевіряємо що це параграф з текстом
	node, ok := nodes[0].(map[string]any)
	if !ok {
		t.Fatal("expected map[string]any")
	}
	if node["tag"] != "p" {
		t.Fatalf("expected tag 'p', got '%v'", node["tag"])
	}
}

func TestTextToNodes_MultipleParagraphs(t *testing.T) {
	c := NewConverter()

	nodes, err := c.TextToNodes("First paragraph.\n\nSecond paragraph.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
}

func TestTextToNodes_EmptyContent(t *testing.T) {
	c := NewConverter()

	_, err := c.TextToNodes("")
	if err != ErrContentEmpty {
		t.Fatalf("expected ErrContentEmpty, got %v", err)
	}
}

func TestTextToNodes_ContentTooLong(t *testing.T) {
	c := NewConverter()

	longContent := strings.Repeat("a", MaxContentLength+1)
	_, err := c.TextToNodes(longContent)
	if err != ErrContentTooLong {
		t.Fatalf("expected ErrContentTooLong, got %v", err)
	}
}

func TestTextToNodes_WithURL(t *testing.T) {
	c := NewConverter()

	nodes, err := c.TextToNodes("Check this: https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Має бути 1 параграф
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	// Перевіряємо що URL перетворено в посилання
	node := nodes[0].(map[string]any)
	children := node["children"].([]any)

	// Шукаємо посилання серед children
	foundLink := false
	for _, child := range children {
		if childMap, ok := child.(map[string]any); ok {
			if childMap["tag"] == "a" {
				foundLink = true
				break
			}
		}
	}

	if !foundLink {
		t.Fatal("expected to find link in children")
	}
}

func TestTextToNodes_WithLineBreak(t *testing.T) {
	c := NewConverter()

	nodes, err := c.TextToNodes("Line 1\nLine 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 paragraph, got %d", len(nodes))
	}

	// Перевіряємо що є <br> тег
	node := nodes[0].(map[string]any)
	children := node["children"].([]any)

	foundBr := false
	for _, child := range children {
		if childMap, ok := child.(map[string]any); ok {
			if childMap["tag"] == "br" {
				foundBr = true
				break
			}
		}
	}

	if !foundBr {
		t.Fatal("expected to find <br> in children")
	}
}
