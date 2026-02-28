import { useState, useEffect } from 'react';
import { useSession } from 'next-auth/react';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import Button from '@mui/material/Button';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import CircularProgress from '@mui/material/CircularProgress';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import Chip from '@mui/material/Chip';
import { formatISK } from '@industry-tool/utils/formatting';

type JobSlotInterestRequest = {
  id: number;
  listingId: number;
  requesterUserId: number;
  requesterName: string;
  slotsRequested: number;
  durationDays: number | null;
  message: string | null;
  status: string;
  createdAt: string;
  updatedAt: string;
  listingActivityType?: string;
  listingCharacterName?: string;
  listingOwnerName?: string;
  listingPriceAmount?: number;
  listingPricingUnit?: string;
};

const ACTIVITY_LABELS: Record<string, string> = {
  manufacturing: 'Manufacturing',
  reaction: 'Reactions',
  copying: 'Copying',
  invention: 'Invention',
  me_research: 'ME Research',
  te_research: 'TE Research',
};

const PRICING_UNIT_LABELS: Record<string, string> = {
  per_slot_day: 'Per Slot/Day',
  per_job: 'Per Job',
  flat_fee: 'Flat Fee',
};

const STATUS_COLORS: Record<string, { bg: string; border: string; text: string }> = {
  pending: { bg: 'rgba(245, 158, 11, 0.1)', border: 'rgba(245, 158, 11, 0.3)', text: '#f59e0b' },
  accepted: { bg: 'rgba(16, 185, 129, 0.1)', border: 'rgba(16, 185, 129, 0.3)', text: '#10b981' },
  declined: { bg: 'rgba(239, 68, 68, 0.1)', border: 'rgba(239, 68, 68, 0.3)', text: '#ef4444' },
  withdrawn: { bg: 'rgba(107, 114, 128, 0.1)', border: 'rgba(107, 114, 128, 0.3)', text: '#6b7280' },
};

export default function InterestRequests() {
  const { data: session } = useSession();
  const [tabIndex, setTabIndex] = useState(0);
  const [sentRequests, setSentRequests] = useState<JobSlotInterestRequest[]>([]);
  const [receivedRequests, setReceivedRequests] = useState<JobSlotInterestRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity?: 'success' | 'error' }>({
    open: false,
    message: '',
    severity: 'success',
  });

  useEffect(() => {
    if (session) {
      fetchRequests();
    }
  }, [session]);

  const fetchRequests = async () => {
    setLoading(true);
    try {
      const [sentResponse, receivedResponse] = await Promise.all([
        fetch('/api/job-slots/interest/sent'),
        fetch('/api/job-slots/interest/received'),
      ]);

      if (sentResponse.ok) {
        const sentData = await sentResponse.json();
        setSentRequests(sentData || []);
      }

      if (receivedResponse.ok) {
        const receivedData = await receivedResponse.json();
        setReceivedRequests(receivedData || []);
      }
    } catch (error) {
      console.error('Failed to fetch interest requests:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleWithdraw = async (id: number) => {
    if (!confirm('Are you sure you want to withdraw this interest request?')) return;

    try {
      const response = await fetch(`/api/job-slots/interest/${id}/status`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: 'withdrawn' }),
      });

      if (response.ok) {
        setSnackbar({ open: true, message: 'Interest withdrawn', severity: 'success' });
        fetchRequests();
      } else {
        const error = await response.json();
        setSnackbar({ open: true, message: error.error || 'Failed to withdraw', severity: 'error' });
      }
    } catch (error) {
      console.error('Withdraw failed:', error);
      setSnackbar({ open: true, message: 'Failed to withdraw', severity: 'error' });
    }
  };

  const handleAccept = async (id: number) => {
    try {
      const response = await fetch(`/api/job-slots/interest/${id}/status`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: 'accepted' }),
      });

      if (response.ok) {
        setSnackbar({ open: true, message: 'Interest accepted', severity: 'success' });
        fetchRequests();
      } else {
        const error = await response.json();
        setSnackbar({ open: true, message: error.error || 'Failed to accept', severity: 'error' });
      }
    } catch (error) {
      console.error('Accept failed:', error);
      setSnackbar({ open: true, message: 'Failed to accept', severity: 'error' });
    }
  };

  const handleDecline = async (id: number) => {
    if (!confirm('Are you sure you want to decline this interest request?')) return;

    try {
      const response = await fetch(`/api/job-slots/interest/${id}/status`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: 'declined' }),
      });

      if (response.ok) {
        setSnackbar({ open: true, message: 'Interest declined', severity: 'success' });
        fetchRequests();
      } else {
        const error = await response.json();
        setSnackbar({ open: true, message: error.error || 'Failed to decline', severity: 'error' });
      }
    } catch (error) {
      console.error('Decline failed:', error);
      setSnackbar({ open: true, message: 'Failed to decline', severity: 'error' });
    }
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      <Tabs value={tabIndex} onChange={(_, newValue) => setTabIndex(newValue)} sx={{ mb: 3 }}>
        <Tab label={`Sent (${sentRequests.length})`} />
        <Tab label={`Received (${receivedRequests.length})`} />
      </Tabs>

      {tabIndex === 0 && (
        <>
          {sentRequests.length === 0 ? (
            <Paper sx={{ p: 4, textAlign: 'center' }}>
              <Typography variant="h6" color="text.secondary">
                No sent interest requests
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                Browse listings and express interest to get started.
              </Typography>
            </Paper>
          ) : (
            <TableContainer component={Paper}>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Owner</TableCell>
                    <TableCell>Character</TableCell>
                    <TableCell>Activity</TableCell>
                    <TableCell align="right">Slots</TableCell>
                    <TableCell align="right">Duration</TableCell>
                    <TableCell align="right">Price</TableCell>
                    <TableCell>Message</TableCell>
                    <TableCell>Status</TableCell>
                    <TableCell align="center">Action</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {sentRequests.map((request) => (
                    <TableRow key={request.id} hover>
                      <TableCell>{request.listingOwnerName || '-'}</TableCell>
                      <TableCell>{request.listingCharacterName || '-'}</TableCell>
                      <TableCell>
                        {request.listingActivityType && (
                          <Chip
                            label={ACTIVITY_LABELS[request.listingActivityType] || request.listingActivityType}
                            size="small"
                            sx={{
                              background: 'rgba(59, 130, 246, 0.1)',
                              borderColor: 'rgba(59, 130, 246, 0.3)',
                              color: '#60a5fa',
                            }}
                          />
                        )}
                      </TableCell>
                      <TableCell align="right">{request.slotsRequested}</TableCell>
                      <TableCell align="right">{request.durationDays ? `${request.durationDays} days` : '-'}</TableCell>
                      <TableCell align="right">
                        {request.listingPriceAmount !== undefined && request.listingPricingUnit
                          ? `${formatISK(request.listingPriceAmount)} ${PRICING_UNIT_LABELS[request.listingPricingUnit] || request.listingPricingUnit}`
                          : '-'}
                      </TableCell>
                      <TableCell>
                        {request.message ? (
                          <Typography variant="caption" sx={{ color: '#94a3b8' }}>
                            {request.message}
                          </Typography>
                        ) : '-'}
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={request.status}
                          size="small"
                          sx={{
                            background: STATUS_COLORS[request.status]?.bg || 'rgba(107, 114, 128, 0.1)',
                            borderColor: STATUS_COLORS[request.status]?.border || 'rgba(107, 114, 128, 0.3)',
                            color: STATUS_COLORS[request.status]?.text || '#6b7280',
                            textTransform: 'capitalize',
                          }}
                        />
                      </TableCell>
                      <TableCell align="center">
                        {request.status === 'pending' && (
                          <Button
                            variant="outlined"
                            size="small"
                            color="error"
                            onClick={() => handleWithdraw(request.id)}
                          >
                            Withdraw
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </>
      )}

      {tabIndex === 1 && (
        <>
          {receivedRequests.length === 0 ? (
            <Paper sx={{ p: 4, textAlign: 'center' }}>
              <Typography variant="h6" color="text.secondary">
                No received interest requests
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                Create listings to receive interest requests.
              </Typography>
            </Paper>
          ) : (
            <TableContainer component={Paper}>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Requester</TableCell>
                    <TableCell align="right">Slots</TableCell>
                    <TableCell align="right">Duration</TableCell>
                    <TableCell>Message</TableCell>
                    <TableCell>Status</TableCell>
                    <TableCell>Requested At</TableCell>
                    <TableCell align="center">Actions</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {receivedRequests.map((request) => (
                    <TableRow key={request.id} hover>
                      <TableCell>
                        <Chip
                          label={request.requesterName}
                          size="small"
                          sx={{
                            background: 'rgba(59, 130, 246, 0.1)',
                            borderColor: 'rgba(59, 130, 246, 0.3)',
                            color: '#60a5fa',
                          }}
                        />
                      </TableCell>
                      <TableCell align="right">{request.slotsRequested}</TableCell>
                      <TableCell align="right">{request.durationDays ? `${request.durationDays} days` : '-'}</TableCell>
                      <TableCell>
                        {request.message ? (
                          <Typography variant="caption" sx={{ color: '#94a3b8' }}>
                            {request.message}
                          </Typography>
                        ) : '-'}
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={request.status}
                          size="small"
                          sx={{
                            background: STATUS_COLORS[request.status]?.bg || 'rgba(107, 114, 128, 0.1)',
                            borderColor: STATUS_COLORS[request.status]?.border || 'rgba(107, 114, 128, 0.3)',
                            color: STATUS_COLORS[request.status]?.text || '#6b7280',
                            textTransform: 'capitalize',
                          }}
                        />
                      </TableCell>
                      <TableCell>{new Date(request.createdAt).toLocaleDateString()}</TableCell>
                      <TableCell align="center">
                        {request.status === 'pending' && (
                          <Box sx={{ display: 'flex', gap: 1, justifyContent: 'center' }}>
                            <Button
                              variant="contained"
                              size="small"
                              color="success"
                              onClick={() => handleAccept(request.id)}
                            >
                              Accept
                            </Button>
                            <Button
                              variant="outlined"
                              size="small"
                              color="error"
                              onClick={() => handleDecline(request.id)}
                            >
                              Decline
                            </Button>
                          </Box>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </>
      )}

      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert
          onClose={() => setSnackbar({ ...snackbar, open: false })}
          severity={snackbar.severity || 'success'}
          sx={{ width: '100%' }}
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
}
