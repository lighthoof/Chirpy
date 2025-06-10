package main

import (
	"strings"
)

func wordFilter(unfiltered string) string {
	filtered := []string{}
	filter := map[string]bool{"kerfuffle": true, "sharbert": true, "fornax": true}

	for _, word := range strings.Split(unfiltered, " ") {
		if filter[strings.ToLower(word)] {
			filtered = append(filtered, "****")
		} else {
			filtered = append(filtered, word)
		}
	}

	return strings.Join(filtered, " ")
}
