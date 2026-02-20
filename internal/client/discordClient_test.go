package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/stretchr/testify/assert"
)

func Test_DiscordClient_SendDM_Success(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// Create DM channel
			assert.Equal(t, "/users/@me/channels", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "Bot test-bot-token", r.Header.Get("Authorization"))

			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, "123456", body["recipient_id"])

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"id": "dm-channel-id"})
		} else if callCount == 2 {
			// Send message
			assert.Equal(t, "/channels/dm-channel-id/messages", r.URL.Path)
			assert.Equal(t, "POST", r.Method)

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"id": "msg-id"})
		}
	}))
	defer server.Close()

	c := client.NewDiscordClientWithHTTPClient("test-bot-token", server.Client(), server.URL)

	embed := &client.DiscordEmbed{
		Title:       "Test",
		Description: "Test message",
		Color:       0x10b981,
	}

	err := c.SendDM(context.Background(), "123456", embed)
	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func Test_DiscordClient_SendDM_CreateChannelFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message": "Cannot send messages to this user"}`))
	}))
	defer server.Close()

	c := client.NewDiscordClientWithHTTPClient("test-bot-token", server.Client(), server.URL)

	err := c.SendDM(context.Background(), "123456", &client.DiscordEmbed{Title: "Test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func Test_DiscordClient_SendChannelMessage_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/channels/chan-123/messages", r.URL.Path)
		assert.Equal(t, "Bot test-bot-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "msg-id"})
	}))
	defer server.Close()

	c := client.NewDiscordClientWithHTTPClient("test-bot-token", server.Client(), server.URL)

	err := c.SendChannelMessage(context.Background(), "chan-123", &client.DiscordEmbed{Title: "Test"})
	assert.NoError(t, err)
}

func Test_DiscordClient_GetUserGuilds_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/@me/guilds", r.URL.Path)
		assert.Equal(t, "Bearer user-token-123", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "guild-1", "name": "Test Guild", "icon": "abc"},
			{"id": "guild-2", "name": "Other Guild", "icon": "def"},
		})
	}))
	defer server.Close()

	c := client.NewDiscordClientWithHTTPClient("test-bot-token", server.Client(), server.URL)

	guilds, err := c.GetUserGuilds(context.Background(), "user-token-123")
	assert.NoError(t, err)
	assert.Len(t, guilds, 2)
	assert.Equal(t, "guild-1", guilds[0].ID)
	assert.Equal(t, "Test Guild", guilds[0].Name)
}

func Test_DiscordClient_GetBotGuilds_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/@me/guilds", r.URL.Path)
		assert.Equal(t, "Bot test-bot-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "guild-1", "name": "Test Guild", "icon": "abc"},
		})
	}))
	defer server.Close()

	c := client.NewDiscordClientWithHTTPClient("test-bot-token", server.Client(), server.URL)

	guilds, err := c.GetBotGuilds(context.Background())
	assert.NoError(t, err)
	assert.Len(t, guilds, 1)
	assert.Equal(t, "guild-1", guilds[0].ID)
}

func Test_DiscordClient_GetGuildChannels_FiltersTextOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/guilds/guild-1/channels", r.URL.Path)
		assert.Equal(t, "Bot test-bot-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "chan-1", "name": "general", "type": 0},     // text
			{"id": "chan-2", "name": "voice", "type": 2},       // voice (filtered out)
			{"id": "chan-3", "name": "announcements", "type": 0}, // text
		})
	}))
	defer server.Close()

	c := client.NewDiscordClientWithHTTPClient("test-bot-token", server.Client(), server.URL)

	channels, err := c.GetGuildChannels(context.Background(), "guild-1")
	assert.NoError(t, err)
	assert.Len(t, channels, 2)
	assert.Equal(t, "general", channels[0].Name)
	assert.Equal(t, "announcements", channels[1].Name)
}

func Test_DiscordClient_GetUserGuilds_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "401: Unauthorized"}`))
	}))
	defer server.Close()

	c := client.NewDiscordClientWithHTTPClient("test-bot-token", server.Client(), server.URL)

	guilds, err := c.GetUserGuilds(context.Background(), "bad-token")
	assert.Error(t, err)
	assert.Nil(t, guilds)
	assert.Contains(t, err.Error(), "401")
}
