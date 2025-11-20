package analyze

import "strings"

func makeIgnoreSet(items []string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		set[item] = struct{}{}
	}
	return set
}
