package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"

	"library-service/src/models"
	"library-service/src/utils"
)

type freeResult struct {
	result *models.SearchResult
	err    error
}

func (h *MovieHandler) SearchFree(c *gin.Context) {
	query, page, ok := parseQueryPage(c)
	if !ok {
		return
	}

	cacheKey := fmt.Sprintf("free:search:%s:page:%d", query, page)
	var cached models.SearchResult
	if cacheGet(cacheKey, &cached) {
		utils.RespondSuccess(c, http.StatusOK, cached)
		return
	}

	legitCh := make(chan freeResult, 1)
	archiveCh := make(chan freeResult, 1)

	go func() {
		r, err := h.legit.Search(query, page)
		legitCh <- freeResult{r, err}
	}()
	go func() {
		r, err := h.archive.Search(query, page)
		archiveCh <- freeResult{r, err}
	}()

	legitRes := <-legitCh
	archiveRes := <-archiveCh

	if legitRes.err != nil && archiveRes.err != nil {
		utils.RespondError(c, http.StatusBadGateway, "failed to fetch free torrent results")
		return
	}

	merged := deduplicateMovies(legitRes, archiveRes)

	totalPages := 1
	if archiveRes.result != nil && archiveRes.result.TotalPages > totalPages {
		totalPages = archiveRes.result.TotalPages
	}

	result := models.SearchResult{
		Page:       page,
		TotalPages: totalPages,
		Results:    merged,
	}

	cacheSet(cacheKey, result, freeCacheTTL)
	utils.RespondSuccess(c, http.StatusOK, result)
}

func deduplicateMovies(sources ...freeResult) []models.Movie {
	seen := make(map[string]struct{})
	var out []models.Movie
	for _, s := range sources {
		if s.err != nil || s.result == nil {
			continue
		}
		for _, m := range s.result.Results {
			key := normalizeTitle(m.Title)
			if key == "" {
				continue
			}
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, m)
		}
	}
	return out
}

func normalizeTitle(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
