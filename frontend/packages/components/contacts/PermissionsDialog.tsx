import { useState, useEffect } from 'react';
import { Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from '@/components/ui/dialog';

type Contact = {
  id: number;
  requesterUserId: number;
  recipientUserId: number;
  requesterName: string;
  recipientName: string;
  status: string;
};

type ContactPermission = {
  id: number;
  contactId: number;
  grantingUserId: number;
  receivingUserId: number;
  serviceType: string;
  canAccess: boolean;
};

type PermissionsDialogProps = {
  open: boolean;
  onClose: () => void;
  contact: Contact;
  currentUserId: number;
};

const SERVICE_TYPES = [
  { type: 'for_sale_browse', label: 'Browse For-Sale Items' },
  { type: 'job_slot_browse', label: 'Browse Job Slot Listings' },
];

export default function PermissionsDialog({
  open,
  onClose,
  contact,
  currentUserId,
}: PermissionsDialogProps) {
  const [permissions, setPermissions] = useState<ContactPermission[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const otherUserId = contact.requesterUserId === currentUserId
    ? contact.recipientUserId
    : contact.requesterUserId;

  const otherUserName = contact.requesterUserId === currentUserId
    ? contact.recipientName
    : contact.requesterName;

  useEffect(() => {
    if (open) {
      fetchPermissions();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, contact.id]);

  const fetchPermissions = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch(`/api/contacts/${contact.id}/permissions`);
      if (response.ok) {
        const data: ContactPermission[] = await response.json();
        setPermissions(data || []);
      } else {
        setError('Failed to load permissions');
      }
    } catch {
      setError('Failed to load permissions');
    } finally {
      setLoading(false);
    }
  };

  const handleTogglePermission = async (serviceType: string, receivingUserId: number, currentValue: boolean) => {
    setSaving(true);
    setError(null);
    try {
      const response = await fetch(`/api/contacts/${contact.id}/permissions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          serviceType,
          receivingUserId,
          canAccess: !currentValue,
        }),
      });

      if (response.ok) {
        await fetchPermissions();
      } else {
        const errorData = await response.json();
        setError(errorData.error || 'Failed to update permission');
      }
    } catch {
      setError('Failed to update permission');
    } finally {
      setSaving(false);
    }
  };

  const getPermission = (grantingUserId: number, receivingUserId: number, serviceType: string): boolean => {
    const perm = permissions.find(
      p => p.grantingUserId === grantingUserId &&
           p.receivingUserId === receivingUserId &&
           p.serviceType === serviceType
    );
    return perm?.canAccess || false;
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => { if (!isOpen) onClose(); }}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Manage Permissions — {otherUserName}</DialogTitle>
        </DialogHeader>

        {loading ? (
          <div className="flex justify-center py-8" role="status">
            <Loader2 className="h-8 w-8 animate-spin text-[var(--color-primary-cyan)]" />
          </div>
        ) : (
          <div className="space-y-6">
            {error && (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            <p className="text-sm text-[var(--color-text-secondary)]">
              Control what services you allow this contact to access. Permissions are unidirectional — you control what they can access from you, and vice versa.
            </p>

            {/* Permissions I Grant to Them */}
            <div>
              <h4 className="text-sm font-semibold text-[var(--color-text-emphasis)] mb-1">
                Permissions I Grant to {otherUserName}
              </h4>
              <p className="text-xs text-[var(--color-text-secondary)] mb-3">
                What {otherUserName} can access from you:
              </p>
              {SERVICE_TYPES.map((service) => {
                const granted = getPermission(currentUserId, otherUserId, service.type);
                return (
                  <div key={`grant-${service.type}`} className="flex items-center justify-between py-1.5">
                    <Label className="cursor-pointer">{service.label}</Label>
                    <Switch
                      checked={granted}
                      onCheckedChange={() => handleTogglePermission(service.type, otherUserId, granted)}
                      disabled={saving}
                    />
                  </div>
                );
              })}
            </div>

            <Separator />

            {/* Permissions They Grant to Me */}
            <div>
              <h4 className="text-sm font-semibold text-[var(--color-text-emphasis)] mb-1">
                Permissions {otherUserName} Grants to Me
              </h4>
              <p className="text-xs text-[var(--color-text-secondary)] mb-3">
                What you can access from {otherUserName}:
              </p>
              {SERVICE_TYPES.map((service) => {
                const granted = getPermission(otherUserId, currentUserId, service.type);
                return (
                  <div key={`receive-${service.type}`} className="flex items-center justify-between py-1.5">
                    <Label>{service.label}</Label>
                    <Switch checked={granted} disabled />
                  </div>
                );
              })}
            </div>
          </div>
        )}

        <DialogFooter>
          <Button variant="ghost" onClick={onClose}>Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
