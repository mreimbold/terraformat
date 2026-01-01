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
	//nolint:revive // add-constant: len check is clear here.
	if len(itemTokens) == 0 {
		return tokenSpan{start: indexFirst, end: indexNotFound}, false
	}

	lastStart := len(allTokens) - len(itemTokens)

	for startIndex := indexFirst; startIndex <= lastStart; startIndex++ {
		if !tokensMatchAt(allTokens, itemTokens, startIndex) {
			continue
		}

		endIndex := startIndex + len(itemTokens) - indexOffset

		return tokenSpan{start: startIndex, end: endIndex}, true
	}

	return tokenSpan{start: indexFirst, end: indexNotFound}, false
}

func tokensMatchAt(
	allTokens hclwrite.Tokens,
	itemTokens hclwrite.Tokens,
	startIndex int,
) bool {
	if allTokens[startIndex] != itemTokens[indexFirst] {
		return false
	}

	for itemIndex := indexOffset; itemIndex < len(itemTokens); itemIndex++ {
		if allTokens[startIndex+itemIndex] != itemTokens[itemIndex] {
			return false
		}
	}

	return true
}
