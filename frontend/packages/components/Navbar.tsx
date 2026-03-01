'use client';
import { useEffect, useRef, useState } from 'react';
import { useSession } from 'next-auth/react';
import Link from 'next/link';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Badge from '@mui/material/Badge';
import Paper from '@mui/material/Paper';
import MenuList from '@mui/material/MenuList';
import MenuItem from '@mui/material/MenuItem';
import ClickAwayListener from '@mui/material/ClickAwayListener';
import Popper from '@mui/material/Popper';
import Grow from '@mui/material/Grow';
import RocketLaunchIcon from '@mui/icons-material/RocketLaunch';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import Alert from '@mui/material/Alert';

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

type NavItem = { label: string; href: string; badge?: number };

type DropdownProps = {
  label: React.ReactNode;
  items: NavItem[];
};

function NavDropdown({ label, items }: DropdownProps) {
  const [open, setOpen] = useState(false);
  const anchorRef = useRef<HTMLButtonElement>(null);

  const handleToggle = () => setOpen((prev) => !prev);
  const handleClose = () => setOpen(false);

  return (
    <>
      <Button
        ref={anchorRef}
        color="inherit"
        onClick={handleToggle}
        endIcon={<KeyboardArrowDownIcon />}
        aria-haspopup="true"
        aria-expanded={open}
      >
        {label}
      </Button>
      <Popper
        open={open}
        anchorEl={anchorRef.current}
        placement="bottom-start"
        transition
        style={{ zIndex: 1300 }}
      >
        {({ TransitionProps }) => (
          <Grow {...TransitionProps}>
            <Paper>
              <ClickAwayListener onClickAway={handleClose}>
                <MenuList>
                  {items.map((item) => (
                    <MenuItem
                      key={item.href}
                      component="a"
                      href={item.href}
                      onClick={handleClose}
                    >
                      {item.badge ? (
                        <Badge badgeContent={item.badge} color="error">
                          {item.label}
                        </Badge>
                      ) : (
                        item.label
                      )}
                    </MenuItem>
                  ))}
                </MenuList>
              </ClickAwayListener>
            </Paper>
          </Grow>
        )}
      </Popper>
    </>
  );
}

export default function Navbar() {
  const { data: session } = useSession();
  const [pendingCount, setPendingCount] = useState(0);
  const [scopeWarning, setScopeWarning] = useState(false);

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

  useEffect(() => {
    if (!session?.providerAccountId) return;

    const fetchScopeStatus = async () => {
      try {
        const response = await fetch('/api/scope-status');
        if (!response.ok) return;
        const data = await response.json();
        setScopeWarning(data.hasOutdated);
      } catch (error) {
        console.error('Failed to fetch scope status:', error);
      }
    };

    fetchScopeStatus();
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
            href="/"
            sx={{ mr: 2 }}
          >
            <RocketLaunchIcon />
          </IconButton>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            <Link href="/" style={{ textDecoration: 'none', color: 'inherit' }}>
              EVE Industry Tool
            </Link>
          </Typography>

          <NavDropdown
            label={
              scopeWarning ? (
                <Badge variant="dot" color="warning">
                  Account
                </Badge>
              ) : (
                "Account"
              )
            }
            items={[
              { label: 'Characters', href: '/characters' },
              { label: 'Corporations', href: '/corporations' },
            ]}
          />

          <NavDropdown
            label="Assets"
            items={[
              { label: 'Inventory', href: '/inventory' },
              { label: 'Stockpiles', href: '/stockpiles' },
            ]}
          />

          {/* Trading — badge on top-level button */}
          <NavDropdown
            label={
              <Badge badgeContent={pendingCount} color="error">
                Trading
              </Badge>
            }
            items={[
              { label: 'Contacts', href: '/contacts', badge: pendingCount },
              { label: 'Marketplace', href: '/marketplace' },
            ]}
          />

          <NavDropdown
            label="Industry"
            items={[
              { label: 'Reactions', href: '/reactions' },
              { label: 'Industry', href: '/industry' },
              { label: 'Plans', href: '/production-plans' },
              { label: 'Runs', href: '/plan-runs' },
              { label: 'Planets', href: '/pi' },
              { label: 'Job Slots', href: '/job-slots' },
            ]}
          />

          <NavDropdown
            label="Logistics"
            items={[
              { label: 'Transport', href: '/transport' },
              { label: 'Stations', href: '/stations' },
              { label: 'Hauling Runs', href: '/hauling' },
              { label: 'Market Scanner', href: '/hauling/scanner' },
            ]}
          />

          {/* Settings — standalone */}
          <Button color="inherit" href="/settings">
            Settings
          </Button>
        </Toolbar>
      </AppBar>
      <Toolbar />
      {scopeWarning && (
        <Alert
          severity="warning"
          sx={{ borderRadius: 0 }}
          action={
            <Button color="inherit" size="small" href="/characters">
              View
            </Button>
          }
        >
          Some characters or corporations need to be re-authorized with updated permissions.
        </Alert>
      )}
    </>
  );
}
