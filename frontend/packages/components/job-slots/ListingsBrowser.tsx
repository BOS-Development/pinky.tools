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
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import CircularProgress from '@mui/material/CircularProgress';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import Chip from '@mui/material/Chip';
import { formatISK } from '@industry-tool/utils/formatting';

type JobSlotRentalListing = {
  id: number;
  userId: number;
  characterId: number;
  characterName: string;
  ownerName: string;
  activityType: string;
  slotsListed: number;
  priceAmount: number;
  pricingUnit: string;
  locationId: number | null;
  locationName: string;
  notes: string | null;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
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

export default function ListingsBrowser() {
  const { data: session } = useSession();
  const [listings, setListings] = useState<JobSlotRentalListing[]>([]);
  const [loading, setLoading] = useState(true);
  const [activityFilter, setActivityFilter] = useState<string>('all');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedListing, setSelectedListing] = useState<JobSlotRentalListing | null>(null);
  const [interestData, setInterestData] = useState<{
    slotsRequested: number;
    durationDays: number | null;
    message: string;
  }>({
    slotsRequested: 1,
    durationDays: null,
    message: '',
  });
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity?: 'success' | 'error' }>({
    open: false,
    message: '',
    severity: 'success',
  });

  useEffect(() => {
    if (session) {
      fetchListings();
    }
  }, [session]);

  const fetchListings = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/job-slots/listings/browse');
      if (response.ok) {
        const data = await response.json();
        setListings(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch listings:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleExpressInterest = (listing: JobSlotRentalListing) => {
    setSelectedListing(listing);
    setInterestData({
      slotsRequested: 1,
      durationDays: null,
      message: '',
    });
    setDialogOpen(true);
  };

  const handleSubmitInterest = async () => {
    if (!selectedListing || interestData.slotsRequested <= 0 || interestData.slotsRequested > selectedListing.slotsListed) {
      setSnackbar({ open: true, message: 'Invalid slots requested', severity: 'error' });
      return;
    }

    try {
      const response = await fetch('/api/job-slots/interest', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          listingId: selectedListing.id,
          slotsRequested: interestData.slotsRequested,
          durationDays: interestData.durationDays,
          message: interestData.message || null,
        }),
      });

      if (response.ok) {
        setDialogOpen(false);
        setSnackbar({ open: true, message: 'Interest submitted successfully', severity: 'success' });
      } else {
        const error = await response.json();
        setSnackbar({ open: true, message: error.error || 'Failed to submit interest', severity: 'error' });
      }
    } catch (error) {
      console.error('Submit interest failed:', error);
      setSnackbar({ open: true, message: 'Failed to submit interest', severity: 'error' });
    }
  };

  const filteredListings = listings.filter(
    listing => activityFilter === 'all' || listing.activityType === activityFilter
  );

  const uniqueActivities = Array.from(new Set(listings.map(l => l.activityType)));

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h5">Browse Listings</Typography>
        <FormControl sx={{ minWidth: 200 }}>
          <InputLabel>Filter by Activity</InputLabel>
          <Select
            value={activityFilter}
            onChange={(e) => setActivityFilter(e.target.value)}
            label="Filter by Activity"
          >
            <MenuItem value="all">All Activities</MenuItem>
            {uniqueActivities.map((activity) => (
              <MenuItem key={activity} value={activity}>
                {ACTIVITY_LABELS[activity] || activity}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      </Box>

      {filteredListings.length === 0 ? (
        <Paper sx={{ p: 4, textAlign: 'center' }}>
          <Typography variant="h6" color="text.secondary">
            No listings available
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
            {listings.length === 0
              ? "Your contacts haven't listed any slots for rent, or they haven't granted you browse permission."
              : "No listings match your selected filter."}
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
                <TableCell align="right">Slots Available</TableCell>
                <TableCell align="right">Price</TableCell>
                <TableCell>Pricing Unit</TableCell>
                <TableCell>Location</TableCell>
                <TableCell>Notes</TableCell>
                <TableCell align="center">Action</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {filteredListings.map((listing) => (
                <TableRow key={listing.id} hover>
                  <TableCell>
                    <Chip
                      label={listing.ownerName}
                      size="small"
                      sx={{
                        background: 'rgba(59, 130, 246, 0.1)',
                        borderColor: 'rgba(59, 130, 246, 0.3)',
                        color: '#60a5fa',
                      }}
                    />
                  </TableCell>
                  <TableCell>{listing.characterName}</TableCell>
                  <TableCell>
                    <Chip
                      label={ACTIVITY_LABELS[listing.activityType] || listing.activityType}
                      size="small"
                      sx={{
                        background: 'rgba(16, 185, 129, 0.1)',
                        borderColor: 'rgba(16, 185, 129, 0.3)',
                        color: '#10b981',
                      }}
                    />
                  </TableCell>
                  <TableCell align="right">{listing.slotsListed}</TableCell>
                  <TableCell align="right">{formatISK(listing.priceAmount)}</TableCell>
                  <TableCell>{PRICING_UNIT_LABELS[listing.pricingUnit] || listing.pricingUnit}</TableCell>
                  <TableCell>{listing.locationName || '-'}</TableCell>
                  <TableCell>
                    {listing.notes ? (
                      <Typography variant="caption" sx={{ color: '#94a3b8' }}>
                        {listing.notes}
                      </Typography>
                    ) : '-'}
                  </TableCell>
                  <TableCell align="center">
                    <Button
                      variant="contained"
                      size="small"
                      onClick={() => handleExpressInterest(listing)}
                    >
                      Express Interest
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Express Interest</DialogTitle>
        <DialogContent>
          {selectedListing && (
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 2 }}>
              <Typography variant="body2" gutterBottom>
                <strong>Owner:</strong> {selectedListing.ownerName}
              </Typography>
              <Typography variant="body2" gutterBottom>
                <strong>Character:</strong> {selectedListing.characterName}
              </Typography>
              <Typography variant="body2" gutterBottom>
                <strong>Activity:</strong> {ACTIVITY_LABELS[selectedListing.activityType] || selectedListing.activityType}
              </Typography>
              <Typography variant="body2" gutterBottom>
                <strong>Price:</strong> {formatISK(selectedListing.priceAmount)} {PRICING_UNIT_LABELS[selectedListing.pricingUnit]}
              </Typography>
              <Typography variant="body2" gutterBottom sx={{ mb: 2 }}>
                <strong>Slots Available:</strong> {selectedListing.slotsListed}
              </Typography>

              <TextField
                label="Slots Requested"
                type="number"
                fullWidth
                required
                value={interestData.slotsRequested}
                onChange={(e) => setInterestData({ ...interestData, slotsRequested: parseInt(e.target.value) || 0 })}
                InputProps={{ inputProps: { min: 1, max: selectedListing.slotsListed } }}
                helperText={`Max: ${selectedListing.slotsListed}`}
              />

              <TextField
                label="Duration (days) - optional"
                type="number"
                fullWidth
                value={interestData.durationDays || ''}
                onChange={(e) => setInterestData({ ...interestData, durationDays: e.target.value ? parseInt(e.target.value) : null })}
                InputProps={{ inputProps: { min: 1 } }}
                helperText="Leave empty if flexible"
              />

              <TextField
                label="Message (optional)"
                multiline
                rows={3}
                fullWidth
                value={interestData.message}
                onChange={(e) => setInterestData({ ...interestData, message: e.target.value })}
                placeholder="Additional information for the owner..."
              />
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>Cancel</Button>
          <Button onClick={handleSubmitInterest} variant="contained">
            Submit Interest
          </Button>
        </DialogActions>
      </Dialog>

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
