# Discord Bot Notifications

## Overview

Discord integration that sends marketplace notifications (e.g., new purchases) to users via a Discord Bot. Users link their Discord account via OAuth, configure notification targets (DM or specific channels), and enable per-event-type preferences.

**Status:** Implemented
**GitHub Issue:** #34

---

## Key Decisions

- **Discord Bot** (not webhooks) — enables DMs, guild/channel discovery, and richer control
- **Discord OAuth** for account linking — users don't need to know their Discord ID; the app discovers it automatically
- **New /settings page** — dedicated page for Discord configuration, separate from existing pages
- **Non-blocking notifications** — sent via goroutine after purchase tx commits; never fails the purchase
- **REST-only Discord client** — no WebSocket gateway needed; just HTTP calls to Discord API v10

---

## Database Schema

### Tables

**discord_links**
- `id` (PK, bigserial)
- `user_id` (FK → users, UNIQUE) — one Discord link per user
- `discord_user_id` (varchar 50) — Discord snowflake ID
- `discord_username` (varchar 100)
- `access_token`, `refresh_token` (text) — Discord OAuth tokens
- `token_expires_at` (timestamp)
- `created_at`, `updated_at`

**discord_notification_targets**
- `id` (PK, bigserial)
- `user_id` (FK → users)
- `target_type` (varchar 20) — `'dm'` or `'channel'`
- `channel_id` (varchar 50, nullable) — Discord channel snowflake (null for DM)
- `guild_name`, `channel_name` (varchar 100) — display names
- `is_active` (boolean, default true)
- `created_at`, `updated_at`
- Index on `user_id`

**notification_preferences**
- `id` (PK, bigserial)
- `target_id` (FK → discord_notification_targets, CASCADE DELETE)
- `event_type` (varchar 50) — e.g., `'purchase_created'`
- `is_enabled` (boolean, default true)
- UNIQUE on `(target_id, event_type)`

---

## Backend Architecture

### File Structure

```
internal/
├── client/discordClient.go          # Discord REST API client
├── repositories/discordNotifications.go  # CRUD for all 3 tables
├── controllers/discordNotifications.go   # HTTP handlers
├── updaters/notifications.go        # Notification delivery logic
└── models/models.go                 # Discord* model structs

cmd/industry-tool/cmd/
├── settings.go                      # DISCORD_BOT_TOKEN, DISCORD_CLIENT_ID, DISCORD_CLIENT_SECRET
└── root.go                          # Wiring (optional: only if DISCORD_BOT_TOKEN is set)
```

### API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/discord/link` | Get user's Discord link status |
| POST | `/v1/discord/link` | Save Discord link (from OAuth callback) |
| DELETE | `/v1/discord/link` | Unlink Discord account |
| GET | `/v1/discord/guilds` | Get available guilds (user ∩ bot) |
| GET | `/v1/discord/guilds/{id}/channels` | Get text channels in guild |
| GET | `/v1/discord/targets` | Get user's notification targets |
| POST | `/v1/discord/targets` | Create target (DM or channel) |
| PUT | `/v1/discord/targets/{id}` | Toggle target active/inactive |
| DELETE | `/v1/discord/targets/{id}` | Delete target |
| GET | `/v1/discord/targets/{id}/prefs` | Get preferences for target |
| POST | `/v1/discord/targets/{id}/prefs` | Upsert preference |
| POST | `/v1/discord/targets/{id}/test` | Send test notification |

### Discord Client Methods

- `SendDM(discordUserID, embed)` — creates DM channel then sends message
- `SendChannelMessage(channelID, embed)` — sends embed to channel
- `GetUserGuilds(userAccessToken)` — user's guilds (OAuth bearer)
- `GetBotGuilds()` — bot's guilds (bot bearer)
- `GetGuildChannels(guildID)` — text channels in guild (bot bearer)

### Notification Flow

1. Purchase completes (tx.Commit in purchases controller)
2. `go notifier.NotifyPurchase(ctx, purchase)` — fire-and-forget goroutine
3. Query `GetActiveTargetsForEvent(sellerUserID, "purchase_created")` — joins targets + preferences
4. For each target: send embed via DM or channel message
5. Errors logged, never propagated

---

## Frontend Architecture

### File Structure

```
frontend/
├── pages/
│   ├── settings.tsx                          # Thin wrapper
│   └── api/discord/
│       ├── login.ts                          # Redirect to Discord OAuth
│       ├── callback.ts                       # Exchange code, save link
│       ├── link.ts                           # GET/DELETE proxy
│       ├── guilds/
│       │   ├── index.ts                      # GET proxy
│       │   └── [id]/channels.ts              # GET proxy
│       └── targets/
│           ├── index.ts                      # GET/POST proxy
│           └── [id]/
│               ├── index.ts                  # PUT/DELETE proxy
│               ├── preferences.ts            # GET/POST proxy
│               └── test.ts                   # POST proxy
└── packages/
    ├── pages/settings.tsx                    # Auth guard + DiscordSettings
    └── components/settings/
        └── DiscordSettings.tsx               # Main settings component
```

### Discord OAuth Flow

1. User clicks "Link Discord Account" → `/api/discord/login`
2. Redirect to Discord OAuth (`identify guilds` scopes)
3. Discord callback → `/api/discord/callback`
4. Exchange code for tokens, fetch `/users/@me`
5. POST tokens + user info to backend `/v1/discord/link`
6. Redirect to `/settings?discord=linked`

### Settings Page Features

- **Discord Account**: Link/unlink button, shows linked username
- **Notification Targets**: Add DM target, add channel target (server → channel picker dialog)
- **Per-Target Controls**: Active toggle, test button, delete button
- **Event Preferences**: Per-target checkboxes for each event type

---

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DISCORD_BOT_TOKEN` | No | Bot token — enables Discord integration when set |
| `DISCORD_CLIENT_ID` | No | Discord OAuth application client ID |
| `DISCORD_CLIENT_SECRET` | No | Discord OAuth application client secret |

When `DISCORD_BOT_TOKEN` is not set, the Discord integration is completely disabled (no routes registered, no notifier injected).

---

## Event Types

| Event Type | Description | Trigger |
|------------|-------------|---------|
| `purchase_created` | Someone purchased from your listings | After purchase tx commits |

Future event types can be added by:
1. Adding to `EVENT_TYPES` array in `DiscordSettings.tsx`
2. Calling `notifier.Notify*(ctx, data)` at the trigger point
3. Building an embed in the updater

---

## Extensibility

- Add new event types without schema changes (just new `event_type` values)
- Add new target types beyond DM/channel if needed
- Notification delivery is interface-based (`PurchaseNotifier`) for easy testing
- Discord client is interface-based (`DiscordClientInterface`) for mocking
