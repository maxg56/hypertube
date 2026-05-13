package utils

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	db "auth-service/src/conf"
	models "auth-service/src/models"
)

// GenerateUsernameSuggestions returns available username alternatives using a
// single WHERE IN query instead of one query per candidate.
func GenerateUsernameSuggestions(baseUsername string) []string {
	baseUsername = strings.ToLower(baseUsername)
	candidates := [5]string{
		baseUsername + strconv.Itoa(generateRandomNumber(10, 99)),
		baseUsername + strconv.Itoa(time.Now().Year()),
		baseUsername + strconv.Itoa(generateRandomNumber(100, 999)),
		baseUsername + "_" + strconv.Itoa(generateRandomNumber(1, 999)),
		baseUsername + strconv.Itoa(generateRandomNumber(1000, 9999)),
	}

	var taken []string
	db.DB.Model(&models.Users{}).
		Where("username IN ?", candidates[:]).
		Pluck("username", &taken)

	takenSet := make(map[string]struct{}, len(taken))
	for _, u := range taken {
		takenSet[u] = struct{}{}
	}

	suggestions := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if _, exists := takenSet[c]; !exists {
			suggestions = append(suggestions, c)
		}
	}
	return suggestions
}

func generateRandomNumber(min, max int) int {
	return min + rand.Intn(max-min+1)
}
