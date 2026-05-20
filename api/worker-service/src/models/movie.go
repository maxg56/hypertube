package models

import "time"

type Movie struct {
	ID            int       `gorm:"primaryKey;column:id"`
	TmdbID        int       `gorm:"column:tmdb_id"`
	ImdbID        string    `gorm:"column:imdb_id"`
	Title         string    `gorm:"column:title"`
	OriginalTitle string    `gorm:"column:original_title"`
	Overview      string    `gorm:"column:overview"`
	ReleaseDate   time.Time `gorm:"column:release_date"`
	Runtime       int       `gorm:"column:runtime"`
	Rating        float64   `gorm:"column:rating"`
	VoteCount     int       `gorm:"column:vote_count"`
	PosterPath    string    `gorm:"column:poster_path"`
	BackdropPath  string    `gorm:"column:backdrop_path"`
	Language      string    `gorm:"column:language"`
	CachedAt      time.Time `gorm:"column:cached_at"`
}

func (Movie) TableName() string { return "movies" }
