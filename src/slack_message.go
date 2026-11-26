package main

import (
	"strings"
	"unicode/utf8"
)

// SplitSlackMessage splits a Slack message into multiple messages if it exceeds
// the maximum number of runes, keeping ``` fences balanced across splits.
func SplitSlackMessage(text string, maxRunes int) []string {
	limit := maxRunes - 250 // Leave room for msg formatting, etc.
	if limit <= 0 {
		limit = maxRunes
	}
	if limit <= 0 || utf8.RuneCountInString(text) <= limit {
		return []string{text}
	}

	var (
		result  []string
		cur     strings.Builder // current chunk being built
		curLen  int             // runes in cur
		inFence bool            // are we inside a ``` block?
	)

	flush := func() {
		if inFence {
			cur.WriteString("```") // close the fence in this chunk
		}
		result = append(result, cur.String())
		cur.Reset()
		curLen = 0
		if inFence {
			cur.WriteString("```\n") // reopen in the next chunk
			curLen = 4
		}
	}

	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		// Detect a ``` fence and toggle state. Copy all three backticks.
		if runes[i] == '`' && i+2 < len(runes) && runes[i+1] == '`' && runes[i+2] == '`' {
			cur.WriteString("```")
			curLen += 3
			inFence = !inFence
			i += 2
		} else {
			cur.WriteRune(runes[i])
			curLen++
		}

		if curLen >= limit && i < len(runes)-1 {
			flush()
		}
	}

	if curLen > 0 {
		if inFence {
			cur.WriteString("```")
		}
		result = append(result, cur.String())
	}

	return result
}
