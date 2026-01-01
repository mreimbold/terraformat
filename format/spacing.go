package format

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func newlineToken() *hclwrite.Token {
	return &hclwrite.Token{
		Type:  hclsyntax.TokenNewline,
		Bytes: []byte{'\n'},
		//nolint:revive // add-constant: zero spaces before newlines.
		SpacesBefore: 0,
	}
}

func normalizePrefixTokens(tokens hclwrite.Tokens) hclwrite.Tokens {
	//nolint:revive // add-constant: len check is clear here.
	if len(tokens) == 0 {
		return nil
	}

	firstComment := indexNotFound

	for tokenIndex, token := range tokens {
		if token.Type == hclsyntax.TokenComment {
			firstComment = tokenIndex

			break
		}
	}

	if firstComment == indexNotFound {
		return nil
	}

	kept := tokens[firstComment:]

	//nolint:revive // add-constant: explicit zero max newlines.
	return trimTrailingNewlines(kept, 0)
}

func normalizeLeadingTokens(tokens hclwrite.Tokens) hclwrite.Tokens {
	//nolint:revive // add-constant: len check is clear here.
	if len(tokens) == 0 {
		return nil
	}

	if containsComment(tokens) {
		//nolint:revive // add-constant: explicit zero max newlines.
		return trimTrailingNewlines(tokens, 0)
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
	//nolint:revive // add-constant: len check is clear here.
	if len(tokens) == 0 {
		return nil
	}

	last := len(tokens) - indexOffset

	for last >= indexFirst && tokens[last].Type == hclsyntax.TokenNewline {
		last--
	}

	if last == len(tokens)-indexOffset {
		return tokens
	}

	newEnd := last + indexOffset + maxNewlines
	newEnd = min(newEnd, len(tokens))

	return tokens[:newEnd]
}
