package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"comment-service/src/conf"
	"comment-service/src/handlers"
)

type standardResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func respondJSON(w http.ResponseWriter, status int, body standardResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body) //nolint:errcheck
}

func notImplemented(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, standardResponse{
		Success: false,
		Error:   "comment service not yet implemented",
	})
}

func main() {
	if err := conf.InitDB(); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8005"
	}

	r := gin.Default()

<<<<<<< 14-library-détail-dun-film
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "comment-service"})
	})

	api := r.Group("/api/v1")
	{
		comments := api.Group("/comments")
		{
			comments.GET("/:movieId", handlers.ListComments)
			comments.POST("/:movieId", handlers.CreateComment)
			comments.DELETE("/:id", handlers.DeleteComment)
		}
	}
=======
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, standardResponse{
			Success: true,
			Data:    map[string]string{"status": "ok", "service": "comment-service"},
		})
	})

	mux.HandleFunc("/api/v1/comments/", notImplemented)
>>>>>>> main

	log.Printf("comment-service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
