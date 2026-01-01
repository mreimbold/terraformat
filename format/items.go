package format

import "github.com/hashicorp/hcl/v2/hclwrite"

func findTokenSpan(all hclwrite.Tokens, item hclwrite.Tokens) (int, int, bool) {
	if len(item) == 0 {
		return 0, -1, false
	}

	for i := 0; i <= len(all)-len(item); i++ {
		if all[i] != item[0] {
			continue
		}
		matched := true
		for j := 1; j < len(item); j++ {
			if all[i+j] != item[j] {
				matched = false
				break
			}
		}
		if matched {
			return i, i + len(item) - 1, true
		}
	}
	return 0, -1, false
}
