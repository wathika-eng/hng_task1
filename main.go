package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
)

type Value struct {
	Input     string    `xorm:"input not null unique" json:"input"`
	CreatedAt time.Time `xorm:"created_at created" json:"created_at"`
}

// var engine *xorm.Engine

func main() {
	logger := slog.Default()
	logger.Info("starting...")

	var err error
	engine, err := xorm.NewEngine("sqlite3", "strings.db")

	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := engine.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	if err := engine.Sync2(new(Value)); err != nil {
		log.Fatalf("failed to sync database schema: %v", err)
	}

	defer engine.Close()

	app := fiber.New(fiber.Config{
		ServerHeader: "hng2",
		AppName:      "HNG Task 2",
		ETag:         true,
		WriteTimeout: time.Duration(3000),
	})

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"time": time.Now().Local(), "status": "okay"})
	})

	app.Post("/strings", func(c *fiber.Ctx) error {
		// bind to struct
		var val Value
		if err := c.BodyParser(&val); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
		}
		// validate input
		if c.Body() == nil || strings.Trim(val.Input, " ") == "" || len(val.Input) <= 0 || len(val.Input) > 500 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "input must be a non-empty string with a maximum length of 500 characters"})
		}

		cleanedInput := strings.Trim(val.Input, " ")

		lenInput := len(cleanedInput)
		pal := isPalindrome(cleanedInput)
		sha := sha256Hash(cleanedInput)
		uniqChars := uniqueCharacters(cleanedInput)
		wordCnt := wordCount(cleanedInput)
		charFreqMap := characterFrequencyMap(cleanedInput)

		response := fiber.Map{
			"id":    sha,
			"value": cleanedInput,
			"properties": fiber.Map{
				"length":                  lenInput,
				"is_palindrome":           pal,
				"sha256_hash":             sha,
				"unique_characters":       uniqChars,
				"word_count":              wordCnt,
				"character_frequency_map": charFreqMap,
			},
			"created_at": time.Now().Format(time.RFC3339),
		}

		data, err := engine.Insert(&val)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save to database"})
		}

		logger.Info("inserted: ", "data", data)

		return c.JSON(response)
	})

	// app.Get("/strings/{}")

	app.Listen(":8000")
}

// is_palindrome: Boolean indicating if the string reads the same forwards and backwards (case-insensitive)
func isPalindrome(s string) bool {
	for i := 0; i < len(s)/2; i++ {
		if s[i] != s[len(s)-1-i] {
			return false
		}
	}
	return true
}

// unique_characters: Count of distinct characters in the string
func uniqueCharacters(s string) int {
	charMap := make(map[rune]bool)
	for _, char := range s {
		charMap[char] = true
	}
	return len(charMap)
}

// word_count: Number of words separated by whitespace
func wordCount(s string) int {
	words := strings.Fields(s)
	return len(words)
}

// sha256_hash: SHA-256 hash of the string for unique identification
func sha256Hash(s string) string {
	hash := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", hash)
}

// character_frequency_map: Object/dictionary mapping each character to its occurrence count
func characterFrequencyMap(s string) map[rune]int {
	freqMap := make(map[rune]int)
	for _, char := range s {
		freqMap[char]++
	}
	return freqMap
}
