package updaters

import (
	"context"
	"fmt"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
)

// regionNames maps common EVE region IDs to their display names.
var regionNames = map[int64]string{
	10000002: "The Forge",
	10000043: "Domain",
	10000030: "Heimatar",
	10000042: "Metropolis",
	10000032: "Sinq Laison",
	10000064: "Essence",
	10000037: "Everyshore",
	10000068: "Verge Vendor",
	10000033: "The Citadel",
	10000044: "Solitude",
	10000048: "Placid",
	10000049: "Khanid",
	10000052: "Kador",
	10000065: "Kor-Azor",
	10000067: "Genesis",
	10000069: "Aridia",
}

func regionName(id int64) string {
	if name, ok := regionNames[id]; ok {
		return name
	}
	return fmt.Sprintf("Region %d", id)
}

// HaulingNotificationsUpdater sends Discord notifications for hauling run events.
type HaulingNotificationsUpdater struct {
	repo        NotificationsDiscordRepo
	discord     client.DiscordClientInterface
	frontendURL string
}

// NewHaulingNotifications creates a new HaulingNotificationsUpdater.
func NewHaulingNotifications(
	repo NotificationsDiscordRepo,
	discord client.DiscordClientInterface,
	frontendURL string,
) *HaulingNotificationsUpdater {
	return &HaulingNotificationsUpdater{
		repo:        repo,
		discord:     discord,
		frontendURL: frontendURL,
	}
}

// NotifyHaulingTier2 sends a "run ready to ship" notification when fill threshold is crossed.
func (u *HaulingNotificationsUpdater) NotifyHaulingTier2(ctx context.Context, userID int64, run *models.HaulingRun, fillPct float64) {
	targets, err := u.repo.GetActiveTargetsForEvent(ctx, userID, "hauling_tier2")
	if err != nil {
		log.Error("failed to get notification targets for hauling_tier2", "user_id", userID, "error", err)
		return
	}
	if len(targets) == 0 {
		return
	}

	embed := buildHaulingTier2Embed(run, fillPct)
	u.sendToTargets(ctx, targets, embed, "hauling_tier2")
}

// NotifyHaulingComplete sends a run completion notification with P&L summary.
func (u *HaulingNotificationsUpdater) NotifyHaulingComplete(ctx context.Context, userID int64, run *models.HaulingRun, summary *models.HaulingRunPnlSummary) {
	targets, err := u.repo.GetActiveTargetsForEvent(ctx, userID, "hauling_complete")
	if err != nil {
		log.Error("failed to get notification targets for hauling_complete", "user_id", userID, "error", err)
		return
	}
	if len(targets) == 0 {
		return
	}

	embed := buildHaulingCompleteEmbed(run, summary)
	u.sendToTargets(ctx, targets, embed, "hauling_complete")
}

// NotifyHaulingItemSold sends notification when a run item is fully sold.
func (u *HaulingNotificationsUpdater) NotifyHaulingItemSold(ctx context.Context, userID int64, run *models.HaulingRun, item *models.HaulingRunItem) {
	targets, err := u.repo.GetActiveTargetsForEvent(ctx, userID, "hauling_item_sold")
	if err != nil {
		log.Error("failed to get notification targets for hauling_item_sold", "user_id", userID, "error", err)
		return
	}
	if len(targets) == 0 {
		return
	}

	embed := buildHaulingItemSoldEmbed(run, item)
	u.sendToTargets(ctx, targets, embed, "hauling_item_sold")
}

// SendHaulingDailyDigest sends a daily summary of active hauling runs.
func (u *HaulingNotificationsUpdater) SendHaulingDailyDigest(ctx context.Context, userID int64, runs []*models.HaulingRun) {
	if len(runs) == 0 {
		return
	}

	targets, err := u.repo.GetActiveTargetsForEvent(ctx, userID, "hauling_daily_digest")
	if err != nil {
		log.Error("failed to get notification targets for hauling_daily_digest", "user_id", userID, "error", err)
		return
	}
	if len(targets) == 0 {
		return
	}

	embed := buildHaulingDailyDigestEmbed(runs)
	u.sendToTargets(ctx, targets, embed, "hauling_daily_digest")
}

func (u *HaulingNotificationsUpdater) sendToTargets(ctx context.Context, targets []*models.DiscordNotificationTarget, embed *client.DiscordEmbed, eventType string) {
	for _, target := range targets {
		var sendErr error
		switch target.TargetType {
		case "dm":
			link, err := u.repo.GetLinkByUser(ctx, target.UserID)
			if err != nil || link == nil {
				log.Error("failed to get discord link for DM target", "user_id", target.UserID, "error", err)
				continue
			}
			sendErr = u.discord.SendDM(ctx, link.DiscordUserID, embed)
		case "channel":
			if target.ChannelID == nil {
				log.Error("channel target has no channel_id", "target_id", target.ID)
				continue
			}
			sendErr = u.discord.SendChannelMessage(ctx, *target.ChannelID, embed)
		default:
			log.Error("unknown target type", "target_type", target.TargetType, "target_id", target.ID)
			continue
		}

		if sendErr != nil {
			log.Error("failed to send hauling notification",
				"event_type", eventType,
				"target_id", target.ID,
				"target_type", target.TargetType,
				"error", sendErr,
			)
		}
	}
}

func buildHaulingTier2Embed(run *models.HaulingRun, fillPct float64) *client.DiscordEmbed {
	route := fmt.Sprintf("%s → %s", regionName(run.FromRegionID), regionName(run.ToRegionID))
	fields := []client.DiscordEmbedField{
		{Name: "Route", Value: route, Inline: true},
		{Name: "Fill", Value: fmt.Sprintf("%.1f%%", fillPct), Inline: true},
	}
	if run.HaulThresholdISK != nil {
		fields = append(fields, client.DiscordEmbedField{
			Name:   "Threshold",
			Value:  fmt.Sprintf("%.2f ISK", *run.HaulThresholdISK),
			Inline: true,
		})
	}

	return &client.DiscordEmbed{
		Title:       "Hauling Run Ready to Ship",
		Description: fmt.Sprintf("**%s** has reached the fill threshold", run.Name),
		Color:       0x10b981, // Green
		Fields:      fields,
		Footer: &client.DiscordEmbedFooter{
			Text: fmt.Sprintf("Pinky.Tools • %s", time.Now().UTC().Format("Jan 2, 2006 15:04 UTC")),
		},
	}
}

func buildHaulingCompleteEmbed(run *models.HaulingRun, summary *models.HaulingRunPnlSummary) *client.DiscordEmbed {
	description := fmt.Sprintf("**%s** complete", run.Name)
	if summary != nil {
		description = fmt.Sprintf("**%s** | +%.2f ISK | %.1f%% margin", run.Name, summary.NetProfitISK, summary.MarginPct)
	}

	fields := []client.DiscordEmbedField{}
	if summary != nil {
		fields = append(fields,
			client.DiscordEmbedField{Name: "Revenue", Value: fmt.Sprintf("%.2f ISK", summary.TotalRevenueISK), Inline: true},
			client.DiscordEmbedField{Name: "Cost", Value: fmt.Sprintf("%.2f ISK", summary.TotalCostISK), Inline: true},
			client.DiscordEmbedField{Name: "Net Profit", Value: fmt.Sprintf("%.2f ISK", summary.NetProfitISK), Inline: true},
		)
	}

	return &client.DiscordEmbed{
		Title:       "Hauling Run Complete",
		Description: description,
		Color:       0x3b82f6, // Blue
		Fields:      fields,
		Footer: &client.DiscordEmbedFooter{
			Text: fmt.Sprintf("Pinky.Tools • %s", time.Now().UTC().Format("Jan 2, 2006 15:04 UTC")),
		},
	}
}

func buildHaulingItemSoldEmbed(run *models.HaulingRun, item *models.HaulingRunItem) *client.DiscordEmbed {
	fields := []client.DiscordEmbedField{
		{Name: "Item", Value: item.TypeName, Inline: true},
		{Name: "Qty Sold", Value: fmt.Sprintf("%d", item.QtySold), Inline: true},
	}
	if item.ActualRevenueISK != nil {
		fields = append(fields, client.DiscordEmbedField{
			Name:   "Revenue",
			Value:  fmt.Sprintf("%.2f ISK", *item.ActualRevenueISK),
			Inline: true,
		})
	}

	return &client.DiscordEmbed{
		Title:       "Hauling Item Sold",
		Description: fmt.Sprintf("**%s** — %s fully sold", run.Name, item.TypeName),
		Color:       0x10b981, // Green
		Fields:      fields,
		Footer: &client.DiscordEmbedFooter{
			Text: fmt.Sprintf("Pinky.Tools • %s", time.Now().UTC().Format("Jan 2, 2006 15:04 UTC")),
		},
	}
}

func buildHaulingDailyDigestEmbed(runs []*models.HaulingRun) *client.DiscordEmbed {
	fields := []client.DiscordEmbedField{}
	for _, run := range runs {
		route := fmt.Sprintf("%s → %s", regionName(run.FromRegionID), regionName(run.ToRegionID))
		value := fmt.Sprintf("Status: %s | %s", run.Status, route)
		fields = append(fields, client.DiscordEmbedField{
			Name:   run.Name,
			Value:  value,
			Inline: false,
		})
	}

	return &client.DiscordEmbed{
		Title:       "Hauling Daily Digest",
		Description: fmt.Sprintf("%d active run(s)", len(runs)),
		Color:       0xf59e0b, // Amber
		Fields:      fields,
		Footer: &client.DiscordEmbedFooter{
			Text: fmt.Sprintf("Pinky.Tools • %s", time.Now().UTC().Format("Jan 2, 2006 15:04 UTC")),
		},
	}
}
