package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type TodosController struct {
	repo TodoRepository
}

func NewTodosController(repo TodoRepository) *TodosController {
	return &TodosController{repo: repo}
}

func (c *TodosController) getTodos(ctx *gin.Context) {
	todos, err := c.repo.GetTodos()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get todos")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Info().
		Str("path", ctx.FullPath()).
		Int("count", len(todos)).
		Msg("Todos received")
	ctx.JSON(http.StatusOK, todos)
}

func (c *TodosController) createTodo(ctx *gin.Context) {
	var requestTodo struct {
		Task string `json:"task" binding:"required"`
	}

	if err := ctx.BindJSON(&requestTodo); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, ok := validateAndLogTask(ctx, requestTodo.Task)
	if !ok {
		return
	}

	newTodo, err := c.repo.AddTodo(task)
	if err != nil {
		log.Error().Err(err).Msg("todo insert failed")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.sendNatsMessage("todo.created", newTodo)

	ctx.JSON(http.StatusCreated, newTodo)
}

func (c *TodosController) welcome(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message":     "Welcome to the Todo API! Use /api/todos to manage your tasks.",
		"status_code": http.StatusOK,
		"Endpoints": []string{
			"GET /api/todos - Retrieve all todos",
			"POST /api/todos - Create a new todo",
			"PUT /api/todos/:id - Mark a todo as done",
			"POST /api/todos/random - Create a random todo",
			"GET /api/todos/db-health - Check database connectivity",
			"GET /api/todos/healthz - Health check endpoint",
		},
	})
}

func (c *TodosController) createRandomTodo(ctx *gin.Context) {
	randomArticleURL := os.Getenv("RANDOM_ARTICLE_URL")
	if randomArticleURL == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "RANDOM_ARTICLE_URL is not set"})
		return
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", randomArticleURL, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create request")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.132 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get random article")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("Error closing response body")
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Error().
			Int("status_code", resp.StatusCode).
			Str("url", randomArticleURL).
			Msg("Failed to fetch random article")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch random article"})
		return
	}

	redirectedURL := resp.Request.URL.String()

	if redirectedURL == randomArticleURL {
		log.Warn().
			Str("path", ctx.FullPath()).
			Str("url", redirectedURL).
			Msg("random article rejected: same as RANDOM_ARTICLE_URL")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot read the same article twice"})
		return
	}

	task := fmt.Sprintf("Read: %s", redirectedURL)
	task, ok := validateAndLogTask(ctx, task)
	if !ok {
		return
	}

	createdTodo, err := c.repo.AddTodo(task)
	if err != nil {
		log.Error().Err(err).Msg("random todo insert failed")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.sendNatsMessage("todo.created", createdTodo)

	ctx.JSON(http.StatusCreated, gin.H{
		"New todo created": createdTodo,
	})
}

func (c *TodosController) markTodoDone(ctx *gin.Context) {
	idParam := ctx.Param("id")
	if idParam == "" {
		log.Warn().
			Str("path", ctx.FullPath()).
			Msg("mark todo done failed: missing id parameter")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing id parameter"})
		return
	}

	idParamInt, err := strconv.Atoi(idParam)
	if err != nil {
		log.Warn().
			Str("path", ctx.FullPath()).
			Msg("mark todo done failed: invalid id parameter")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id parameter"})
		return
	}

	var todo Todo
	todo, err = c.repo.markTodoDone(idParamInt)
	if err != nil {
		log.Error().Err(err).Msg("mark todo done failed")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Info().
		Str("path", ctx.FullPath()).
		Str("id", idParam).
		Msg("Todo marked as done")

	c.sendNatsMessage("todo.updated", todo)

	ctx.JSON(http.StatusOK, gin.H{"Todo updated": todo})
}

func (c *TodosController) dbHealthCheck(ctx *gin.Context) {
	health, err := c.repo.dbHealthCheck()
	if err != nil {
		log.Error().Err(err).Msg("Database health check failed")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database is not reachable"})
		return
	}
	if !health {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database is not reachable"})
		return
	}
	msg := "Database is reachable"
	log.Info().Msg(msg)
	ctx.JSON(http.StatusOK, gin.H{"message": msg})
}

func (c *TodosController) healthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "Service is healthy"})
}

func validateAndLogTask(ctx *gin.Context, task string) (string, bool) {
	const maxLen = 140
	length := len(task)

	if task == "" {
		log.Warn().
			Str("path", ctx.FullPath()).
			Int("length", length).
			Msg("todo rejected: empty task")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Task cannot be empty"})
		return "", false
	}

	if length > maxLen {
		log.Warn().
			Str("path", ctx.FullPath()).
			Int("length", length).
			Int("max_length", maxLen).
			Str("task", task).
			Msg("todo rejected: too long")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Task cannot exceed 140 characters"})
		return "", false
	}

	log.Info().
		Str("path", ctx.FullPath()).
		Int("length", length).
		Str("task", task).
		Msg("todo received")

	return task, true
}

func (c *TodosController) sendNatsMessage(subject string, todo Todo) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		log.Warn().Msg("NATS_URL is not set, skipping NATS message sending")
		return
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to NATS")
		return
	}
	defer nc.Close()

	data, err := json.Marshal(todo)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal todo for NATS message")
		return
	}

	if err := nc.Publish(subject, data); err != nil {
		log.Error().Err(err).Msg("Failed to publish NATS message")
		return
	}
}
