package models

type Movie struct {
	ID         int    `gorm:"primaryKey;column:id"`
	TmdbID     int    `gorm:"column:tmdb_id;uniqueIndex;not null"`
	Title      string `gorm:"column:title;not null"`
	PosterPath string `gorm:"column:poster_path"`
}

func (Movie) TableName() string { return "movies" }
