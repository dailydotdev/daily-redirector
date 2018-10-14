package main

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"unicode"
	"unicode/utf8"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getJson(req *http.Request, target interface{}) error {
	r, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

// Regexp definitions
var keyMatchRegex = regexp.MustCompile(`\"(\w+)\":`)

func MarshalJSON(v interface{}) ([]byte, error) {
	marshalled, err := json.Marshal(v)

	converted := keyMatchRegex.ReplaceAllFunc(
		marshalled,
		func(match []byte) []byte {
			// Empty keys are valid JSON, only lowercase if we do not have an
			// empty key.
			if len(match) > 2 {
				// Decode first rune after the double quotes
				r, width := utf8.DecodeRune(match[1:])
				r = unicode.ToLower(r)
				utf8.EncodeRune(match[1:width+1], r)
			}
			return match
		},
	)

	return converted, err
}
