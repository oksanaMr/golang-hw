package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

func Top10(s string) []string {
	arr := strings.Fields(s)

	freq := make(map[string]int)
	for _, item := range arr {
		freq[item]++
	}

	unique := make([]string, 0, len(freq))
	for num := range freq {
		unique = append(unique, num)
	}

	sort.Slice(unique, func(i, j int) bool {
		if freq[unique[i]] == freq[unique[j]] {
			return unique[i] < unique[j]
		}
		return freq[unique[i]] > freq[unique[j]]
	})

	n := 10
	if len(unique) < n {
		n = len(unique)
	}

	return unique[:n]
}
