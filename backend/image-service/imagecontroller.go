package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type ImageInfo struct {
	Path     string    `json:"path"`
	CachedAt time.Time `json:"cached_at"`
}

type ImageController struct {
	repo ImageRepository
}

func NewImageController(repo ImageRepository) *ImageController {
	return &ImageController{repo: repo}
}

func (c *ImageController) welcome(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message":     "Welcome to the image todo-service. ",
		"status_code": http.StatusOK,
		"Endpoints": []string{
			"GET api/image/current - Get cached image info",
			"POST api/image/shutdown - Shutdown the server",
		},
	})
}

func (c *ImageController) getImageInfo(ctx *gin.Context) {
	imageInfo, err := c.repo.GetCachedImage()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get image info")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Info().
		Str("path", imageInfo.Path).
		Msg("Image info received")
	ctx.JSON(http.StatusOK, imageInfo)
}

func (c *ImageController) shutdownServer(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "Shutting down server..."})
	go func() {
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()
}
