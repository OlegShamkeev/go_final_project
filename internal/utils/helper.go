package utils

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

const secretLength = 20

func ValidateTaskID(id string) (int, error) {
	if len(id) == 0 {
		return 0, fmt.Errorf("no id parameter")
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return 0, err
	}
	return idInt, nil
}

func GenerateSecret() []byte {
	rnd := rand.NewSource(time.Now().Unix())
	result := make([]byte, 0, secretLength)
	for i := 0; i < secretLength; i++ {
		randomNumber := rnd.Int63()
		result = append(result, byte(randomNumber%26+97))
	}
	return result
}
