import { useState, useEffect, useCallback } from 'react';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Switch from '@mui/material/Switch';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';
import CircularProgress from '@mui/material/CircularProgress';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import Divider from '@mui/material/Divider';
import DeleteIcon from '@mui/icons-material/Delete';
import SendIcon from '@mui/icons-material/Send';
import LinkIcon from '@mui/icons-material/Link';
import LinkOffIcon from '@mui/icons-material/LinkOff';
import AddIcon from '@mui/icons-material/Add';
import NotificationsIcon from '@mui/icons-material/Notifications';

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
];

export default function DiscordSettings() {
  const [link, setLink] = useState<DiscordLink | null>(null);
  const [targets, setTargets] = useState<NotificationTarget[]>([]);
  const [preferences, setPreferences] = useState<Record<number, NotificationPreference[]>>({});
  const [loading, setLoading] = useState(true);
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' }>({ open: false, message: '', severity: 'success' });

  // Channel picker dialog
  const [channelDialogOpen, setChannelDialogOpen] = useState(false);
  const [guilds, setGuilds] = useState<Guild[]>([]);
  const [channels, setChannels] = useState<Channel[]>([]);
  const [selectedGuild, setSelectedGuild] = useState<Guild | null>(null);
  const [loadingGuilds, setLoadingGuilds] = useState(false);
  const [loadingChannels, setLoadingChannels] = useState(false);

  const showSnackbar = (message: string, severity: 'success' | 'error') => {
    setSnackbar({ open: true, message, severity });
  };

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
        // Fetch preferences for each target
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
        showSnackbar('Discord account unlinked', 'success');
      } else {
        showSnackbar('Failed to unlink Discord account', 'error');
      }
    } catch {
      showSnackbar('Failed to unlink Discord account', 'error');
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
        showSnackbar('DM target added', 'success');
        await fetchTargets();
      } else {
        showSnackbar('Failed to add DM target', 'error');
      }
    } catch {
      showSnackbar('Failed to add DM target', 'error');
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
      showSnackbar('Failed to load servers', 'error');
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
      showSnackbar('Failed to load channels', 'error');
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
        showSnackbar(`Channel #${channel.name} added`, 'success');
        await fetchTargets();
      } else {
        showSnackbar('Failed to add channel target', 'error');
      }
    } catch {
      showSnackbar('Failed to add channel target', 'error');
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
      showSnackbar('Failed to update target', 'error');
    }
  };

  const handleDeleteTarget = async (targetId: number) => {
    try {
      const res = await fetch(`/api/discord/targets/${targetId}`, { method: 'DELETE' });
      if (res.ok) {
        setTargets(prev => prev.filter(t => t.id !== targetId));
        showSnackbar('Target removed', 'success');
      } else {
        showSnackbar('Failed to remove target', 'error');
      }
    } catch {
      showSnackbar('Failed to remove target', 'error');
    }
  };

  const handleTestTarget = async (targetId: number) => {
    try {
      const res = await fetch(`/api/discord/targets/${targetId}/test`, { method: 'POST' });
      if (res.ok) {
        showSnackbar('Test notification sent!', 'success');
      } else {
        showSnackbar('Failed to send test notification', 'error');
      }
    } catch {
      showSnackbar('Failed to send test notification', 'error');
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
      showSnackbar('Failed to update preference', 'error');
    }
  };

  const isPreferenceEnabled = (targetId: number, eventType: string): boolean => {
    const prefs = preferences[targetId] || [];
    const pref = prefs.find(p => p.eventType === eventType);
    return pref?.isEnabled ?? false;
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      {/* Discord Account Section */}
      <Card sx={{ mb: 3, bgcolor: '#12151f' }}>
        <CardContent>
          <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
            <NotificationsIcon sx={{ mr: 1, color: '#5865F2' }} />
            <Typography variant="h6">Discord Notifications</Typography>
          </Box>

          {!link?.linked ? (
            <Box>
              <Typography variant="body2" sx={{ mb: 2, color: 'text.secondary' }}>
                Link your Discord account to receive notifications when marketplace events occur.
              </Typography>
              <Button
                variant="contained"
                startIcon={<LinkIcon />}
                href="/api/discord/login"
                sx={{ bgcolor: '#5865F2', '&:hover': { bgcolor: '#4752C4' } }}
              >
                Link Discord Account
              </Button>
            </Box>
          ) : (
            <Box>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                <Chip
                  label={link.discordUsername}
                  color="primary"
                  sx={{ mr: 2, bgcolor: '#5865F2' }}
                />
                <Button
                  variant="outlined"
                  size="small"
                  startIcon={<LinkOffIcon />}
                  onClick={handleUnlink}
                  color="error"
                >
                  Unlink
                </Button>
              </Box>
            </Box>
          )}
        </CardContent>
      </Card>

      {/* Notification Targets Section */}
      {link?.linked && (
        <Card sx={{ mb: 3, bgcolor: '#12151f' }}>
          <CardContent>
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
              <Typography variant="h6">Notification Targets</Typography>
              <Box>
                <Button
                  size="small"
                  startIcon={<AddIcon />}
                  onClick={handleAddDMTarget}
                  sx={{ mr: 1 }}
                  disabled={targets.some(t => t.targetType === 'dm')}
                >
                  Add DM
                </Button>
                <Button
                  size="small"
                  startIcon={<AddIcon />}
                  onClick={handleOpenChannelDialog}
                >
                  Add Channel
                </Button>
              </Box>
            </Box>

            {targets.length === 0 ? (
              <Typography variant="body2" sx={{ color: 'text.secondary' }}>
                No notification targets configured. Add a DM or channel target to start receiving notifications.
              </Typography>
            ) : (
              targets.map(target => (
                <Card key={target.id} sx={{ mb: 2, bgcolor: '#0f1219', border: '1px solid #1e2130' }}>
                  <CardContent sx={{ pb: '16px !important' }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 1 }}>
                      <Box sx={{ display: 'flex', alignItems: 'center' }}>
                        <Chip
                          label={target.targetType === 'dm' ? 'Direct Message' : `#${target.channelName}`}
                          size="small"
                          sx={{ mr: 1 }}
                          color={target.isActive ? 'primary' : 'default'}
                        />
                        {target.targetType === 'channel' && target.guildName && (
                          <Typography variant="body2" sx={{ color: 'text.secondary' }}>
                            {target.guildName}
                          </Typography>
                        )}
                      </Box>
                      <Box>
                        <Switch
                          checked={target.isActive}
                          onChange={() => handleToggleTarget(target)}
                          size="small"
                        />
                        <IconButton
                          size="small"
                          onClick={() => handleTestTarget(target.id)}
                          title="Send test notification"
                        >
                          <SendIcon fontSize="small" />
                        </IconButton>
                        <IconButton
                          size="small"
                          onClick={() => handleDeleteTarget(target.id)}
                          color="error"
                          title="Remove target"
                        >
                          <DeleteIcon fontSize="small" />
                        </IconButton>
                      </Box>
                    </Box>

                    <Divider sx={{ my: 1 }} />

                    <Typography variant="caption" sx={{ color: 'text.secondary', display: 'block', mb: 1 }}>
                      Event Types
                    </Typography>
                    {EVENT_TYPES.map(evt => (
                      <Box key={evt.value} sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                        <Typography variant="body2">{evt.label}</Typography>
                        <Switch
                          checked={isPreferenceEnabled(target.id, evt.value)}
                          onChange={() => handleTogglePreference(target.id, evt.value, isPreferenceEnabled(target.id, evt.value))}
                          size="small"
                        />
                      </Box>
                    ))}
                  </CardContent>
                </Card>
              ))
            )}
          </CardContent>
        </Card>
      )}

      {/* Channel Picker Dialog */}
      <Dialog open={channelDialogOpen} onClose={() => setChannelDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>
          {selectedGuild ? `Select Channel in ${selectedGuild.name}` : 'Select Server'}
        </DialogTitle>
        <DialogContent>
          {!selectedGuild ? (
            loadingGuilds ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', py: 3 }}>
                <CircularProgress />
              </Box>
            ) : guilds.length === 0 ? (
              <Typography variant="body2" sx={{ color: 'text.secondary', py: 2 }}>
                No shared servers found. Make sure the Pinky.Tools bot is added to one of your Discord servers.
              </Typography>
            ) : (
              <List>
                {guilds.map(guild => (
                  <ListItem
                    key={guild.id}
                    onClick={() => handleSelectGuild(guild)}
                    sx={{ cursor: 'pointer', '&:hover': { bgcolor: '#1e2130' }, borderRadius: 1 }}
                  >
                    <ListItemText primary={guild.name} />
                  </ListItem>
                ))}
              </List>
            )
          ) : (
            loadingChannels ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', py: 3 }}>
                <CircularProgress />
              </Box>
            ) : channels.length === 0 ? (
              <Typography variant="body2" sx={{ color: 'text.secondary', py: 2 }}>
                No text channels found in this server that the bot can access.
              </Typography>
            ) : (
              <List>
                {channels.map(channel => (
                  <ListItem
                    key={channel.id}
                    onClick={() => handleSelectChannel(channel)}
                    sx={{ cursor: 'pointer', '&:hover': { bgcolor: '#1e2130' }, borderRadius: 1 }}
                  >
                    <ListItemText primary={`# ${channel.name}`} />
                  </ListItem>
                ))}
              </List>
            )
          )}
        </DialogContent>
        <DialogActions>
          {selectedGuild && (
            <Button onClick={() => { setSelectedGuild(null); setChannels([]); }}>
              Back
            </Button>
          )}
          <Button onClick={() => setChannelDialogOpen(false)}>Cancel</Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={4000}
        onClose={() => setSnackbar(prev => ({ ...prev, open: false }))}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert
          onClose={() => setSnackbar(prev => ({ ...prev, open: false }))}
          severity={snackbar.severity}
          variant="filled"
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
}
