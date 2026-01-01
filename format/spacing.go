package format

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func newlineToken() *hclwrite.Token {
	return &hclwrite.Token{
		Type:  hclsyntax.TokenNewline,
		Bytes: []byte{'\n'},
	}
}

func normalizePrefixTokens(tokens hclwrite.Tokens) hclwrite.Tokens {
	if len(tokens) == 0 {
		return nil
	}

	firstComment := -1
	for i, tok := range tokens {
		if tok.Type == hclsyntax.TokenComment {
			firstComment = i
			break
		}
	}
	if firstComment == -1 {
		return nil
	}

	kept := tokens[firstComment:]
	return trimTrailingNewlines(kept, 0)
}

func normalizeLeadingTokens(tokens hclwrite.Tokens) hclwrite.Tokens {
	if len(tokens) == 0 {
		return nil
	}
	if containsComment(tokens) {
		return trimTrailingNewlines(tokens, 0)
	}
	if containsNewline(tokens) {
		return hclwrite.Tokens{newlineToken()}
	}
	return nil
}

func containsNewline(tokens hclwrite.Tokens) bool {
	for _, tok := range tokens {
		if tok.Type == hclsyntax.TokenNewline {
			return true
		}
	}
	return false
}

func containsComment(tokens hclwrite.Tokens) bool {
	for _, tok := range tokens {
		if tok.Type == hclsyntax.TokenComment {
			return true
		}
	}
	return false
}

func trimTrailingNewlines(tokens hclwrite.Tokens, max int) hclwrite.Tokens {
	if len(tokens) == 0 {
		return nil
	}

	last := len(tokens) - 1
	for last >= 0 && tokens[last].Type == hclsyntax.TokenNewline {
		last--
	}
	if last == len(tokens)-1 {
		return tokens
	}

	newEnd := last + 1 + max
	if newEnd > len(tokens) {
		newEnd = len(tokens)
	}
	return tokens[:newEnd]
}
