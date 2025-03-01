package chat

import (
	"log"
	"strings"

	"github.com/TwiN/go-away"
)

// Initialize profanity detector
var profanityDetector *goaway.ProfanityDetector

func LoadProfanities(profanities string) {
	log.Println("Loading profanities")
	customProfanities := strings.Split(profanities, ",")
	profanityDetector = goaway.NewProfanityDetector().WithCustomDictionary(customProfanities, nil, nil)
}

func Censor(str string) string {
	// Replace specific words and use the profanity detector
	str = strings.ReplaceAll(str, "хуй", "***")
	return profanityDetector.Censor(str)
}
