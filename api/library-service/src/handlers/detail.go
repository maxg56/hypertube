package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"library-service/src/conf"
	"library-service/src/models"
	"library-service/src/utils"
)

var commentHTTPClient = &http.Client{Timeout: 5 * time.Second}

func commentServiceURL() string {
	u := os.Getenv("COMMENT_SERVICE_URL")
	if u == "" {
		u = "http://comment-service:8005"
	}
	return u
}

func (h *MovieHandler) GetMovie(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid movie id")
		return
	}

	cacheKey := fmt.Sprintf("movie:%d", id)
	var movie models.Movie
	fromCache := cacheGet(cacheKey, &movie)

	if !fromCache {
		var result *models.Movie

		// Try TMDB first with fallback to YTS for torrents
		if h.tmdb.Available() {
			var err error
			result, err = h.tmdb.GetMovie(id)
			if err == nil && result != nil && result.IMDbID != "" {
				// Got TMDB result, try to enrich with YTS torrents
				if ytsMovie, ytsErr := h.yts.GetMovieByIMDbID(result.IMDbID); ytsErr == nil && ytsMovie != nil {
					result.Torrents = ytsMovie.Torrents
				}
			}
		}

		// Fall back to YTS if TMDB didn't find anything
		if result == nil {
			var ytsErr error
			result, ytsErr = h.yts.GetMovieByID(id)
			if ytsErr != nil {
				// Log error but continue to check if result was found
				log.Printf("YTS GetMovieByID error: %v", ytsErr)
			}
		}

		if result == nil {
			utils.RespondError(c, http.StatusNotFound, "movie not found")
			return
		}

		for i, t := range result.Torrents {
			if t.Hash != "" {
				result.Torrents[i].Magnet = buildMagnet(t.Hash, result.Title)
			}
		}

		cacheSet(cacheKey, result, conf.MovieCacheTTL)
		movie = *result

		if conf.DB != nil {
			go upsertMovieToDB(&movie)
		}
	}

	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr != "" && conf.DB != nil {
		movie.Watched = checkWatched(id, userIDStr)
	}

	movie.Comments = fetchComments(id, c.GetHeader("X-JWT-Token"))

	utils.RespondSuccess(c, http.StatusOK, movie)
}

func buildMagnet(hash, title string) string {
	trackers := []string{
		"udp://open.demonii.com:1337/announce",
		"udp://tracker.openbittorrent.com:80",
		"udp://tracker.coppersurfer.tk:6969",
		"udp://tracker.opentrackr.org:1337/announce",
		"udp://exodus.desync.com:6969/announce",
	}
	var sb strings.Builder
	sb.WriteString("magnet:?xt=urn:btih:")
	sb.WriteString(strings.ToLower(hash))
	sb.WriteString("&dn=")
	sb.WriteString(url.QueryEscape(title))
	for _, tr := range trackers {
		sb.WriteString("&tr=")
		sb.WriteString(url.QueryEscape(tr))
	}
	return sb.String()
}

func upsertMovieToDB(m *models.Movie) {
	if conf.DB == nil {
		return
	}
	releaseDate := m.ReleaseDate
	if releaseDate == "" && m.Year != "" {
		releaseDate = m.Year + "-01-01"
	}

	result := conf.DB.Exec(`
		INSERT INTO movies (tmdb_id, imdb_id, title, overview, release_date, runtime, rating, poster_path, backdrop_path, cached_at)
		VALUES (?, ?, ?, ?, NULLIF(?, '')::DATE, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (tmdb_id) DO UPDATE SET
			imdb_id = EXCLUDED.imdb_id,
			title = EXCLUDED.title,
			overview = EXCLUDED.overview,
			release_date = EXCLUDED.release_date,
			runtime = EXCLUDED.runtime,
			rating = EXCLUDED.rating,
			poster_path = EXCLUDED.poster_path,
			backdrop_path = EXCLUDED.backdrop_path,
			cached_at = CURRENT_TIMESTAMP
	`, m.ID, m.IMDbID, m.Title, m.Overview, releaseDate, m.Runtime, m.Rating, m.PosterURL, m.BackdropURL)

	if result.Error != nil {
		log.Printf("upsertMovieToDB: %v", result.Error)
	}
}

func checkWatched(tmdbID int, userIDStr string) bool {
	if conf.DB == nil {
		return false
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return false
	}
	var count int64
	conf.DB.Raw(`
		SELECT COUNT(*) FROM watch_history wh
		INNER JOIN movies m ON m.id = wh.movie_id
		WHERE m.tmdb_id = ? AND wh.user_id = ?
	`, tmdbID, userID).Scan(&count)
	return count > 0
}

func fetchComments(tmdbID int, jwtToken string) []models.Comment {
	reqURL := fmt.Sprintf("%s/api/v1/comments/%d", commentServiceURL(), tmdbID)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil
	}
	if jwtToken != "" {
		req.Header.Set("Authorization", "Bearer "+jwtToken)
	}

	resp, err := commentHTTPClient.Do(req)
	if err != nil {
		log.Printf("fetchComments: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var payload struct {
		Success bool             `json:"success"`
		Data    []models.Comment `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}
	if payload.Data == nil {
		return []models.Comment{}
	}
	return payload.Data
}
