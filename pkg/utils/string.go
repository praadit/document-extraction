package utils

import (
	"math/rand"
	"strings"
	"time"
)

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const numeric = "0123456789"

func RandomString(n int, prefix, charset string) string {
	randomSet := letters + "" + numeric
	switch strings.ToLower(charset) {
	case "alphabet":
		randomSet = letters
	case "numeric":
		randomSet = numeric
	default:
		randomSet = letters + "" + numeric
	}
	arrRandSet := strings.Split(randomSet, "")

	randomStr := prefix
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < n; i++ {
		randomStr += arrRandSet[rand.Intn(len(arrRandSet)-1)]
	}
	return randomStr
}
