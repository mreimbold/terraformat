// Package spacing normalizes whitespace between formatter items.
package spacing

import (
	"bytes"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/mreimbold/terraformat/internal/format/model"
)

const (
	maxNewlinesNone   = 0
	maxNewlinesSingle = 1
)

// NewlineToken returns a newline token.
func NewlineToken() *hclwrite.Token {
	return &hclwrite.Token{
		Type:  hclsyntax.TokenNewline,
		Bytes: []byte{'\n'},
		//nolint:revive // add-constant: zero spaces before newlines.
		SpacesBefore: 0,
	}
}

// NormalizePrefixTokens trims prefixes and preserves trailing comment spacing.
func NormalizePrefixTokens(tokens hclwrite.Tokens) hclwrite.Tokens {
	//nolint:revive // add-constant: len check is clear here.
	if len(tokens) == 0 {
		return nil
	}

	firstComment := model.IndexNotFound

	for tokenIndex, token := range tokens {
		if token.Type == hclsyntax.TokenComment {
			firstComment = tokenIndex

			break
		}
	}

	if firstComment == model.IndexNotFound {
		return nil
	}

	kept := tokens[firstComment:]

	maxNewlines := trailingNewlinesAfterComment(kept)

	return trimTrailingNewlines(kept, maxNewlines)
}

// NormalizeLeadingTokens normalizes leading tokens to a single newline.
func NormalizeLeadingTokens(tokens hclwrite.Tokens) hclwrite.Tokens {
	//nolint:revive // add-constant: len check is clear here.
	if len(tokens) == 0 {
		return nil
	}

	if ContainsComment(tokens) {
		maxNewlines := trailingNewlinesAfterComment(tokens)

		return trimTrailingNewlines(tokens, maxNewlines)
	}

	if containsNewline(tokens) {
		return hclwrite.Tokens{NewlineToken()}
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

// ContainsComment reports whether tokens include any comments.
func ContainsComment(tokens hclwrite.Tokens) bool {
	for _, token := range tokens {
		if token.Type == hclsyntax.TokenComment {
			return true
		}
	}

	return false
}

func trailingNewlinesAfterComment(tokens hclwrite.Tokens) int {
	lastComment := model.IndexNotFound

	startIndex := len(tokens) - model.IndexOffset
	for tokenIndex := startIndex; tokenIndex >= model.IndexFirst; tokenIndex-- {
		if tokens[tokenIndex].Type == hclsyntax.TokenComment {
			lastComment = tokenIndex

			break
		}
	}

	if lastComment == model.IndexNotFound {
		return maxNewlinesNone
	}

	if bytes.HasSuffix(tokens[lastComment].Bytes, []byte("\n")) {
		return maxNewlinesNone
	}

	return maxNewlinesSingle
}

func trimTrailingNewlines(
	tokens hclwrite.Tokens,
	maxNewlines int,
) hclwrite.Tokens {
	//nolint:revive // add-constant: len check is clear here.
	if len(tokens) == 0 {
		return nil
	}

	last := len(tokens) - model.IndexOffset

	for last >= model.IndexFirst &&
		tokens[last].Type == hclsyntax.TokenNewline {
		last--
	}

	if last == len(tokens)-model.IndexOffset {
		return tokens
	}

	newEnd := last + model.IndexOffset + maxNewlines
	newEnd = min(newEnd, len(tokens))

	return tokens[:newEnd]
}
