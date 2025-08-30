package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

type ImageRepository interface {
	GetCachedImage() (*ImageInfo, error)
}

type localImageRepository struct {
	imageDirectory     string
	cachedImageName    string
	imageCacheDuration time.Duration
	imageUrl           string
}

func NewLocalImageRepository(dir, name, url string, cacheDuration time.Duration) ImageRepository {
	return &localImageRepository{
		imageDirectory:     dir,
		cachedImageName:    name,
		imageCacheDuration: cacheDuration,
		imageUrl:           url,
	}
}

func (r *localImageRepository) GetCachedImage() (*ImageInfo, error) {
	imagePath := filepath.Join(r.imageDirectory, r.cachedImageName)

	if fileInfo, err := os.Stat(imagePath); err == nil {
		cacheAge := time.Since(fileInfo.ModTime())
		if cacheAge < r.imageCacheDuration {
			log.Info().
				Dur("cache_age", cacheAge).
				Str("path", imagePath).
				Msg("Serving cached image")
			return &ImageInfo{
				Path:     "/files/" + r.cachedImageName,
				CachedAt: fileInfo.ModTime(),
			}, nil
		}
		log.Info().
			Dur("cache_age", cacheAge).
			Str("path", imagePath).
			Msg("Cache expired, downloading new image")

	} else {
		log.Info().Msg("No cached image found, downloading new image")
	}

	return r.downloadNewImage()
}

func (r *localImageRepository) downloadNewImage() (*ImageInfo, error) {
	imagePath := filepath.Join(r.imageDirectory, r.cachedImageName)
	log.Info().
		Str("url", r.imageUrl).
		Str("path", imagePath).
		Msg("Downloading new image")

	resp, err := http.Get(r.imageUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("Error closing response body")
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: HTTP %d", resp.StatusCode)
	}

	file, err := os.Create(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create image file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error().Err(err).Str("path", imagePath).Msg("Error closing file")
		}
	}(file)

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to save image: %v", err)
	}

	now := time.Now()
	log.Info().
		Str("path", imagePath).
		Time("saved_at", now).
		Msg("Image downloaded and saved")
	return &ImageInfo{
		Path:     "/files/" + r.cachedImageName,
		CachedAt: now,
	}, nil
}
