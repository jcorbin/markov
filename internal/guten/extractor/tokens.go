package extractor

import (
	"unicode"
	"unicode/utf8"
)

// ScanTokens implements a bufio.SplitFunc that emits saner document tokens;
// punctuation splits as its own token, rather than at the start/end of a word.
func ScanTokens(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip leading spaces.
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !unicode.IsSpace(r) {
			break
		}
	}

	i := start
	r, width := utf8.DecodeRune(data[i:])
	i += width

	if unicode.IsPunct(r) {
		// scan punct token
		for ; i < len(data); i += width {
			r, width = utf8.DecodeRune(data[i:])
			if !unicode.IsPunct(r) {
				return i, data[start:i], nil
			}
		}
	} else {
		// scan non-punct token
		for ; i < len(data); i += width {
			r, width = utf8.DecodeRune(data[i:])
			if unicode.IsSpace(r) {
				return i + width, data[start:i], nil
			}
			if unicode.IsPunct(r) {
				switch r {
				case '-', '\'', '"', '`', '‘', '’':
				default:
					return i, data[start:i], nil
				}
			}
		}
	}

	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}

	// Request more data.
	return start, nil, nil
}
