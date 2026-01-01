package format

import "github.com/hashicorp/hcl/v2/hclwrite"

type tokenSpan struct {
	start int
	end   int
}

func findTokenSpan(
	allTokens hclwrite.Tokens,
	itemTokens hclwrite.Tokens,
) (tokenSpan, bool) {
	if len(itemTokens) == zero {
		return tokenSpan{start: zero, end: negativeOne}, false
	}

	lastStart := len(allTokens) - len(itemTokens)

	for startIndex := zero; startIndex <= lastStart; startIndex++ {
		if !tokensMatchAt(allTokens, itemTokens, startIndex) {
			continue
		}

		endIndex := startIndex + len(itemTokens) - one

		return tokenSpan{start: startIndex, end: endIndex}, true
	}

	return tokenSpan{start: zero, end: negativeOne}, false
}

func tokensMatchAt(
	allTokens hclwrite.Tokens,
	itemTokens hclwrite.Tokens,
	startIndex int,
) bool {
	if allTokens[startIndex] != itemTokens[zero] {
		return false
	}

	for itemIndex := one; itemIndex < len(itemTokens); itemIndex++ {
		if allTokens[startIndex+itemIndex] != itemTokens[itemIndex] {
			return false
		}
	}

	return true
}
