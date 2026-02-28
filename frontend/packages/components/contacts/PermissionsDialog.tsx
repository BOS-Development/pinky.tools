import { useState, useEffect } from 'react';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Divider from '@mui/material/Divider';
import Alert from '@mui/material/Alert';
import CircularProgress from '@mui/material/CircularProgress';

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
    } catch (err) {
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
        // Refresh permissions
        await fetchPermissions();
      } else {
        const errorData = await response.json();
        setError(errorData.error || 'Failed to update permission');
      }
    } catch (err) {
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
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>Manage Permissions - {otherUserName}</DialogTitle>
      <DialogContent>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
            <CircularProgress />
          </Box>
        ) : (
          <>
            {error && (
              <Alert severity="error" sx={{ mb: 2 }}>
                {error}
              </Alert>
            )}

            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              Control what services you allow this contact to access. Permissions are unidirectional - you control what they can access from you, and vice versa.
            </Typography>

            {/* Permissions I Grant to Them */}
            <Box sx={{ mb: 4 }}>
              <Typography variant="h6" gutterBottom>
                Permissions I Grant to {otherUserName}
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                What {otherUserName} can access from you:
              </Typography>
              {SERVICE_TYPES.map((service) => {
                const granted = getPermission(currentUserId, otherUserId, service.type);
                return (
                  <FormControlLabel
                    key={`grant-${service.type}`}
                    control={
                      <Switch
                        checked={granted}
                        onChange={() => handleTogglePermission(service.type, otherUserId, granted)}
                        disabled={saving}
                      />
                    }
                    label={service.label}
                  />
                );
              })}
            </Box>

            <Divider sx={{ my: 3 }} />

            {/* Permissions They Grant to Me */}
            <Box>
              <Typography variant="h6" gutterBottom>
                Permissions {otherUserName} Grants to Me
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                What you can access from {otherUserName}:
              </Typography>
              {SERVICE_TYPES.map((service) => {
                const granted = getPermission(otherUserId, currentUserId, service.type);
                return (
                  <FormControlLabel
                    key={`receive-${service.type}`}
                    control={<Switch checked={granted} disabled />}
                    label={service.label}
                  />
                );
              })}
            </Box>
          </>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Close</Button>
      </DialogActions>
    </Dialog>
  );
}
