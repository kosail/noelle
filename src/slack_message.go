package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// SplitSlackMessage splits a Slack message into multiple messages if it exceeds the maximum number of runes.
func SplitSlackMessage(text string, maxRunes int) []string {
	limit := maxRunes - 250 // Leave room for msg formatting, etc.
	count := utf8.RuneCountInString(text)

	// If the message is shorter than the limit, then just return it
	if count < limit {
		return []string{text}
	}

	// Split the message into multiple parts.
	// 1. Make sure to break after a paragraph or on last line break.
	// 2. If the break is found to be make inside a block of code, then close the block and reopen it in the new part
	var result = make([]string, count/limit)
	var lastOpensBlk = false

	for i := 0; i < count/limit; i++ {
		block := text[i*limit : i*limit+limit]

		needsCorrection := strings.Count(block, "```")
		if needsCorrection%2 == 0 {
			result[i] = block
			continue
		}

		if lastOpensBlk == false {
			result[i] = fmt.Sprintf("%s```", block)
			lastOpensBlk = true
		} else {
			result[i] = fmt.Sprintf("```%s", block)
			lastOpensBlk = false
		}
	}

	return result
}
