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
import IconButton from '@mui/material/IconButton';
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
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import { formatISK } from '@industry-tool/utils/formatting';

type JobSlotRentalListing = {
  id: number;
  userId: number;
  characterId: number;
  characterName: string;
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

type CharacterSlotInventory = {
  characterId: number;
  characterName: string;
  slotsByActivity: Record<string, { slotsAvailable: number }>;
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

export default function MyListings() {
  const { data: session } = useSession();
  const [listings, setListings] = useState<JobSlotRentalListing[]>([]);
  const [inventory, setInventory] = useState<CharacterSlotInventory[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedListing, setSelectedListing] = useState<JobSlotRentalListing | null>(null);
  const [formData, setFormData] = useState<{
    characterId: number;
    activityType: string;
    slotsListed: number;
    priceAmount: number;
    pricingUnit: string;
    locationName: string;
    notes: string;
  }>({
    characterId: 0,
    activityType: '',
    slotsListed: 1,
    priceAmount: 0,
    pricingUnit: 'per_slot_day',
    locationName: '',
    notes: '',
  });
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity?: 'success' | 'error' }>({
    open: false,
    message: '',
    severity: 'success',
  });

  useEffect(() => {
    if (session) {
      fetchListings();
      fetchInventory();
    }
  }, [session]);

  const fetchListings = async () => {
    try {
      const response = await fetch('/api/job-slots/listings');
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

  const fetchInventory = async () => {
    try {
      const response = await fetch('/api/job-slots/inventory');
      if (response.ok) {
        const data = await response.json();
        setInventory(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch inventory:', error);
    }
  };

  const handleCreate = () => {
    setSelectedListing(null);
    setFormData({
      characterId: 0,
      activityType: '',
      slotsListed: 1,
      priceAmount: 0,
      pricingUnit: 'per_slot_day',
      locationName: '',
      notes: '',
    });
    setDialogOpen(true);
  };

  const handleEdit = (listing: JobSlotRentalListing) => {
    setSelectedListing(listing);
    setFormData({
      characterId: listing.characterId,
      activityType: listing.activityType,
      slotsListed: listing.slotsListed,
      priceAmount: listing.priceAmount,
      pricingUnit: listing.pricingUnit,
      locationName: listing.locationName || '',
      notes: listing.notes || '',
    });
    setDialogOpen(true);
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this listing?')) return;

    try {
      const response = await fetch(`/api/job-slots/listings/${id}`, {
        method: 'DELETE',
      });

      if (response.ok) {
        setSnackbar({ open: true, message: 'Listing deleted successfully', severity: 'success' });
        fetchListings();
      } else {
        const error = await response.json();
        setSnackbar({ open: true, message: error.error || 'Failed to delete listing', severity: 'error' });
      }
    } catch (error) {
      console.error('Delete failed:', error);
      setSnackbar({ open: true, message: 'Failed to delete listing', severity: 'error' });
    }
  };

  const handleSave = async () => {
    if (!formData.characterId || !formData.activityType || formData.slotsListed <= 0 || formData.priceAmount < 0) {
      setSnackbar({ open: true, message: 'Please fill in all required fields', severity: 'error' });
      return;
    }

    try {
      const url = selectedListing ? `/api/job-slots/listings/${selectedListing.id}` : '/api/job-slots/listings';
      const method = selectedListing ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          characterId: formData.characterId,
          activityType: formData.activityType,
          slotsListed: formData.slotsListed,
          priceAmount: formData.priceAmount,
          pricingUnit: formData.pricingUnit,
          locationName: formData.locationName || null,
          notes: formData.notes || null,
        }),
      });

      if (response.ok) {
        setDialogOpen(false);
        setSnackbar({ open: true, message: selectedListing ? 'Listing updated' : 'Listing created', severity: 'success' });
        fetchListings();
        fetchInventory();
      } else {
        const error = await response.json();
        setSnackbar({ open: true, message: error.error || 'Failed to save listing', severity: 'error' });
      }
    } catch (error) {
      console.error('Save failed:', error);
      setSnackbar({ open: true, message: 'Failed to save listing', severity: 'error' });
    }
  };

  const getAvailableActivities = () => {
    if (!formData.characterId) return [];
    const char = inventory.find(c => c.characterId === formData.characterId);
    if (!char) return [];
    return Object.entries(char.slotsByActivity)
      .filter(([_, info]) => info.slotsAvailable > 0)
      .map(([activityType]) => activityType);
  };

  const getMaxSlots = () => {
    if (!formData.characterId || !formData.activityType) return 0;
    const char = inventory.find(c => c.characterId === formData.characterId);
    if (!char) return 0;
    return char.slotsByActivity[formData.activityType]?.slotsAvailable || 0;
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
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h5">My Slot Listings</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={handleCreate}>
          Create Listing
        </Button>
      </Box>

      {listings.length === 0 ? (
        <Paper sx={{ p: 4, textAlign: 'center' }}>
          <Typography variant="h6" color="text.secondary">
            No listings yet
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
            Create your first listing to rent out idle job slots.
          </Typography>
        </Paper>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Character</TableCell>
                <TableCell>Activity</TableCell>
                <TableCell align="right">Slots Listed</TableCell>
                <TableCell align="right">Price</TableCell>
                <TableCell>Pricing Unit</TableCell>
                <TableCell>Location</TableCell>
                <TableCell>Notes</TableCell>
                <TableCell align="center">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {listings.map((listing) => (
                <TableRow key={listing.id} hover>
                  <TableCell>{listing.characterName}</TableCell>
                  <TableCell>
                    <Chip
                      label={ACTIVITY_LABELS[listing.activityType] || listing.activityType}
                      size="small"
                      sx={{
                        background: 'rgba(59, 130, 246, 0.1)',
                        borderColor: 'rgba(59, 130, 246, 0.3)',
                        color: '#60a5fa',
                      }}
                    />
                  </TableCell>
                  <TableCell align="right">{listing.slotsListed}</TableCell>
                  <TableCell align="right">{formatISK(listing.priceAmount)}</TableCell>
                  <TableCell>{PRICING_UNIT_LABELS[listing.pricingUnit] || listing.pricingUnit}</TableCell>
                  <TableCell>{listing.locationName || '-'}</TableCell>
                  <TableCell>{listing.notes || '-'}</TableCell>
                  <TableCell align="center">
                    <IconButton size="small" color="primary" onClick={() => handleEdit(listing)} aria-label="Edit">
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" color="error" onClick={() => handleDelete(listing.id)} aria-label="Delete">
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{selectedListing ? 'Edit Listing' : 'Create Listing'}</DialogTitle>
        <DialogContent>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 2 }}>
            <FormControl fullWidth required disabled={!!selectedListing}>
              <InputLabel>Character</InputLabel>
              <Select
                value={formData.characterId}
                onChange={(e) => setFormData({ ...formData, characterId: e.target.value as number, activityType: '' })}
                label="Character"
              >
                <MenuItem value={0} disabled>Select a character</MenuItem>
                {inventory.map((char) => (
                  <MenuItem key={char.characterId} value={char.characterId}>
                    {char.characterName}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <FormControl fullWidth required disabled={!formData.characterId || !!selectedListing}>
              <InputLabel>Activity Type</InputLabel>
              <Select
                value={formData.activityType}
                onChange={(e) => setFormData({ ...formData, activityType: e.target.value })}
                label="Activity Type"
              >
                <MenuItem value="" disabled>Select an activity</MenuItem>
                {getAvailableActivities().map((activity) => (
                  <MenuItem key={activity} value={activity}>
                    {ACTIVITY_LABELS[activity] || activity}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <TextField
              label="Slots to List"
              type="number"
              fullWidth
              required
              value={formData.slotsListed}
              onChange={(e) => setFormData({ ...formData, slotsListed: parseInt(e.target.value) || 0 })}
              InputProps={{ inputProps: { min: 1, max: getMaxSlots() } }}
              helperText={`Max available: ${getMaxSlots()}`}
              disabled={!formData.activityType}
            />

            <TextField
              label="Price Amount (ISK)"
              type="number"
              fullWidth
              required
              value={formData.priceAmount}
              onChange={(e) => setFormData({ ...formData, priceAmount: parseFloat(e.target.value) || 0 })}
              InputProps={{ inputProps: { min: 0 } }}
            />

            <FormControl fullWidth required>
              <InputLabel>Pricing Unit</InputLabel>
              <Select
                value={formData.pricingUnit}
                onChange={(e) => setFormData({ ...formData, pricingUnit: e.target.value })}
                label="Pricing Unit"
              >
                {Object.entries(PRICING_UNIT_LABELS).map(([value, label]) => (
                  <MenuItem key={value} value={value}>
                    {label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <TextField
              label="Location (optional)"
              fullWidth
              value={formData.locationName}
              onChange={(e) => setFormData({ ...formData, locationName: e.target.value })}
              placeholder="e.g., Jita 4-4"
            />

            <TextField
              label="Notes (optional)"
              multiline
              rows={3}
              fullWidth
              value={formData.notes}
              onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
              placeholder="Additional information about this listing..."
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>Cancel</Button>
          <Button onClick={handleSave} variant="contained">
            {selectedListing ? 'Update' : 'Create'}
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
