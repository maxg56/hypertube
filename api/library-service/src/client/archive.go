package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"library-service/src/models"
)

const (
	archiveBaseURL  = "https://archive.org"
	archivePageSize = 20
)

type ArchiveClient struct {
	httpClient *http.Client
}

func NewArchiveClient() *ArchiveClient {
	return &ArchiveClient{
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *ArchiveClient) Available() bool { return true }

type archiveSearchResp struct {
	Response struct {
		NumFound int          `json:"numFound"`
		Docs     []archiveDoc `json:"docs"`
	} `json:"response"`
}

type archiveDoc struct {
	Identifier  string      `json:"identifier"`
	Title       interface{} `json:"title"`
	Year        interface{} `json:"year"`
	Description interface{} `json:"description"`
}

func (c *ArchiveClient) Search(query string, page int) (*models.SearchResult, error) {
	if page < 1 {
		page = 1
	}
	start := (page - 1) * archivePageSize

	reqURL := fmt.Sprintf(
		"%s/advancedsearch.php?q=%s&fl[]=identifier&fl[]=title&fl[]=year&fl[]=description&output=json&rows=%d&start=%d",
		archiveBaseURL,
		url.QueryEscape(fmt.Sprintf("title:(%s) AND mediatype:movies", query)),
		archivePageSize,
		start,
	)

	body, err := c.get(reqURL)
	if err != nil {
		return nil, err
	}

	var raw archiveSearchResp
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("archive.org JSON parse: %w", err)
	}

	totalPages := (raw.Response.NumFound + archivePageSize - 1) / archivePageSize
	result := &models.SearchResult{
		Page:       page,
		TotalPages: totalPages,
	}
	for _, doc := range raw.Response.Docs {
		result.Results = append(result.Results, archiveDocToModel(doc))
	}
	return result, nil
}

func (c *ArchiveClient) get(u string) ([]byte, error) {
	resp, err := c.httpClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("archive.org request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("archive.org returned status %d", resp.StatusCode)
	}
	return body, nil
}

func archiveDocToModel(doc archiveDoc) models.Movie {
	title := archiveFirstString(doc.Title)
	overview := archiveFirstString(doc.Description)
	year := archiveFirstString(doc.Year)

	torrentURL := fmt.Sprintf("%s/download/%s/%s_archive.torrent",
		archiveBaseURL, doc.Identifier, doc.Identifier)

	return models.Movie{
		Title:    title,
		Year:     year,
		Overview: overview,
		Source:   "archive",
		Torrents: []models.Torrent{
			{
				URL:  torrentURL,
				Type: "archive",
			},
		},
	}
}

func archiveFirstString(v interface{}) string {
	switch s := v.(type) {
	case string:
		return s
	case []interface{}:
		if len(s) > 0 {
			if str, ok := s[0].(string); ok {
				return str
			}
		}
	}
	return ""
}

func archiveYearToString(v interface{}) string {
	switch s := v.(type) {
	case string:
		return s
	case float64:
		return strconv.Itoa(int(s))
	}
	return ""
}
