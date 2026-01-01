// Package tokens provides helpers for token span matching.
package tokens

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/mreimbold/terraformat/internal/format/model"
)

// Span describes a contiguous token span.
type Span struct {
	Start int
	End   int
}

// FindSpan locates the token span for an item sequence.
func FindSpan(
	allTokens hclwrite.Tokens,
	itemTokens hclwrite.Tokens,
) (Span, bool) {
	//nolint:revive // add-constant: len check is clear here.
	if len(itemTokens) == 0 {
		return Span{Start: model.IndexFirst, End: model.IndexNotFound}, false
	}

	lastStart := len(allTokens) - len(itemTokens)

	for startIndex := model.IndexFirst; startIndex <= lastStart; startIndex++ {
		if !tokensMatchAt(allTokens, itemTokens, startIndex) {
			continue
		}

		endIndex := startIndex + len(itemTokens) - model.IndexOffset

		return Span{Start: startIndex, End: endIndex}, true
	}

	return Span{Start: model.IndexFirst, End: model.IndexNotFound}, false
}

func tokensMatchAt(
	allTokens hclwrite.Tokens,
	itemTokens hclwrite.Tokens,
	startIndex int,
) bool {
	if allTokens[startIndex] != itemTokens[model.IndexFirst] {
		return false
	}

	itemLen := len(itemTokens)
	for itemIndex := model.IndexOffset; itemIndex < itemLen; itemIndex++ {
		if allTokens[startIndex+itemIndex] != itemTokens[itemIndex] {
			return false
		}
	}

	return true
}
