package handlers

import (
	"net/http/httptest"

	"github.com/gin-gonic/gin"

	"library-service/src/client"
	"library-service/src/models"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newCtx creates a gin context with a recorder for testing handlers.
func newCtx(method, url string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, url, nil)
	return c, w
}

// mockTMDb implements tmdbClient for tests.
type mockTMDb struct {
	available    bool
	searchResult *models.SearchResult
	searchErr    error
	movieResult  *models.Movie
	movieErr     error
}

func (m *mockTMDb) Available() bool { return m.available }
func (m *mockTMDb) Search(_ string, _ int) (*models.SearchResult, error) {
	return m.searchResult, m.searchErr
}
func (m *mockTMDb) GetMovie(_ int) (*models.Movie, error) {
	return m.movieResult, m.movieErr
}

// mockSearch implements searchClient (omdb, etc.) for tests.
type mockSearch struct {
	available bool
	result    *models.SearchResult
	err       error
}

func (m *mockSearch) Available() bool { return m.available }
func (m *mockSearch) Search(_ string, _ int) (*models.SearchResult, error) {
	return m.result, m.err
}

// mockYTS implements ytsClient for tests.
type mockYTS struct {
	searchResult  *models.SearchResult
	searchErr     error
	imdbResult    *models.Movie
	imdbErr       error
}

func (m *mockYTS) Search(_ string, _ int) (*models.SearchResult, error) {
	return m.searchResult, m.searchErr
}
func (m *mockYTS) List(_ client.ListParams) (*models.SearchResult, error) {
	return m.searchResult, m.searchErr
}
func (m *mockYTS) GetMovieByIMDbID(_ string) (*models.Movie, error) {
	return m.imdbResult, m.imdbErr
}

// mockFree implements freeClient (legit, archive) for tests.
type mockFree struct {
	result *models.SearchResult
	err    error
}

func (m *mockFree) Search(_ string, _ int) (*models.SearchResult, error) {
	return m.result, m.err
}

func sampleMovies(titles ...string) []models.Movie {
	movies := make([]models.Movie, len(titles))
	for i, t := range titles {
		movies[i] = models.Movie{Title: t, Source: "test"}
	}
	return movies
}

func sampleResult(titles ...string) *models.SearchResult {
	return &models.SearchResult{
		Page:       1,
		TotalPages: 1,
		Results:    sampleMovies(titles...),
	}
}
