package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	discordwebhook "github.com/bensch777/discord-webhook-golang"
	"github.com/nats-io/nats.go"

	"github.com/rs/zerolog/log"
)

type Todo struct {
	ID   int    `json:"id"`
	Task string `json:"task"`
	Done bool   `json:"done"`
}

func main() {
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		log.Error().Msg("DISCORD_WEBHOOK_URL is not set")
		return
	}

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		log.Error().Msg("NATS_URL is not set")
		return
	}

	mode := os.Getenv("MODE")
	if mode == "" {
		mode = "log-only"
		log.Warn().Msg("MODE is not set, defaulting to log-only")
	}

	var handleEvent func(event string, todo Todo)
	switch mode {
	case "forward":
		handleEvent = func(event string, todo Todo) {
			sendDiscordEmbed(webhookURL, event, todo)
			log.Info().Msg("Todo embed sent to Discord successfully")
		}
	default:
		handleEvent = func(event string, todo Todo) {
			log.Info().Msg("Running in log-only mode, not forwarding to Discord")
		}
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to NATS")
	}
	defer nc.Close()

	if _, err := nc.QueueSubscribe("todo.created", "broadcaster", func(m *nats.Msg) {
		var todo Todo
		if err := json.Unmarshal(m.Data, &todo); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal todo.created message")
			return
		}
		log.Info().Interface("todo", todo).Msg("Received todo.created event")
		handleEvent("Todo Created", todo)
	}); err != nil {
		log.Error().Err(err).Msg("Failed to subscribe to todo.created")
		return
	}

	if _, err := nc.QueueSubscribe("todo.updated", "broadcaster", func(m *nats.Msg) {
		var todo Todo
		if err := json.Unmarshal(m.Data, &todo); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal todo.updated message")
			return
		}
		log.Info().Interface("todo", todo).Msg("Received todo.updated event")
		handleEvent("Todo Updated", todo)
	}); err != nil {
		log.Error().Err(err).Msg("Failed to subscribe to todo.updated")
		return
	}

	select {}
}

func sendDiscordEmbed(webhookURL string, title string, todo Todo) {
	embed := discordwebhook.Embed{
		Title:     title,
		Color:     3066993,
		Timestamp: time.Now(),
		Fields: []discordwebhook.Field{
			{
				Name:   "ID",
				Value:  fmt.Sprintf("%d", todo.ID),
				Inline: true,
			},
			{
				Name:   "Task",
				Value:  todo.Task,
				Inline: true,
			},
			{
				Name:   "Done",
				Value:  fmt.Sprintf("%v", todo.Done),
				Inline: true,
			},
		},
		Footer: discordwebhook.Footer{
			Text: "NATS → Discord Broadcaster",
		},
	}

	hook := discordwebhook.Hook{
		Username: "Broadcaster Bot",
		Content:  "",
		Embeds:   []discordwebhook.Embed{embed},
	}

	payload, err := json.Marshal(hook)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal Discord payload")
		return
	}

	if err := discordwebhook.ExecuteWebhook(webhookURL, payload); err != nil {
		log.Error().Err(err).Msg("Failed to send message to Discord")
	} else {
		log.Info().Msg("Todo embed sent to Discord successfully")
	}
}
