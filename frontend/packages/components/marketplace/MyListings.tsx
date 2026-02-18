import { useState, useEffect, useRef, useMemo } from 'react';
import { useSession } from "next-auth/react";
import Container from '@mui/material/Container';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Button from '@mui/material/Button';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import IconButton from '@mui/material/IconButton';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import AddIcon from '@mui/icons-material/Add';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import SearchIcon from '@mui/icons-material/Search';
import SellIcon from '@mui/icons-material/Sell';
import AutorenewIcon from '@mui/icons-material/Autorenew';
import Chip from '@mui/material/Chip';
import Badge from '@mui/material/Badge';
import Tooltip from '@mui/material/Tooltip';

const AutoSellIcon = () => (
  <Badge
    badgeContent={<AutorenewIcon sx={{ fontSize: '0.6rem' }} />}
    sx={{
      '& .MuiBadge-badge': {
        minWidth: 'unset',
        height: 'unset',
        padding: 0,
        backgroundColor: 'transparent',
        color: 'inherit',
        top: 2,
        right: 2,
      },
    }}
  >
    <SellIcon />
  </Badge>
);
import Loading from "@industry-tool/components/loading";

export type ForSaleItem = {
  id: number;
  userId: number;
  typeId: number;
  typeName: string;
  ownerType: string;
  ownerId: number;
  ownerName: string;
  locationId: number;
  locationName: string;
  containerId?: number;
  divisionNumber?: number;
  quantityAvailable: number;
  pricePerUnit: number;
  notes?: string;
  autoSellContainerId?: number;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

type ListingFormData = {
  typeId: number;
  ownerType: string;
  ownerId: number;
  locationId: number;
  containerId?: number;
  divisionNumber?: number;
  quantityAvailable: number;
  pricePerUnit: number;
  notes?: string;
};

export default function MyListings() {
  const { data: session } = useSession();
  const [listings, setListings] = useState<ForSaleItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [snackbarOpen, setSnackbarOpen] = useState(false);
  const [snackbarMessage, setSnackbarMessage] = useState('');
  const [snackbarSeverity, setSnackbarSeverity] = useState<'success' | 'error'>('success');
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [selectedListing, setSelectedListing] = useState<ForSaleItem | null>(null);
  const [formData, setFormData] = useState<Partial<ListingFormData>>({});
  const hasFetchedRef = useRef(false);

  useEffect(() => {
    if (session && !hasFetchedRef.current) {
      hasFetchedRef.current = true;
      fetchListings();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [session]);

  const fetchListings = async () => {
    if (!session) return;

    setLoading(true);
    try {
      const response = await fetch('/api/for-sale');
      if (response.ok) {
        const data: ForSaleItem[] = await response.json();
        setListings(data || []);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleEditClick = (listing: ForSaleItem) => {
    setSelectedListing(listing);
    setFormData({
      quantityAvailable: listing.quantityAvailable,
      pricePerUnit: listing.pricePerUnit,
      notes: listing.notes,
    });
    setEditDialogOpen(true);
  };

  const handleEditSave = async () => {
    if (!selectedListing || !session) return;

    try {
      const response = await fetch(`/api/for-sale/${selectedListing.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      });

      if (response.ok) {
        showSnackbar('Listing updated successfully', 'success');
        setEditDialogOpen(false);
        setSelectedListing(null);
        setFormData({});
        await fetchListings();
      } else {
        const error = await response.json();
        showSnackbar(error.error || 'Failed to update listing', 'error');
      }
    } catch (err) {
      showSnackbar('Failed to update listing', 'error');
    }
  };

  const handleDelete = async (listingId: number) => {
    if (!confirm('Are you sure you want to delete this listing?')) return;

    try {
      const response = await fetch(`/api/for-sale/${listingId}`, {
        method: 'DELETE',
      });

      if (response.ok) {
        showSnackbar('Listing deleted successfully', 'success');
        await fetchListings();
      } else {
        const error = await response.json();
        showSnackbar(error.error || 'Failed to delete listing', 'error');
      }
    } catch (err) {
      showSnackbar('Failed to delete listing', 'error');
    }
  };

  const showSnackbar = (message: string, severity: 'success' | 'error') => {
    setSnackbarMessage(message);
    setSnackbarSeverity(severity);
    setSnackbarOpen(true);
  };

  // Filter listings based on search
  const filteredListings = useMemo(() => {
    if (!searchQuery) return listings;

    const query = searchQuery.toLowerCase();
    return listings.filter(
      (item) =>
        item.typeName.toLowerCase().includes(query) ||
        item.ownerName.toLowerCase().includes(query) ||
        item.locationName.toLowerCase().includes(query)
    );
  }, [listings, searchQuery]);

  // Calculate totals
  const totalValue = useMemo(() => {
    return filteredListings.reduce((sum, item) => {
      return sum + (item.quantityAvailable * item.pricePerUnit);
    }, 0);
  }, [filteredListings]);

  if (!session) {
    return null;
  }

  if (loading) {
    return <Loading />;
  }

  return (
    <Container maxWidth={false} sx={{ mt: 4, mb: 4 }}>
      <Box sx={{ mb: 3 }}>
        <Typography variant="h4" gutterBottom>
          My Listings
        </Typography>

        {/* Summary Stats */}
        <Box sx={{ display: 'flex', gap: 2, mb: 3 }}>
          <Card sx={{ flex: 1 }}>
            <CardContent>
              <Typography variant="h6" color="text.secondary" gutterBottom>
                Active Listings
              </Typography>
              <Typography variant="h3">{filteredListings.length}</Typography>
            </CardContent>
          </Card>
          <Card sx={{ flex: 1 }}>
            <CardContent>
              <Typography variant="h6" color="text.secondary" gutterBottom>
                Total Value
              </Typography>
              <Typography variant="h3">
                {totalValue.toLocaleString(undefined, { maximumFractionDigits: 0 })} ISK
              </Typography>
            </CardContent>
          </Card>
        </Box>

        {/* Search */}
        <Box sx={{ mb: 2 }}>
          <TextField
            fullWidth
            size="small"
            placeholder="Search items, owners, or locations..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon fontSize="small" />
                </InputAdornment>
              ),
            }}
          />
        </Box>
      </Box>

      {/* Listings Table */}
      {filteredListings.length === 0 ? (
        <Card>
          <CardContent>
            <Typography variant="h6" align="center" color="text.secondary">
              {listings.length === 0
                ? 'No active listings. Create your first listing to get started!'
                : 'No items match your search.'}
            </Typography>
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardContent sx={{ p: 0 }}>
            <TableContainer component={Paper} variant="outlined">
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Item</TableCell>
                    <TableCell>Owner</TableCell>
                    <TableCell>Location</TableCell>
                    <TableCell align="right">Quantity</TableCell>
                    <TableCell align="right">Price/Unit</TableCell>
                    <TableCell align="right">Total Value</TableCell>
                    <TableCell>Notes</TableCell>
                    <TableCell align="center">Actions</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {filteredListings.map((item) => (
                    <TableRow
                      key={item.id}
                      hover
                      sx={{
                        '&:nth-of-type(odd)': {
                          backgroundColor: 'action.hover',
                        },
                      }}
                    >
                      <TableCell sx={{ fontWeight: 600 }}>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          {item.typeName}
                          {item.autoSellContainerId && (
                            <Tooltip title="Auto-managed listing — changes will be overwritten on next sync">
                              <Chip
                                icon={<AutoSellIcon />}
                                label="Auto"
                                size="small"
                                sx={{
                                  fontSize: '0.65rem',
                                  fontWeight: 600,
                                  height: 22,
                                  background: 'rgba(59, 130, 246, 0.15)',
                                  color: '#3b82f6',
                                  border: '1px solid rgba(59, 130, 246, 0.3)',
                                  '& .MuiChip-icon': { color: '#3b82f6', fontSize: '0.8rem' },
                                }}
                              />
                            </Tooltip>
                          )}
                        </Box>
                      </TableCell>
                      <TableCell>{item.ownerName}</TableCell>
                      <TableCell>{item.locationName}</TableCell>
                      <TableCell align="right">{item.quantityAvailable.toLocaleString()}</TableCell>
                      <TableCell align="right">
                        {item.pricePerUnit.toLocaleString(undefined, { maximumFractionDigits: 2 })}
                      </TableCell>
                      <TableCell align="right">
                        {(item.quantityAvailable * item.pricePerUnit).toLocaleString(undefined, { maximumFractionDigits: 0 })}
                      </TableCell>
                      <TableCell>{item.notes || '-'}</TableCell>
                      <TableCell align="center">
                        <IconButton
                          size="small"
                          color="primary"
                          onClick={() => handleEditClick(item)}
                          aria-label="edit"
                        >
                          <EditIcon fontSize="small" />
                        </IconButton>
                        <IconButton
                          size="small"
                          color="error"
                          onClick={() => handleDelete(item.id)}
                          aria-label="delete"
                        >
                          <DeleteIcon fontSize="small" />
                        </IconButton>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </CardContent>
        </Card>
      )}

      {/* Edit Dialog */}
      <Dialog open={editDialogOpen} onClose={() => setEditDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Edit Listing</DialogTitle>
        <DialogContent>
          {selectedListing && (
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 1 }}>
              <Typography variant="body2" color="text.secondary">
                Item: <strong>{selectedListing.typeName}</strong>
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Location: <strong>{selectedListing.locationName}</strong>
              </Typography>

              <TextField
                label="Quantity Available"
                type="number"
                fullWidth
                value={formData.quantityAvailable || ''}
                onChange={(e) => setFormData({ ...formData, quantityAvailable: parseInt(e.target.value) })}
                InputProps={{ inputProps: { min: 1 } }}
              />

              <TextField
                label="Price Per Unit (ISK)"
                type="number"
                fullWidth
                value={formData.pricePerUnit || ''}
                onChange={(e) => setFormData({ ...formData, pricePerUnit: parseInt(e.target.value) })}
                InputProps={{ inputProps: { min: 0 } }}
              />

              <TextField
                label="Notes (optional)"
                multiline
                rows={3}
                fullWidth
                value={formData.notes || ''}
                onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
              />
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setEditDialogOpen(false)}>Cancel</Button>
          <Button
            onClick={handleEditSave}
            variant="contained"
            disabled={!formData.quantityAvailable || formData.quantityAvailable <= 0 || formData.pricePerUnit === undefined || formData.pricePerUnit < 0}
          >
            Save Changes
          </Button>
        </DialogActions>
      </Dialog>

      {/* Snackbar */}
      <Snackbar
        open={snackbarOpen}
        autoHideDuration={3000}
        onClose={() => setSnackbarOpen(false)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert onClose={() => setSnackbarOpen(false)} severity={snackbarSeverity} sx={{ width: '100%' }}>
          {snackbarMessage}
        </Alert>
      </Snackbar>
    </Container>
  );
}
