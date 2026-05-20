package censor

import (
	goaway "github.com/TwiN/go-away"
)

// Censor filters profanity from chat text.
type Censor struct {
	detector *goaway.ProfanityDetector
}

// New builds a Censor whose dictionary is extended with the given custom words.
func New(customProfanities []string) *Censor {
	detector := goaway.NewProfanityDetector().
		WithCustomDictionary(customProfanities, nil, nil)
	return &Censor{detector: detector}
}

// Clean returns the input with any detected profanity masked.
func (c *Censor) Clean(s string) string {
	return c.detector.Censor(s)
}
