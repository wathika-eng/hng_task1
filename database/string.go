package database

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
	"unicode"

	"gorm.io/datatypes"
)

// Struct to store string analysis results
type Properties struct {
	Length      int            `json:"length"`
	Palindrome  bool           `json:"is_palindrome"`
	UniqueChars int            `json:"unique_characters"`
	WordCount   int            `json:"word_count"`
	SHA256      string         `json:"sha256_hash"`
	CharFreqMap datatypes.JSON `json:"character_frequency_map" gorm:"type:jsonb"`
}

// Struct for storing user-submitted values
type Value struct {
	// Use SHA256 as primary key to match spec
	ID         string     `gorm:"primaryKey;type:char(64)" json:"id"`
	Value      string     `gorm:"uniqueIndex;not null" json:"value"`
	Properties Properties `gorm:"embedded" json:"properties"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

// AnalyzeString computes Properties for a given input string and returns a Value
func AnalyzeString(s string) (*Value, error) {
	props := Properties{}
	props.Length = len([]rune(s))
	props.Palindrome = isPalindrome(s)
	props.UniqueChars = uniqueCharsCount(s)
	props.WordCount = wordCount(s)
	props.SHA256 = sha256Hash(s)

	freq := charFrequencyMap(s)
	b, err := json.Marshal(freq)
	if err != nil {
		return nil, err
	}
	props.CharFreqMap = datatypes.JSON(b)

	v := &Value{
		ID:         props.SHA256,
		Value:      s,
		Properties: props,
		CreatedAt:  time.Now().UTC(),
	}
	return v, nil
}

func sha256Hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func isPalindrome(s string) bool {
	// case-insensitive, consider full unicode runes
	rs := []rune(strings.ToLower(s))
	i, j := 0, len(rs)-1
	for i < j {
		if rs[i] != rs[j] {
			return false
		}
		i++
		j--
	}
	return true
}

func uniqueCharsCount(s string) int {
	seen := map[rune]struct{}{}
	for _, r := range s {
		seen[r] = struct{}{}
	}
	return len(seen)
}

func wordCount(s string) int {
	fields := strings.FieldsFunc(s, unicodeIsSpace)
	if len(fields) == 0 {
		return 0
	}
	return len(fields)
}

func unicodeIsSpace(r rune) bool {
	return unicode.IsSpace(r)
}

func charFrequencyMap(s string) map[string]int {
	m := map[string]int{}
	for _, r := range s {
		key := string(r)
		m[key]++
	}
	return m
}
