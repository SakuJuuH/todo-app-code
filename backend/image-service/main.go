package main

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal().Msg("$PORT must be set")
	}

	imageDirectory := os.Getenv("IMAGE_DIR")
	imageUrl := getEnvWithDefault("IMAGE_URL", "https://picsum.photos/300")
	cachedImageName := getEnvWithDefault("CACHED_IMAGE_NAME", "current_image.jpg")
	imageCacheDuration := getCacheDuration()

	if imageDirectory == "" {
		cwd, _ := os.Getwd()
		parentDir := filepath.Dir(cwd)
		imageDirectory = filepath.Join(parentDir, "image")
	}

	err := os.MkdirAll(imageDirectory, 0755)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create image directory")
	}

	repo := NewLocalImageRepository(imageDirectory, cachedImageName, imageUrl, imageCacheDuration)
	controller := NewImageController(repo)

	router := gin.Default()

	router.Use(CorsMiddleware)

	router.Static("/api/image/files", imageDirectory)

	router.GET("/", controller.welcome)

	router.GET("/api/image/current", controller.getImageInfo)

	router.POST("/api/image/shutdown", controller.shutdownServer)

	log.Info().Str("port", port).Msg("Server started")

	err = router.Run(":" + port)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start server:")
	}
}
