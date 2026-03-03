import { useState, useEffect, useCallback } from 'react';
import { Bell, Link2, Link2Off, Trash2, Send, Plus, Loader2 } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Switch } from '@/components/ui/switch';
import { Separator } from '@/components/ui/separator';
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { toast } from '@/components/ui/sonner';

type DiscordLink = {
  linked: boolean;
  discordUserId?: string;
  discordUsername?: string;
  linkedAt?: string;
};

type NotificationTarget = {
  id: number;
  userId: number;
  targetType: string;
  channelId?: string;
  guildName: string;
  channelName: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

type NotificationPreference = {
  id: number;
  targetId: number;
  eventType: string;
  isEnabled: boolean;
};

type Guild = {
  id: string;
  name: string;
  icon: string;
};

type Channel = {
  id: string;
  name: string;
  type: number;
};

const EVENT_TYPES = [
  { value: 'purchase_created', label: 'New Purchase' },
  { value: 'contract_created', label: 'Contract Created' },
  { value: 'pi_stall', label: 'PI Stall Alert' },
];

const DISCORD_ERROR_MESSAGES: Record<string, string> = {
  discord_dm_disabled: 'Cannot send DMs to this user. They must be in a server where the bot is a member, and have "Allow direct messages from server members" enabled in that server\'s privacy settings.',
  discord_no_channel_access: 'The bot does not have access to this channel. Check the bot\'s permissions in your Discord server.',
  discord_missing_permissions: 'The bot is missing permissions to send messages in this channel.',
  discord_unknown_channel: 'Unknown channel. The channel may have been deleted.',
  discord_unknown_error: 'An unknown Discord error occurred. Please try again later.',
  discord_send_failed: 'Failed to send test notification. Please try again later.',
};

export default function DiscordSettings() {
  const [link, setLink] = useState<DiscordLink | null>(null);
  const [targets, setTargets] = useState<NotificationTarget[]>([]);
  const [preferences, setPreferences] = useState<Record<number, NotificationPreference[]>>({});
  const [loading, setLoading] = useState(true);

  // Channel picker dialog
  const [channelDialogOpen, setChannelDialogOpen] = useState(false);
  const [guilds, setGuilds] = useState<Guild[]>([]);
  const [channels, setChannels] = useState<Channel[]>([]);
  const [selectedGuild, setSelectedGuild] = useState<Guild | null>(null);
  const [loadingGuilds, setLoadingGuilds] = useState(false);
  const [loadingChannels, setLoadingChannels] = useState(false);

  const fetchLink = useCallback(async () => {
    try {
      const res = await fetch('/api/discord/link');
      if (res.ok) {
        setLink(await res.json());
      }
    } catch {
      // Link fetch failed silently
    }
  }, []);

  const fetchTargets = useCallback(async () => {
    try {
      const res = await fetch('/api/discord/targets');
      if (res.ok) {
        const data: NotificationTarget[] = await res.json();
        setTargets(data);
        for (const target of data) {
          fetchPreferences(target.id);
        }
      }
    } catch {
      // Targets fetch failed silently
    }
  }, []);

  const fetchPreferences = async (targetId: number) => {
    try {
      const res = await fetch(`/api/discord/targets/${targetId}/preferences`);
      if (res.ok) {
        const data: NotificationPreference[] = await res.json();
        setPreferences(prev => ({ ...prev, [targetId]: data }));
      }
    } catch {
      // Preferences fetch failed silently
    }
  };

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      await fetchLink();
      await fetchTargets();
      setLoading(false);
    };
    load();
  }, [fetchLink, fetchTargets]);

  const handleUnlink = async () => {
    try {
      const res = await fetch('/api/discord/link', { method: 'DELETE' });
      if (res.ok) {
        setLink({ linked: false });
        setTargets([]);
        setPreferences({});
        toast.success('Discord account unlinked');
      } else {
        toast.error('Failed to unlink Discord account');
      }
    } catch {
      toast.error('Failed to unlink Discord account');
    }
  };

  const handleAddDMTarget = async () => {
    try {
      const res = await fetch('/api/discord/targets', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ targetType: 'dm' }),
      });
      if (res.ok) {
        toast.success('DM target added');
        await fetchTargets();
      } else {
        toast.error('Failed to add DM target');
      }
    } catch {
      toast.error('Failed to add DM target');
    }
  };

  const handleOpenChannelDialog = async () => {
    setChannelDialogOpen(true);
    setLoadingGuilds(true);
    setSelectedGuild(null);
    setChannels([]);
    try {
      const res = await fetch('/api/discord/guilds');
      if (res.ok) {
        setGuilds(await res.json());
      }
    } catch {
      toast.error('Failed to load servers');
    }
    setLoadingGuilds(false);
  };

  const handleSelectGuild = async (guild: Guild) => {
    setSelectedGuild(guild);
    setLoadingChannels(true);
    try {
      const res = await fetch(`/api/discord/guilds/${guild.id}/channels`);
      if (res.ok) {
        setChannels(await res.json());
      }
    } catch {
      toast.error('Failed to load channels');
    }
    setLoadingChannels(false);
  };

  const handleSelectChannel = async (channel: Channel) => {
    if (!selectedGuild) return;
    try {
      const res = await fetch('/api/discord/targets', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          targetType: 'channel',
          channelId: channel.id,
          guildName: selectedGuild.name,
          channelName: channel.name,
        }),
      });
      if (res.ok) {
        setChannelDialogOpen(false);
        toast.success(`Channel #${channel.name} added`);
        await fetchTargets();
      } else {
        toast.error('Failed to add channel target');
      }
    } catch {
      toast.error('Failed to add channel target');
    }
  };

  const handleToggleTarget = async (target: NotificationTarget) => {
    try {
      const res = await fetch(`/api/discord/targets/${target.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ isActive: !target.isActive }),
      });
      if (res.ok) {
        setTargets(prev => prev.map(t => t.id === target.id ? { ...t, isActive: !t.isActive } : t));
      }
    } catch {
      toast.error('Failed to update target');
    }
  };

  const handleDeleteTarget = async (targetId: number) => {
    try {
      const res = await fetch(`/api/discord/targets/${targetId}`, { method: 'DELETE' });
      if (res.ok) {
        setTargets(prev => prev.filter(t => t.id !== targetId));
        toast.success('Target removed');
      } else {
        toast.error('Failed to remove target');
      }
    } catch {
      toast.error('Failed to remove target');
    }
  };

  const handleTestTarget = async (targetId: number) => {
    try {
      const res = await fetch(`/api/discord/targets/${targetId}/test`, { method: 'POST' });
      if (res.ok) {
        toast.success('Test notification sent!');
      } else {
        const data = await res.json().catch(() => null);
        const errorCode = data?.error || '';
        const message = DISCORD_ERROR_MESSAGES[errorCode] || 'Failed to send test notification';
        toast.error(message);
      }
    } catch {
      toast.error('Failed to send test notification');
    }
  };

  const handleTogglePreference = async (targetId: number, eventType: string, currentlyEnabled: boolean) => {
    try {
      const res = await fetch(`/api/discord/targets/${targetId}/preferences`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ eventType, isEnabled: !currentlyEnabled }),
      });
      if (res.ok) {
        await fetchPreferences(targetId);
      }
    } catch {
      toast.error('Failed to update preference');
    }
  };

  const isPreferenceEnabled = (targetId: number, eventType: string): boolean => {
    const prefs = preferences[targetId] || [];
    const pref = prefs.find(p => p.eventType === eventType);
    return pref?.isEnabled ?? false;
  };

  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <Loader2 className="h-8 w-8 animate-spin text-[var(--color-primary-cyan)]" />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Discord Account Section */}
      <Card>
        <CardContent className="p-6">
          <div className="flex items-center gap-2 mb-4">
            <Bell className="h-5 w-5 text-[#5865F2]" />
            <h3 className="text-lg font-semibold text-[var(--color-text-emphasis)]">Discord Notifications</h3>
          </div>

          {!link?.linked ? (
            <div>
              <p className="text-sm text-[var(--color-text-secondary)] mb-4">
                Link your Discord account to receive notifications when marketplace events occur.
              </p>
              <Button asChild className="bg-[#5865F2] hover:bg-[#4752C4] text-white">
                <a href="/api/discord/login">
                  <Link2 className="h-4 w-4 mr-2" />
                  Link Discord Account
                </a>
              </Button>
            </div>
          ) : (
            <div className="flex items-center gap-3">
              <Badge className="bg-[#5865F2] text-white border-[#5865F2]">
                {link.discordUsername}
              </Badge>
              <Button variant="outline" size="sm" onClick={handleUnlink} className="text-[var(--color-danger-rose)] border-[var(--color-danger-rose)]/30 hover:bg-[var(--color-danger-rose)]/10">
                <Link2Off className="h-4 w-4 mr-1" />
                Unlink
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Notification Targets Section */}
      {link?.linked && (
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-[var(--color-text-emphasis)]">Notification Targets</h3>
              <div className="flex gap-2">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleAddDMTarget}
                  disabled={targets.some(t => t.targetType === 'dm')}
                >
                  <Plus className="h-4 w-4 mr-1" />
                  Add DM
                </Button>
                <Button variant="ghost" size="sm" onClick={handleOpenChannelDialog}>
                  <Plus className="h-4 w-4 mr-1" />
                  Add Channel
                </Button>
              </div>
            </div>

            {targets.length === 0 ? (
              <p className="text-sm text-[var(--color-text-secondary)]">
                No notification targets configured. Add a DM or channel target to start receiving notifications.
              </p>
            ) : (
              <div className="space-y-3">
                {targets.map(target => (
                  <Card key={target.id} className="bg-[var(--color-bg-void)]">
                    <CardContent className="p-4">
                      <div className="flex items-center justify-between mb-2">
                        <div className="flex items-center gap-2">
                          <Badge variant={target.isActive ? 'default' : 'secondary'}>
                            {target.targetType === 'dm' ? 'Direct Message' : `#${target.channelName}`}
                          </Badge>
                          {target.targetType === 'channel' && target.guildName && (
                            <span className="text-sm text-[var(--color-text-secondary)]">
                              {target.guildName}
                            </span>
                          )}
                        </div>
                        <div className="flex items-center gap-1">
                          <Switch
                            checked={target.isActive}
                            onCheckedChange={() => handleToggleTarget(target)}
                          />
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleTestTarget(target.id)}
                            title="Send test notification"
                          >
                            <Send className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleDeleteTarget(target.id)}
                            title="Remove target"
                            className="text-[var(--color-danger-rose)] hover:text-[var(--color-danger-rose)]"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>

                      <Separator className="my-2" />

                      <p className="text-xs text-[var(--color-text-muted)] uppercase tracking-wider mb-2">
                        Event Types
                      </p>
                      {EVENT_TYPES.map(evt => (
                        <div key={evt.value} className="flex items-center justify-between py-1">
                          <span className="text-sm text-[var(--color-text-primary)]">{evt.label}</span>
                          <Switch
                            checked={isPreferenceEnabled(target.id, evt.value)}
                            onCheckedChange={() => handleTogglePreference(target.id, evt.value, isPreferenceEnabled(target.id, evt.value))}
                          />
                        </div>
                      ))}
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Channel Picker Dialog */}
      <Dialog open={channelDialogOpen} onOpenChange={setChannelDialogOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>
              {selectedGuild ? `Select Channel in ${selectedGuild.name}` : 'Select Server'}
            </DialogTitle>
          </DialogHeader>

          {!selectedGuild ? (
            loadingGuilds ? (
              <div className="flex justify-center py-6">
                <Loader2 className="h-8 w-8 animate-spin text-[var(--color-primary-cyan)]" />
              </div>
            ) : guilds.length === 0 ? (
              <p className="text-sm text-[var(--color-text-secondary)] py-4">
                No shared servers found. Make sure the Pinky.Tools bot is added to one of your Discord servers.
              </p>
            ) : (
              <div className="space-y-1">
                {guilds.map(guild => (
                  <button
                    key={guild.id}
                    onClick={() => handleSelectGuild(guild)}
                    className="w-full text-left px-3 py-2 rounded-sm text-sm text-[var(--color-text-primary)] hover:bg-[var(--color-surface-elevated)] transition-colors cursor-pointer"
                  >
                    {guild.name}
                  </button>
                ))}
              </div>
            )
          ) : (
            loadingChannels ? (
              <div className="flex justify-center py-6">
                <Loader2 className="h-8 w-8 animate-spin text-[var(--color-primary-cyan)]" />
              </div>
            ) : channels.length === 0 ? (
              <p className="text-sm text-[var(--color-text-secondary)] py-4">
                No text channels found in this server that the bot can access.
              </p>
            ) : (
              <div className="space-y-1">
                {channels.map(channel => (
                  <button
                    key={channel.id}
                    onClick={() => handleSelectChannel(channel)}
                    className="w-full text-left px-3 py-2 rounded-sm text-sm text-[var(--color-text-primary)] hover:bg-[var(--color-surface-elevated)] transition-colors cursor-pointer"
                  >
                    # {channel.name}
                  </button>
                ))}
              </div>
            )
          )}

          <DialogFooter>
            {selectedGuild && (
              <Button variant="ghost" onClick={() => { setSelectedGuild(null); setChannels([]); }}>
                Back
              </Button>
            )}
            <Button variant="ghost" onClick={() => setChannelDialogOpen(false)}>Cancel</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
