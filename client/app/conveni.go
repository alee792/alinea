package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

func getenv(key string, _default string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return _default
}

func renderTemplate(w http.ResponseWriter, dir, tmpl string, d interface{}) {
	path := filepath.Join(fmt.Sprintf("templates/%s", dir), tmpl+".html")
	t, _ := template.ParseFiles(path)
	log.Debugf("Serving %s", path)
	t.Execute(w, d)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// ConvertibleBoolean allows strings and ints to be converted to bools
type ConvertibleBoolean bool

// UnmarshalJSON allows strings and ints to be converted to bools
func (bit *ConvertibleBoolean) UnmarshalJSON(data []byte) error {
	asString := string(data)
	if asString == "1" || asString == "true" {
		*bit = true
	} else if asString == "0" || asString == "false" {
		*bit = false
	} else {
		return fmt.Errorf("Boolean unmarshal error: invalid input %s", asString)
	}
	return nil
}
