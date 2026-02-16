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
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import ShoppingCartIcon from '@mui/icons-material/ShoppingCart';
import { formatISK, formatNumber } from '@industry-tool/utils/formatting';

type ForSaleListing = {
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
};

export default function MarketplaceBrowser() {
  const { data: session } = useSession();
  const [listings, setListings] = useState<ForSaleListing[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [purchaseDialogOpen, setPurchaseDialogOpen] = useState(false);
  const [selectedListing, setSelectedListing] = useState<ForSaleListing | null>(null);
  const [purchaseQuantity, setPurchaseQuantity] = useState('');
  const [submittingPurchase, setSubmittingPurchase] = useState(false);
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
      const response = await fetch('/api/for-sale/browse');
      if (response.ok) {
        const data = await response.json();
        setListings(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch marketplace listings:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleOpenPurchaseDialog = (listing: ForSaleListing) => {
    setSelectedListing(listing);
    setPurchaseQuantity(listing.quantityAvailable.toLocaleString());
    setPurchaseDialogOpen(true);
  };

  const handlePurchaseQuantityChange = (value: string) => {
    const numericValue = value.replace(/\D/g, '');
    const formatted = numericValue ? parseInt(numericValue).toLocaleString() : '';
    setPurchaseQuantity(formatted);
  };

  const handlePurchase = async () => {
    if (!selectedListing) return;

    const quantity = parseInt(purchaseQuantity.replace(/,/g, ''));
    if (quantity <= 0 || quantity > selectedListing.quantityAvailable) {
      setSnackbar({ open: true, message: 'Invalid quantity', severity: 'error' });
      return;
    }

    setSubmittingPurchase(true);
    try {
      const response = await fetch('/api/purchases', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          forSaleItemId: selectedListing.id,
          quantityPurchased: quantity,
        }),
      });

      if (response.ok) {
        setPurchaseDialogOpen(false);
        await fetchListings(); // Refresh listings
        setSnackbar({ open: true, message: 'Purchase successful', severity: 'success' });
      } else {
        const error = await response.json();
        setSnackbar({ open: true, message: error.error || 'Purchase failed', severity: 'error' });
      }
    } catch (error) {
      console.error('Purchase failed:', error);
      setSnackbar({ open: true, message: 'Purchase failed', severity: 'error' });
    } finally {
      setSubmittingPurchase(false);
    }
  };

  const filteredListings = listings.filter(listing =>
    listing.typeName.toLowerCase().includes(searchQuery.toLowerCase()) ||
    listing.ownerName.toLowerCase().includes(searchQuery.toLowerCase()) ||
    listing.locationName.toLowerCase().includes(searchQuery.toLowerCase())
  );

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      <Box sx={{ mb: 3 }}>
        <TextField
          fullWidth
          label="Search listings..."
          variant="outlined"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder="Search by item name, seller, or location"
        />
      </Box>

      {filteredListings.length === 0 ? (
        <Paper sx={{ p: 4, textAlign: 'center' }}>
          <Typography variant="h6" color="text.secondary">
            No listings available
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
            {listings.length === 0
              ? "Your contacts haven't listed any items for sale, or they haven't granted you browse permission."
              : "No listings match your search."}
          </Typography>
        </Paper>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Item</TableCell>
                <TableCell>Seller</TableCell>
                <TableCell>Location</TableCell>
                <TableCell align="right">Quantity</TableCell>
                <TableCell align="right">Price per Unit</TableCell>
                <TableCell align="right">Total Value</TableCell>
                <TableCell>Notes</TableCell>
                <TableCell align="center">Action</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {filteredListings.map((listing) => (
                <TableRow key={listing.id} hover>
                  <TableCell>
                    <Typography variant="body2" sx={{ fontWeight: 500 }}>
                      {listing.typeName}
                    </Typography>
                  </TableCell>
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
                  <TableCell>
                    <Typography variant="body2" sx={{ color: '#94a3b8' }}>
                      {listing.locationName}
                    </Typography>
                  </TableCell>
                  <TableCell align="right">
                    <Typography variant="body2">{formatNumber(listing.quantityAvailable)}</Typography>
                  </TableCell>
                  <TableCell align="right">
                    <Typography variant="body2" sx={{ fontWeight: 500 }}>
                      {formatISK(listing.pricePerUnit)}
                    </Typography>
                  </TableCell>
                  <TableCell align="right">
                    <Typography variant="body2" sx={{ color: '#10b981', fontWeight: 600 }}>
                      {formatISK(listing.quantityAvailable * listing.pricePerUnit)}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    {listing.notes && (
                      <Typography variant="caption" sx={{ color: '#94a3b8' }}>
                        {listing.notes}
                      </Typography>
                    )}
                  </TableCell>
                  <TableCell align="center">
                    <Button
                      variant="contained"
                      size="small"
                      startIcon={<ShoppingCartIcon />}
                      onClick={() => handleOpenPurchaseDialog(listing)}
                    >
                      Buy
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {/* Purchase Dialog */}
      <Dialog
        open={purchaseDialogOpen}
        onClose={() => setPurchaseDialogOpen(false)}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>Purchase Item</DialogTitle>
        <DialogContent>
          {selectedListing && (
            <Box sx={{ pt: 1 }}>
              <Typography variant="body2" gutterBottom>
                <strong>Item:</strong> {selectedListing.typeName}
              </Typography>
              <Typography variant="body2" gutterBottom>
                <strong>Seller:</strong> {selectedListing.ownerName}
              </Typography>
              <Typography variant="body2" gutterBottom>
                <strong>Location:</strong> {selectedListing.locationName}
              </Typography>
              <Typography variant="body2" gutterBottom>
                <strong>Price per Unit:</strong> {selectedListing.pricePerUnit.toLocaleString()} ISK
              </Typography>
              <Typography variant="body2" gutterBottom sx={{ mb: 2 }}>
                <strong>Available:</strong> {selectedListing.quantityAvailable.toLocaleString()}
              </Typography>

              <TextField
                fullWidth
                label="Quantity to Purchase"
                type="text"
                value={purchaseQuantity}
                onChange={(e) => handlePurchaseQuantityChange(e.target.value)}
                required
                placeholder="0"
                helperText={
                  purchaseQuantity
                    ? `Total Cost: ${(
                        parseInt(purchaseQuantity.replace(/,/g, '')) * selectedListing.pricePerUnit
                      ).toLocaleString()} ISK`
                    : `Max: ${selectedListing.quantityAvailable.toLocaleString()}`
                }
              />
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setPurchaseDialogOpen(false)}>Cancel</Button>
          <Button
            onClick={handlePurchase}
            variant="contained"
            disabled={!purchaseQuantity || submittingPurchase}
          >
            {submittingPurchase ? 'Purchasing...' : 'Confirm Purchase'}
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
