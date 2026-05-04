package types

type DownloadRequest struct {
	MagnetURI string `json:"magnet_uri" binding:"required"`
	MovieID   int    `json:"movie_id" binding:"required"`
}

type DownloadResponse struct {
	InfoHash string `json:"info_hash"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

type StatusResponse struct {
	InfoHash   string  `json:"info_hash"`
	Status     string  `json:"status"`
	Progress   float64 `json:"progress"`
	Downloaded int64   `json:"downloaded_bytes"`
	FileSize   int64   `json:"file_size_bytes"`
	FilePath   string  `json:"file_path,omitempty"`
	ErrorMsg   string  `json:"error,omitempty"`
}
