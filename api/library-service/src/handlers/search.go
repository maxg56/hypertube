package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"library-service/src/conf"
	"library-service/src/models"
	"library-service/src/utils"
)

func (h *MovieHandler) Search(c *gin.Context) {
	query, page, ok := parseQueryPage(c)
	if !ok {
		return
	}

	cacheKey := fmt.Sprintf("search:%s:page:%d", query, page)
	var result models.SearchResult
	if cacheGet(cacheKey, &result) {
		utils.RespondSuccess(c, http.StatusOK, result)
		return
	}

	var searchResult *models.SearchResult
	var err error

	if h.tmdb.Available() {
		searchResult, err = h.tmdb.Search(query, page)
	}
	if (searchResult == nil || err != nil) && h.omdb.Available() {
		searchResult, err = h.omdb.Search(query, page)
	}

	if err != nil {
		utils.RespondError(c, http.StatusBadGateway, "failed to fetch search results")
		return
	}
	if searchResult == nil {
		utils.RespondError(c, http.StatusServiceUnavailable, "no metadata provider available")
		return
	}

	cacheSet(cacheKey, searchResult, conf.MovieCacheTTL)
	utils.RespondSuccess(c, http.StatusOK, searchResult)
}

func (h *MovieHandler) SearchYTS(c *gin.Context) {
	query, page, ok := parseQueryPage(c)
	if !ok {
		return
	}

	cacheKey := fmt.Sprintf("yts:search:%s:page:%d", query, page)
	var result models.SearchResult
	if cacheGet(cacheKey, &result) {
		utils.RespondSuccess(c, http.StatusOK, result)
		return
	}

	searchResult, err := h.yts.Search(query, page)
	if err != nil {
		utils.RespondError(c, http.StatusBadGateway, "failed to fetch YTS results")
		return
	}

	cacheSet(cacheKey, searchResult, ytsCacheTTL)
	utils.RespondSuccess(c, http.StatusOK, searchResult)
}
