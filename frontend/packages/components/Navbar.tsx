'use client';
import { useEffect, useState } from 'react';
import { usePathname } from 'next/navigation';
import { useSession } from 'next-auth/react';
import Link from 'next/link';
import { Rocket, ChevronDown, AlertTriangle, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@/components/ui/dropdown-menu';

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

type NavDropdownProps = {
  label: React.ReactNode;
  items: NavItem[];
  pathname?: string | null;
};

function NavDropdown({ label, items, pathname }: NavDropdownProps) {
  const isActive = items.some(
    (item) => item.href !== '/' && pathname?.startsWith(item.href)
  );

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          className={`flex items-center gap-1 px-3 py-2 text-sm font-medium hover:text-[var(--color-primary-cyan)] transition-colors ${
            isActive
              ? 'text-[var(--color-primary-cyan)]'
              : 'text-[var(--color-text-primary)]'
          }`}
        >
          {label}
          <ChevronDown className="h-4 w-4 opacity-60" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start">
        {items.map((item) => {
          const itemActive = item.href !== '/' && pathname?.startsWith(item.href);
          return (
            <DropdownMenuItem key={item.href} asChild>
              <a
                href={item.href}
                className={`flex items-center gap-2 ${
                  itemActive ? 'text-[var(--color-primary-cyan)] font-medium' : ''
                }`}
              >
                {item.label}
                {item.badge ? (
                  <Badge variant="destructive" className="ml-auto text-[10px] px-1.5 py-0">
                    {item.badge}
                  </Badge>
                ) : null}
              </a>
            </DropdownMenuItem>
          );
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export default function Navbar() {
  const { data: session } = useSession();
  const pathname = usePathname();
  const [pendingCount, setPendingCount] = useState(0);
  const [scopeWarning, setScopeWarning] = useState(false);
  const [scopeDismissed, setScopeDismissed] = useState(false);

  useEffect(() => {
    if (!session?.providerAccountId) return;

    const fetchPendingCount = async () => {
      try {
        const response = await fetch('/api/contacts');
        if (!response.ok) return;

        const contacts: Contact[] | null = await response.json();
        if (!contacts) return;
        const currentUserId = parseInt(session.providerAccountId);

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
      <nav className="fixed top-0 left-0 right-0 z-50 h-16 border-b border-[var(--color-border-dim)] bg-[var(--color-bg-panel)]/95 backdrop-blur-sm">
        <div className="flex h-full items-center px-4 gap-1">
          {/* Logo */}
          <Link href="/" className="flex items-center gap-2 mr-4 text-[var(--color-primary-cyan)] hover:text-[var(--color-primary-cyan)] transition-colors">
            <Rocket className="h-5 w-5" />
            <span className="font-display font-semibold text-lg hidden sm:inline">pinky.tools</span>
          </Link>

          {/* Nav Items */}
          <NavDropdown
            label={
              scopeWarning ? (
                <span className="relative">
                  Account
                  <span className="absolute -top-1 -right-2 h-2 w-2 rounded-full bg-[var(--color-manufacturing-amber)]" />
                </span>
              ) : (
                'Account'
              )
            }
            items={[
              { label: 'Characters', href: '/characters' },
              { label: 'Corporations', href: '/corporations' },
            ]}
            pathname={pathname}
          />

          <NavDropdown
            label={
              <span className="flex items-center gap-1.5">
                <img src="/icons/assets.png" alt="" className="h-4 w-4 object-contain" />
                Assets
              </span>
            }
            items={[
              { label: 'Inventory', href: '/inventory' },
              { label: 'Stockpiles', href: '/stockpiles' },
            ]}
            pathname={pathname}
          />

          <NavDropdown
            label={
              pendingCount > 0 ? (
                <span className="flex items-center gap-1.5">
                  <img src="/icons/trading.png" alt="" className="h-4 w-4 object-contain" />
                  Trading
                  <Badge variant="destructive" className="text-[10px] px-1.5 py-0">
                    {pendingCount}
                  </Badge>
                </span>
              ) : (
                <span className="flex items-center gap-1.5">
                  <img src="/icons/trading.png" alt="" className="h-4 w-4 object-contain" />
                  Trading
                </span>
              )
            }
            items={[
              { label: 'Contacts', href: '/contacts', badge: pendingCount },
              { label: 'Marketplace', href: '/marketplace' },
            ]}
            pathname={pathname}
          />

          <NavDropdown
            label={
              <span className="flex items-center gap-1.5">
                <img src="/icons/industry.png" alt="" className="h-4 w-4 object-contain" />
                Industry
              </span>
            }
            items={[
              { label: 'Reactions', href: '/reactions' },
              { label: 'Industry', href: '/industry' },
              { label: 'Plans', href: '/production-plans' },
              { label: 'Runs', href: '/plan-runs' },
              { label: 'Planets', href: '/pi' },
              { label: 'Job Slots', href: '/job-slots' },
            ]}
            pathname={pathname}
          />

          <NavDropdown
            label={
              <span className="flex items-center gap-1.5">
                <img src="/icons/logistics.png" alt="" className="h-4 w-4 object-contain" />
                Logistics
              </span>
            }
            items={[
              { label: 'Transport', href: '/transport' },
              { label: 'Stations', href: '/stations' },
              { label: 'Hauling Runs', href: '/hauling' },
              { label: 'Market Scanner', href: '/hauling/scanner' },
            ]}
            pathname={pathname}
          />

          <Link
            href="/settings"
            className={`flex items-center px-3 py-2 text-sm font-medium text-muted-foreground hover:text-[var(--color-primary-cyan)] transition-colors ${
              pathname === '/settings'
                ? 'text-[var(--color-primary-cyan)]'
                : 'text-[var(--color-text-primary)]'
            }`}
          >
            Settings
          </Link>
        </div>
      </nav>

      {/* Spacer for fixed navbar */}
      <div className="h-16" />

      {/* Scope warning banner */}
      {scopeWarning && !scopeDismissed && (
        <Alert className="rounded-none border-x-0 border-t-0 border-[var(--color-manufacturing-amber)]/30 bg-[var(--color-manufacturing-amber)]/10">
          <AlertTriangle className="h-4 w-4 text-[var(--color-manufacturing-amber)]" />
          <AlertDescription className="flex items-center justify-between w-full">
            <span className="text-[var(--color-manufacturing-amber)]">
              Some characters or corporations need to be re-authorized with updated permissions.
            </span>
            <div className="flex items-center gap-1">
              <Button variant="ghost" size="sm" asChild className="text-[var(--color-manufacturing-amber)] hover:text-[var(--color-manufacturing-amber)]">
                <a href="/characters">View</a>
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setScopeDismissed(true)}
                className="text-[var(--color-manufacturing-amber)] hover:text-[var(--color-manufacturing-amber)] h-6 w-6 p-0"
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
          </AlertDescription>
        </Alert>
      )}
    </>
  );
}
