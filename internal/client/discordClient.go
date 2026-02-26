package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

// DiscordAPIError represents a Discord API error mapped to an app-level error code
type DiscordAPIError struct {
	StatusCode   int
	DiscordCode  int    `json:"code"`
	Message      string `json:"message"`
	AppErrorCode string // app-level code sent to the frontend
}

func (e *DiscordAPIError) Error() string {
	return e.AppErrorCode
}

// Discord error code → app-level error code (frontend maps these to messages)
var discordErrorCodes = map[int]string{
	50007: "discord_dm_disabled",
	50001: "discord_no_channel_access",
	50013: "discord_missing_permissions",
	10003: "discord_unknown_channel",
}

// parseDiscordError reads the response body and returns a DiscordAPIError if possible
func parseDiscordError(statusCode int, body []byte) error {
	var apiErr DiscordAPIError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.DiscordCode != 0 {
		apiErr.StatusCode = statusCode
		if code, ok := discordErrorCodes[apiErr.DiscordCode]; ok {
			apiErr.AppErrorCode = code
		} else {
			apiErr.AppErrorCode = "discord_unknown_error"
		}
		return &apiErr
	}
	return fmt.Errorf("discord API error: status %d, body: %s", statusCode, strings.TrimSpace(string(body)))
}

//go:generate mockgen -source=./discordClient.go -destination=./discordClient_mock_test.go -package=client_test

// DiscordClientInterface abstracts the Discord API for testing
type DiscordClientInterface interface {
	SendDM(ctx context.Context, discordUserID string, embed *DiscordEmbed) error
	SendChannelMessage(ctx context.Context, channelID string, embed *DiscordEmbed) error
	GetUserGuilds(ctx context.Context, userAccessToken string) ([]*models.DiscordGuild, error)
	GetBotGuilds(ctx context.Context) ([]*models.DiscordGuild, error)
	GetGuildChannels(ctx context.Context, guildID string) ([]*models.DiscordChannel, error)
}

// DiscordEmbed represents a Discord message embed
type DiscordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	URL         string              `json:"url,omitempty"`
	Color       int                 `json:"color,omitempty"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
	Footer      *DiscordEmbedFooter `json:"footer,omitempty"`
}

// DiscordEmbedField represents a field in a Discord embed
type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// DiscordEmbedFooter represents a footer in a Discord embed
type DiscordEmbedFooter struct {
	Text string `json:"text"`
}

type discordCreateDMRequest struct {
	RecipientID string `json:"recipient_id"`
}

type discordCreateDMResponse struct {
	ID string `json:"id"`
}

type discordSendMessageRequest struct {
	Embeds []DiscordEmbed `json:"embeds"`
}

type DiscordClient struct {
	botToken string
	client   HTTPDoer
	baseURL  string
}

func NewDiscordClient(botToken string) *DiscordClient {
	return NewDiscordClientWithHTTPClient(botToken, &http.Client{}, "")
}

func NewDiscordClientWithHTTPClient(botToken string, httpClient HTTPDoer, baseURL string) *DiscordClient {
	if baseURL == "" {
		baseURL = "https://discord.com/api/v10"
	}
	return &DiscordClient{
		botToken: botToken,
		client:   httpClient,
		baseURL:  baseURL,
	}
}

// SendDM sends a direct message embed to a Discord user via the bot
func (c *DiscordClient) SendDM(ctx context.Context, discordUserID string, embed *DiscordEmbed) error {
	// Step 1: Create a DM channel with the user
	dmReqBody, err := json.Marshal(discordCreateDMRequest{RecipientID: discordUserID})
	if err != nil {
		return errors.Wrap(err, "failed to marshal DM request")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/users/@me/channels", bytes.NewReader(dmReqBody))
	if err != nil {
		return errors.Wrap(err, "failed to create DM channel request")
	}
	req.Header.Set("Authorization", "Bot "+c.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to create DM channel")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return errors.Wrap(parseDiscordError(resp.StatusCode, body), "failed to create DM channel")
	}

	var dmResp discordCreateDMResponse
	if err := json.NewDecoder(resp.Body).Decode(&dmResp); err != nil {
		return errors.Wrap(err, "failed to decode DM channel response")
	}

	// Step 2: Send the message to the DM channel
	return c.SendChannelMessage(ctx, dmResp.ID, embed)
}

// SendChannelMessage sends an embed message to a Discord channel via the bot
func (c *DiscordClient) SendChannelMessage(ctx context.Context, channelID string, embed *DiscordEmbed) error {
	msgBody, err := json.Marshal(discordSendMessageRequest{Embeds: []DiscordEmbed{*embed}})
	if err != nil {
		return errors.Wrap(err, "failed to marshal message request")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/channels/%s/messages", c.baseURL, channelID), bytes.NewReader(msgBody))
	if err != nil {
		return errors.Wrap(err, "failed to create message request")
	}
	req.Header.Set("Authorization", "Bot "+c.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return errors.Wrap(parseDiscordError(resp.StatusCode, body), "failed to send message")
	}

	return nil
}

// GetUserGuilds returns the guilds the user belongs to (using user's OAuth access token)
func (c *DiscordClient) GetUserGuilds(ctx context.Context, userAccessToken string) ([]*models.DiscordGuild, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/users/@me/guilds", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create user guilds request")
	}
	req.Header.Set("Authorization", "Bearer "+userAccessToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user guilds")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user guilds: status %d, body: %s", resp.StatusCode, string(body))
	}

	guilds := []*models.DiscordGuild{}
	if err := json.NewDecoder(resp.Body).Decode(&guilds); err != nil {
		return nil, errors.Wrap(err, "failed to decode user guilds")
	}

	return guilds, nil
}

// GetBotGuilds returns the guilds the bot belongs to
func (c *DiscordClient) GetBotGuilds(ctx context.Context) ([]*models.DiscordGuild, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/users/@me/guilds", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create bot guilds request")
	}
	req.Header.Set("Authorization", "Bot "+c.botToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get bot guilds")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get bot guilds: status %d, body: %s", resp.StatusCode, string(body))
	}

	guilds := []*models.DiscordGuild{}
	if err := json.NewDecoder(resp.Body).Decode(&guilds); err != nil {
		return nil, errors.Wrap(err, "failed to decode bot guilds")
	}

	return guilds, nil
}

// GetGuildChannels returns text channels in a guild visible to the bot
func (c *DiscordClient) GetGuildChannels(ctx context.Context, guildID string) ([]*models.DiscordChannel, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/guilds/%s/channels", c.baseURL, guildID), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create guild channels request")
	}
	req.Header.Set("Authorization", "Bot "+c.botToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get guild channels")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get guild channels: status %d, body: %s", resp.StatusCode, string(body))
	}

	allChannels := []*models.DiscordChannel{}
	if err := json.NewDecoder(resp.Body).Decode(&allChannels); err != nil {
		return nil, errors.Wrap(err, "failed to decode guild channels")
	}

	// Filter to only text channels (type 0)
	textChannels := []*models.DiscordChannel{}
	for _, ch := range allChannels {
		if ch.Type == 0 {
			textChannels = append(textChannels, ch)
		}
	}

	return textChannels, nil
}
