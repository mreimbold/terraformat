package format

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func newlineToken() *hclwrite.Token {
	return &hclwrite.Token{
		Type:         hclsyntax.TokenNewline,
		Bytes:        []byte{'\n'},
		SpacesBefore: zero,
	}
}

func normalizePrefixTokens(tokens hclwrite.Tokens) hclwrite.Tokens {
	if len(tokens) == zero {
		return nil
	}

	firstComment := negativeOne

	for tokenIndex, token := range tokens {
		if token.Type == hclsyntax.TokenComment {
			firstComment = tokenIndex

			break
		}
	}

	if firstComment == negativeOne {
		return nil
	}

	kept := tokens[firstComment:]

	return trimTrailingNewlines(kept, zero)
}

func normalizeLeadingTokens(tokens hclwrite.Tokens) hclwrite.Tokens {
	if len(tokens) == zero {
		return nil
	}

	if containsComment(tokens) {
		return trimTrailingNewlines(tokens, zero)
	}

	if containsNewline(tokens) {
		return hclwrite.Tokens{newlineToken()}
	}

	return nil
}

func containsNewline(tokens hclwrite.Tokens) bool {
	for _, token := range tokens {
		if token.Type == hclsyntax.TokenNewline {
			return true
		}
	}

	return false
}

func containsComment(tokens hclwrite.Tokens) bool {
	for _, token := range tokens {
		if token.Type == hclsyntax.TokenComment {
			return true
		}
	}

	return false
}

func trimTrailingNewlines(
	tokens hclwrite.Tokens,
	maxNewlines int,
) hclwrite.Tokens {
	if len(tokens) == zero {
		return nil
	}

	last := len(tokens) - one

	for last >= zero && tokens[last].Type == hclsyntax.TokenNewline {
		last--
	}

	if last == len(tokens)-one {
		return tokens
	}

	newEnd := last + one + maxNewlines
	newEnd = min(newEnd, len(tokens))

	return tokens[:newEnd]
}
