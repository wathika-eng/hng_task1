package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"hng1/database"

	"github.com/gofiber/fiber/v2"
)

var store *database.Store

func main() {
	store = database.NewStore()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Default code is 500
			code := fiber.StatusInternalServerError
			var msg interface{} = "Internal Server Error"

			// Check if it's a *fiber.Error
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				msg = e.Message
			}

			// Log the error (like Echo)
			fmt.Printf("Error: %v\n", err)

			// Otherwise, send JSON
			return c.Status(code).JSON(fiber.Map{
				"error": msg,
			})
		},
	})

	app.Post("/strings", createString)
	app.Get("/strings/:value", getString)
	app.Get("/strings", listStrings)
	app.Get("/strings/filter-by-natural-language", naturalFilter)
	app.Delete("/strings/:value", deleteString)

	app.Listen(":3000")
}

type createReq struct {
	Value interface{} `json:"value"`
}

func createString(c *fiber.Ctx) error {
	var req createReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.Value == nil {
		return fiber.NewError(fiber.StatusBadRequest, `missing "value" field`)
	}
	s, ok := req.Value.(string)
	if !ok {
		return fiber.NewError(fiber.StatusUnprocessableEntity, `"value" must be a string`)
	}
	if strings.TrimSpace(s) == "" {
		return fiber.NewError(fiber.StatusBadRequest, `"value" must not be empty`)
	}

	v, err := database.AnalyzeString(s)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "analysis failed")
	}
	if err := store.Save(v); err != nil {
		if err == database.ErrExists {
			return fiber.NewError(fiber.StatusConflict, "string already exists")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "storage error")
	}
	return c.Status(fiber.StatusCreated).JSON(v)
}

func getString(c *fiber.Ctx) error {
	raw := c.Params("value")
	if raw == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing value param")
	}
	// URL param will be encoded by client; decode not needed here as Fiber gives decoded value
	v, ok := store.GetByValue(raw)
	if !ok {
		// try by hash id as fallback
		v, ok = store.GetByHash(raw)
		if !ok {
			return fiber.NewError(fiber.StatusNotFound, "string not found")
		}
	}
	return c.Status(fiber.StatusOK).JSON(v)
}

func listStrings(c *fiber.Ctx) error {
	// parse query params
	q := c.Query("is_palindrome")
	var isPal *bool
	if q != "" {
		b, err := strconv.ParseBool(q)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid is_palindrome")
		}
		isPal = &b
	}
	minLenStr := c.Query("min_length")
	maxLenStr := c.Query("max_length")
	var minLen, maxLen *int
	if minLenStr != "" {
		v, err := strconv.Atoi(minLenStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid min_length")
		}
		minLen = &v
	}
	if maxLenStr != "" {
		v, err := strconv.Atoi(maxLenStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid max_length")
		}
		maxLen = &v
	}
	wcStr := c.Query("word_count")
	var wc *int
	if wcStr != "" {
		v, err := strconv.Atoi(wcStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid word_count")
		}
		wc = &v
	}
	contains := c.Query("contains_character")

	all := store.GetAll()
	out := make([]*database.Value, 0)
	for _, v := range all {
		p := v.Properties
		if isPal != nil && p.Palindrome != *isPal {
			continue
		}
		if minLen != nil && p.Length < *minLen {
			continue
		}
		if maxLen != nil && p.Length > *maxLen {
			continue
		}
		if wc != nil && p.WordCount != *wc {
			continue
		}
		if contains != "" {
			if !strings.Contains(v.Value, contains) {
				continue
			}
		}
		out = append(out, v)
	}
	resp := fiber.Map{
		"data":  out,
		"count": len(out),
		"filters_applied": map[string]interface{}{
			"is_palindrome":      q,
			"min_length":         minLenStr,
			"max_length":         maxLenStr,
			"word_count":         wcStr,
			"contains_character": contains,
		},
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

func naturalFilter(c *fiber.Ctx) error {
	q := c.Query("query")
	if q == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing query param")
	}
	// very small heuristic parser
	parsed := map[string]interface{}{}
	lq := strings.ToLower(q)
	if strings.Contains(lq, "single word") || strings.Contains(lq, "single-word") {
		parsed["word_count"] = 1
	}
	if strings.Contains(lq, "palind") {
		parsed["is_palindrome"] = true
	}
	// "longer than N" or "longer than 10 characters"
	if strings.Contains(lq, "longer than") {
		parts := strings.Split(lq, "longer than")
		if len(parts) > 1 {
			tail := strings.TrimSpace(parts[1])
			// pick first number we find
			fields := strings.Fields(tail)
			for _, f := range fields {
				if n, err := strconv.Atoi(f); err == nil {
					// "longer than 10" -> min_length = 11
					parsed["min_length"] = n + 1
					break
				}
			}
		}
	}
	// contains the letter X
	if strings.Contains(lq, "contain") || strings.Contains(lq, "containing") || strings.Contains(lq, "containing the") {
		// handle "contain the first vowel"
		if strings.Contains(lq, "first vowel") {
			parsed["contains_character"] = "a"
		} else {
			// look for single-letter tokens or "letter x"
			toks := strings.Fields(lq)
			for i, tok := range toks {
				if len(tok) == 1 && tok >= "a" && tok <= "z" {
					parsed["contains_character"] = tok
					break
				}
				if tok == "letter" && i+1 < len(toks) {
					next := toks[i+1]
					if len(next) == 1 && next >= "a" && next <= "z" {
						parsed["contains_character"] = next
						break
					}
				}
			}
		}
	}

	if len(parsed) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "unable to parse natural language query")
	}

	// now filter the store using parsed
	all := store.GetAll()
	out := make([]*database.Value, 0)
	for _, v := range all {
		ok := true
		if val, found := parsed["word_count"]; found {
			if v.Properties.WordCount != val.(int) {
				ok = false
			}
		}
		if val, found := parsed["is_palindrome"]; found {
			if v.Properties.Palindrome != val.(bool) {
				ok = false
			}
		}
		if val, found := parsed["min_length"]; found {
			if v.Properties.Length < val.(int) {
				ok = false
			}
		}
		if val, found := parsed["contains_character"]; found {
			if !strings.Contains(v.Value, val.(string)) {
				ok = false
			}
		}
		if ok {
			out = append(out, v)
		}
	}

	resp := fiber.Map{
		"data":  out,
		"count": len(out),
		"interpreted_query": fiber.Map{
			"original":       q,
			"parsed_filters": parsed,
		},
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

func deleteString(c *fiber.Ctx) error {
	raw := c.Params("value")
	if raw == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing value param")
	}
	if err := store.DeleteByValue(raw); err != nil {
		if err == database.ErrNotFound {
			return fiber.NewError(fiber.StatusNotFound, "string not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "delete failed")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// helper for testing uniqueness of hash if needed
func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
