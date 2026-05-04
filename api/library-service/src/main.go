package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"library-service/src/conf"
	"library-service/src/handlers"
)

func main() {
	if err := conf.InitRedis(); err != nil {
		log.Printf("Redis initialization failed: %v — caching disabled", err)
	} else {
		log.Println("Redis connected")
	}

	r := gin.Default()

	r.GET("/health", handlers.HealthCheckHandler)

	api := r.Group("/api/v1")
	{
		library := api.Group("/library")
		{
			movies := library.Group("/movies")
			h := handlers.NewMovieHandler()
			movies.GET("/search", h.Search)
			movies.GET("/yts", h.SearchYTS)
			movies.GET("/free", h.SearchFree)
			movies.GET("/:id", h.GetMovie)
		}
	}

	log.Println("library-service starting on port 8003")
	log.Fatal(http.ListenAndServe(":8003", r))
}
