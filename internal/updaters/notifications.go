package updaters

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// PurchaseNotifier is the interface used by the purchases controller
type PurchaseNotifier interface {
	NotifyPurchase(ctx context.Context, purchase *models.PurchaseTransaction)
}

// PiStallNotifier is the interface used by the PI updater
type PiStallNotifier interface {
	NotifyPiStalls(ctx context.Context, userID int64, alerts []*PiStallAlert)
}

// PiStallAlert contains information about a single stalled planet
type PiStallAlert struct {
	CharacterName   string
	PlanetType      string
	SolarSystemName string
	StalledPins     []PiStalledPin
}

// PiStalledPin describes a single stalled pin
type PiStalledPin struct {
	PinCategory string // "extractor" or "factory"
	Reason      string // "expired", "stalled"
}

type NotificationsDiscordRepo interface {
	GetActiveTargetsForEvent(ctx context.Context, userID int64, eventType string) ([]*models.DiscordNotificationTarget, error)
	GetLinkByUser(ctx context.Context, userID int64) (*models.DiscordLink, error)
}

type NotificationsUpdater struct {
	repo          NotificationsDiscordRepo
	discordClient client.DiscordClientInterface
}

func NewNotifications(
	repo NotificationsDiscordRepo,
	discordClient client.DiscordClientInterface,
) *NotificationsUpdater {
	return &NotificationsUpdater{
		repo:          repo,
		discordClient: discordClient,
	}
}

// NotifyPurchase sends notifications to the seller when a purchase is made
func (u *NotificationsUpdater) NotifyPurchase(ctx context.Context, purchase *models.PurchaseTransaction) {
	targets, err := u.repo.GetActiveTargetsForEvent(ctx, purchase.SellerUserID, "purchase_created")
	if err != nil {
		log.Error("failed to get notification targets", "user_id", purchase.SellerUserID, "error", err)
		return
	}

	if len(targets) == 0 {
		return
	}

	embed := buildPurchaseEmbed(purchase)

	for _, target := range targets {
		var sendErr error
		switch target.TargetType {
		case "dm":
			link, err := u.repo.GetLinkByUser(ctx, target.UserID)
			if err != nil || link == nil {
				log.Error("failed to get discord link for DM target", "user_id", target.UserID, "error", err)
				continue
			}
			sendErr = u.discordClient.SendDM(ctx, link.DiscordUserID, embed)
		case "channel":
			if target.ChannelID == nil {
				log.Error("channel target has no channel_id", "target_id", target.ID)
				continue
			}
			sendErr = u.discordClient.SendChannelMessage(ctx, *target.ChannelID, embed)
		default:
			log.Error("unknown target type", "target_type", target.TargetType, "target_id", target.ID)
			continue
		}

		if sendErr != nil {
			log.Error("failed to send notification", "target_id", target.ID, "target_type", target.TargetType, "error", sendErr)
		}
	}
}

// SendTestNotification sends a test notification to verify a target is configured correctly
func (u *NotificationsUpdater) SendTestNotification(ctx context.Context, target *models.DiscordNotificationTarget, discordLink *models.DiscordLink) error {
	embed := &client.DiscordEmbed{
		Title:       "Test Notification",
		Description: "This is a test notification from Pinky.Tools. If you can see this, your notification target is configured correctly!",
		Color:       0x3b82f6, // Primary blue
		Footer: &client.DiscordEmbedFooter{
			Text: fmt.Sprintf("Sent at %s", time.Now().UTC().Format(time.RFC3339)),
		},
	}

	switch target.TargetType {
	case "dm":
		return u.discordClient.SendDM(ctx, discordLink.DiscordUserID, embed)
	case "channel":
		if target.ChannelID == nil {
			return fmt.Errorf("channel target has no channel_id")
		}
		return u.discordClient.SendChannelMessage(ctx, *target.ChannelID, embed)
	default:
		return fmt.Errorf("unknown target type: %s", target.TargetType)
	}
}

// NotifyPiStalls sends a single Discord notification for all newly stalled planets
func (u *NotificationsUpdater) NotifyPiStalls(ctx context.Context, userID int64, alerts []*PiStallAlert) {
	if len(alerts) == 0 {
		return
	}

	targets, err := u.repo.GetActiveTargetsForEvent(ctx, userID, "pi_stall")
	if err != nil {
		log.Error("failed to get notification targets for pi_stall", "user_id", userID, "error", err)
		return
	}

	if len(targets) == 0 {
		return
	}

	embed := buildPiStallEmbed(alerts)

	for _, target := range targets {
		var sendErr error
		switch target.TargetType {
		case "dm":
			link, err := u.repo.GetLinkByUser(ctx, target.UserID)
			if err != nil || link == nil {
				log.Error("failed to get discord link for DM target", "user_id", target.UserID, "error", err)
				continue
			}
			sendErr = u.discordClient.SendDM(ctx, link.DiscordUserID, embed)
		case "channel":
			if target.ChannelID == nil {
				log.Error("channel target has no channel_id", "target_id", target.ID)
				continue
			}
			sendErr = u.discordClient.SendChannelMessage(ctx, *target.ChannelID, embed)
		default:
			log.Error("unknown target type", "target_type", target.TargetType, "target_id", target.ID)
			continue
		}

		if sendErr != nil {
			log.Error("failed to send pi stall notification", "target_id", target.ID, "target_type", target.TargetType, "error", sendErr)
		}
	}
}

func buildPiStallEmbed(alerts []*PiStallAlert) *client.DiscordEmbed {
	description := fmt.Sprintf("**%d** planet(s) need attention", len(alerts))
	if len(alerts) == 1 {
		description = fmt.Sprintf("**%s**'s colony needs attention", alerts[0].CharacterName)
	}

	fields := []client.DiscordEmbedField{}
	for _, alert := range alerts {
		extractorCount := 0
		factoryCount := 0
		for _, pin := range alert.StalledPins {
			switch pin.PinCategory {
			case "extractor":
				extractorCount++
			case "factory":
				factoryCount++
			}
		}

		parts := []string{}
		if extractorCount > 0 {
			if extractorCount == 1 {
				parts = append(parts, "1 extractor expired")
			} else {
				parts = append(parts, fmt.Sprintf("%d extractors expired", extractorCount))
			}
		}
		if factoryCount > 0 {
			if factoryCount == 1 {
				parts = append(parts, "1 factory stalled")
			} else {
				parts = append(parts, fmt.Sprintf("%d factories stalled", factoryCount))
			}
		}

		name := fmt.Sprintf("%s — %s (%s)", alert.CharacterName, alert.SolarSystemName, alert.PlanetType)
		fields = append(fields, client.DiscordEmbedField{
			Name:   name,
			Value:  strings.Join(parts, ", "),
			Inline: false,
		})
	}

	return &client.DiscordEmbed{
		Title:       "PI Stall Detected",
		Description: description,
		Color:       0xef4444, // Red for alert
		Fields:      fields,
		Footer: &client.DiscordEmbedFooter{
			Text: fmt.Sprintf("Pinky.Tools • %s", time.Now().UTC().Format("Jan 2, 2006 15:04 UTC")),
		},
	}
}

var iskPrinter = message.NewPrinter(language.English)

func formatISK(value float64) string {
	return iskPrinter.Sprintf("%.2f ISK", value)
}

func buildPurchaseEmbed(purchase *models.PurchaseTransaction) *client.DiscordEmbed {
	return &client.DiscordEmbed{
		Title:       "New Purchase",
		Description: fmt.Sprintf("**%s** purchased from your listings", purchase.BuyerName),
		Color:       0x10b981, // Green for success/revenue
		Fields: []client.DiscordEmbedField{
			{
				Name:   "Item",
				Value:  purchase.TypeName,
				Inline: true,
			},
			{
				Name:   "Quantity",
				Value:  iskPrinter.Sprintf("%d", purchase.QuantityPurchased),
				Inline: true,
			},
			{
				Name:   "Price Per Unit",
				Value:  formatISK(purchase.PricePerUnit),
				Inline: true,
			},
			{
				Name:   "Total",
				Value:  formatISK(purchase.TotalPrice),
				Inline: true,
			},
			{
				Name:   "Location",
				Value:  purchase.LocationName,
				Inline: false,
			},
		},
		Footer: &client.DiscordEmbedFooter{
			Text: fmt.Sprintf("Pinky.Tools • %s", purchase.PurchasedAt.UTC().Format("Jan 2, 2006 15:04 UTC")),
		},
	}
}
