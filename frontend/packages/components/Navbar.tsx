'use client';
import { useEffect, useState } from 'react';
import { useSession } from 'next-auth/react';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Badge from '@mui/material/Badge';
import RocketLaunchIcon from '@mui/icons-material/RocketLaunch';

type Contact = {
  id: number;
  requesterUserId: number;
  recipientUserId: number;
  requesterName: string;
  recipientName: string;
  status: 'pending' | 'accepted' | 'rejected';
  requestedAt: string;
  respondedAt?: string;
};

export default function Navbar() {
  const { data: session } = useSession();
  const [pendingCount, setPendingCount] = useState(0);

  useEffect(() => {
    if (!session?.providerAccountId) return;

    const fetchPendingCount = async () => {
      try {
        const response = await fetch('/api/contacts');
        if (!response.ok) return;

        const contacts: Contact[] | null = await response.json();
        if (!contacts) return;
        const currentUserId = parseInt(session.providerAccountId);

        // Count pending requests where current user is the recipient
        const pending = contacts.filter(
          (contact) =>
            contact.status === 'pending' &&
            contact.recipientUserId === currentUserId
        );

        setPendingCount(pending.length);
      } catch (error) {
        console.error('Failed to fetch pending contacts:', error);
      }
    };

    fetchPendingCount();

    // Poll every 30 seconds for updates
    const interval = setInterval(fetchPendingCount, 30000);

    return () => clearInterval(interval);
  }, [session]);

  return (
    <>
      <AppBar position="fixed">
        <Toolbar>
          <IconButton
            size="large"
            edge="start"
            color="inherit"
            aria-label="menu"
            sx={{ mr: 2 }}
          >
            <RocketLaunchIcon />
          </IconButton>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            EVE Industry Tool
          </Typography>
          <Button color="inherit" href="/characters">
            Characters
          </Button>
          <Button color="inherit" href="/corporations">
            Corporations
          </Button>
          <Button color="inherit" href="/inventory">
            Inventory
          </Button>
          <Button color="inherit" href="/stockpiles">
            Stockpiles
          </Button>
          <Button color="inherit" href="/contacts">
            <Badge badgeContent={pendingCount} color="error">
              Contacts
            </Badge>
          </Button>
          <Button color="inherit" href="/marketplace">
            Marketplace
          </Button>
          <Button color="inherit" href="/reactions">
            Reactions
          </Button>
          <Button color="inherit" href="/settings">
            Settings
          </Button>
        </Toolbar>
      </AppBar>
      <Toolbar />
    </>
  );
}
