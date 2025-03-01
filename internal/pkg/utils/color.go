package utils

import (
	"math/rand"
	"time"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func GetRandomColor() string {
	letters := "6789ABCDEF"
	color := "#"
	for i := 0; i < 6; i++ {
		color += string(letters[rnd.Intn(len(letters))])
	}
	return color
}
