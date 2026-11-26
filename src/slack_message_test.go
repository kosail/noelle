package main

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSplitSlackMessage(t *testing.T) {
	texts := []string{
		// 1. Inline code split across a boundary: one ` opens, closes in next chunk.
		//    Tests lastOpensInl-style correction (parity within a chunk is odd).
		"Here is some text with `inline code` in it. " +
			strings.Repeat("filler words to push past the limit. ", 20) +
			"and more `inline` here.",

		// 2. Fenced block split across a boundary: ``` opens in chunk 1, closes in chunk 2.
		//    Tests block correction — chunk 1 should get a closing ``` appended,
		//    chunk 2 should get an opening ``` prepended.
		"Intro paragraph.\n```go\n" +
			strings.Repeat("fmt.Println(\"hello world\")\n", 40) +
			"```\nOutro paragraph.",

		// 3. The bug case: a ``` block followed by a single ` in the same chunk.
		//    needsInlCorrection counts 3 (from ```) + 1 = 4 → even → your code thinks
		//    inline code is balanced, but the single ` is actually unclosed.
		//    This is the case that exposes counting ``` as three `s.
		"Start.\n```\ncode block\n```\nthen an unclosed `inline start " +
			strings.Repeat("padding text to force a split. ", 25),
	}

	for i, text := range texts {
		parts := SplitSlackMessage(text, 300) // limit = 50 runes per chunk
		t.Logf("case %d: got %d parts", i, len(parts))
		for j, p := range parts {
			t.Logf("  [%d] %q", j, p)

			if strings.Count(p, "```")%2 != 0 {
				t.Errorf("unbalanced fence in chunk: %q", p)
			}
		}
	}
}

func TestSplitSlackMessageEdgeCases(t *testing.T) {
	t.Run("empty and short messages", func(t *testing.T) {
		for _, text := range []string{"", "hello", strings.Repeat("a", 50)} {
			parts := SplitSlackMessage(text, 300) // limit = 50
			if len(parts) != 1 || parts[0] != text {
				t.Errorf("expected single unchanged chunk for %q, got %q", text, parts)
			}
		}
	})

	t.Run("exact boundary inside fence does not emit trailing empty fence", func(t *testing.T) {
		// Chunk 1 hits limit=50 inside a fence, leaving 46 runes of code + "```\n" (4 runes) = 50 on the last rune.
		text := "```" + strings.Repeat("a", 47) + strings.Repeat("b", 46)
		parts := SplitSlackMessage(text, 300) // limit = 50
		for i, p := range parts {
			if p == "```\n```" || p == "```" {
				t.Errorf("chunk %d is an empty phantom fence: %q", i, p)
			}
			if strings.Count(p, "```")%2 != 0 {
				t.Errorf("chunk %d has unbalanced fence: %q", i, p)
			}
		}
		if len(parts) != 2 {
			t.Errorf("expected 2 chunks, got %d: %q", len(parts), parts)
		}
	})

	t.Run("small or invalid maxRunes does not split per character", func(t *testing.T) {
		text := "hello world"
		if parts := SplitSlackMessage(text, 50); len(parts) != 1 || parts[0] != text {
			t.Errorf("maxRunes=50 should not split 11-rune text, got %q", parts)
		}
		if parts := SplitSlackMessage(text, 0); len(parts) != 1 || parts[0] != text {
			t.Errorf("maxRunes=0 should return single chunk, got %q", parts)
		}
	})

	t.Run("multi-byte UTF-8 runes stay valid and within maxRunes", func(t *testing.T) {
		text := strings.Repeat("🚀", 120)
		parts := SplitSlackMessage(text, 300) // limit = 50
		if len(parts) != 3 {
			t.Fatalf("expected 3 chunks, got %d", len(parts))
		}
		var rebuilt strings.Builder
		for i, p := range parts {
			if !utf8.ValidString(p) {
				t.Errorf("chunk %d is invalid UTF-8", i)
			}
			if utf8.RuneCountInString(p) > 300 {
				t.Errorf("chunk %d exceeded maxRunes: %d", i, utf8.RuneCountInString(p))
			}
			rebuilt.WriteString(p)
		}
		if rebuilt.String() != text {
			t.Errorf("rebuilt text mismatch")
		}
	})
}
