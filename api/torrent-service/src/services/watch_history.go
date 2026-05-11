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

// GetProgress returns the saved playback position in seconds for a user/movie pair.
// Returns 0 if no record exists yet.
func GetProgress(userID, movieID int) (int, error) {
	var entry models.WatchHistory
	err := conf.DB.
		Where("user_id = ? AND movie_id = ?", userID, movieID).
		First(&entry).Error
	if err != nil {
		return 0, nil // no record yet → start from beginning
	}
	return entry.ProgressSec, nil
}

// SaveProgress upserts the playback position in seconds for a user/movie pair.
func SaveProgress(userID, movieID, progressSec int) error {
	entry := models.WatchHistory{
		UserID:    userID,
		MovieID:   movieID,
		WatchedAt: time.Now(),
	}
	return conf.DB.
		Where(models.WatchHistory{UserID: userID, MovieID: movieID}).
		Assign(models.WatchHistory{ProgressSec: progressSec, WatchedAt: entry.WatchedAt}).
		FirstOrCreate(&entry).Error
}
