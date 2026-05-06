package services

import (
	"time"

	"torrent-service/src/conf"
	"torrent-service/src/models"
)

// RecordWatch upserts a watch_history row for the given user/movie pair.
func RecordWatch(userID, movieID int) error {
	entry := models.WatchHistory{
		UserID:    userID,
		MovieID:   movieID,
		WatchedAt: time.Now(),
	}
	return conf.DB.
		Where(models.WatchHistory{UserID: userID, MovieID: movieID}).
		Assign(models.WatchHistory{WatchedAt: entry.WatchedAt}).
		FirstOrCreate(&entry).Error
}

// IsWatched returns true when the user has a watch_history record for the movie.
func IsWatched(userID, movieID int) (bool, error) {
	var count int64
	err := conf.DB.Model(&models.WatchHistory{}).
		Where("user_id = ? AND movie_id = ?", userID, movieID).
		Count(&count).Error
	return count > 0, err
}
