package client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	"library-service/src/models"
)

const legitBaseURL = "http://www.legittorrents.info"

type LegitTorrentsClient struct {
	httpClient *http.Client
}

func NewLegitTorrentsClient() *LegitTorrentsClient {
	return &LegitTorrentsClient{
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *LegitTorrentsClient) Available() bool { return true }

func (c *LegitTorrentsClient) Search(query string, page int) (*models.SearchResult, error) {
	if page < 1 {
		page = 1
	}

	params := url.Values{}
	params.Set("page", "main")
	params.Set("op", "search")
	params.Set("name", query)
	params.Set("category", "0")
	params.Set("Submit", "Go")

	reqURL := fmt.Sprintf("%s/?%s", legitBaseURL, params.Encode())
	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("legittorrents request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("legittorrents returned status %d", resp.StatusCode)
	}

	movies, err := parseLegitHTML(resp.Body)
	if err != nil {
		return nil, err
	}

	return &models.SearchResult{
		Page:       page,
		TotalPages: 1,
		Results:    movies,
	}, nil
}

func parseLegitHTML(r io.Reader) ([]models.Movie, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("legittorrents HTML parse: %w", err)
	}

	var movies []models.Movie
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			if m, ok := parseLegitRow(n); ok {
				movies = append(movies, m)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return movies, nil
}

func parseLegitRow(tr *html.Node) (models.Movie, bool) {
	var cells []*html.Node
	for c := tr.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "td" {
			cells = append(cells, c)
		}
	}
	if len(cells) < 4 {
		return models.Movie{}, false
	}

	title, href := legitExtractLink(cells[0])
	if title == "" || !strings.Contains(href, "torrent-details") {
		return models.Movie{}, false
	}

	torrentID := legitQueryParam(href, "id")
	if torrentID == "" {
		return models.Movie{}, false
	}

	primaryURL := fmt.Sprintf("%s/download.php?id=%s", legitBaseURL, torrentID)
	if legitIsHexHash(torrentID) {
		primaryURL = fmt.Sprintf("magnet:?xt=urn:btih:%s&dn=%s", torrentID, url.QueryEscape(title))
	}

	var size string
	var seeds int
	seedFound := false
	for i := 1; i < len(cells); i++ {
		text := strings.TrimSpace(legitNodeText(cells[i]))
		if legitIsSize(text) {
			size = text
		} else if !seedFound {
			if n, err := strconv.Atoi(text); err == nil && n >= 0 {
				seeds = n
				seedFound = true
			}
		}
	}

	return models.Movie{
		Title:  title,
		Source: "legittorrents",
		Torrents: []models.Torrent{
			{
				URL:   primaryURL,
				Hash:  torrentID,
				Type:  "web",
				Size:  size,
				Seeds: seeds,
			},
		},
	}, true
}

func legitExtractLink(n *html.Node) (text, href string) {
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					href = a.Val
				}
			}
			text = strings.TrimSpace(legitNodeText(n))
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return
}

func legitNodeText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(legitNodeText(c))
	}
	return sb.String()
}

func legitQueryParam(rawURL, key string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Query().Get(key)
}

func legitIsHexHash(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, c := range strings.ToLower(s) {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

func legitIsSize(s string) bool {
	return strings.Contains(s, "GB") || strings.Contains(s, "MB") || strings.Contains(s, "KB")
}
