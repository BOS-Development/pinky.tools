import { useState, useEffect, useRef } from 'react';
import { useSession } from "next-auth/react";
import { getItemIconUrl } from "@industry-tool/utils/eveImages";
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
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import Autocomplete from '@mui/material/Autocomplete';
import CircularProgress from '@mui/material/CircularProgress';
import Avatar from '@mui/material/Avatar';
import Loading from "@industry-tool/components/loading";

export type BuyOrder = {
  id: number;
  buyerUserId: number;
  typeId: number;
  typeName: string;
  locationId: number;
  locationName: string;
  quantityDesired: number;
  minPricePerUnit: number;
  maxPricePerUnit: number;
  notes?: string;
  autoBuyConfigId?: number;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

type BuyOrderFormData = {
  typeId: number;
  typeName?: string;
  locationId: number;
  quantityDesired: number;
  minPricePerUnit: number;
  maxPricePerUnit: number;
  notes?: string;
};

type ItemType = {
  TypeID: number;
  TypeName: string;
  Volume: number;
  IconID?: number;
};

type StationOption = {
  stationId: number;
  name: string;
  solarSystemName: string;
};

export default function BuyOrders() {
  const { data: session } = useSession();
  const [orders, setOrders] = useState<BuyOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [snackbarOpen, setSnackbarOpen] = useState(false);
  const [snackbarMessage, setSnackbarMessage] = useState('');
  const [snackbarSeverity, setSnackbarSeverity] = useState<'success' | 'error'>('success');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedOrder, setSelectedOrder] = useState<BuyOrder | null>(null);
  const [formData, setFormData] = useState<Partial<BuyOrderFormData>>({});
  const [itemOptions, setItemOptions] = useState<ItemType[]>([]);
  const [itemSearchLoading, setItemSearchLoading] = useState(false);
  const [selectedItem, setSelectedItem] = useState<ItemType | null>(null);
  const [stationOptions, setStationOptions] = useState<StationOption[]>([]);
  const [stationSearchLoading, setStationSearchLoading] = useState(false);
  const [selectedStation, setSelectedStation] = useState<StationOption | null>(null);
  const hasFetchedRef = useRef(false);
  const searchTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const stationSearchTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    if (session && !hasFetchedRef.current) {
      hasFetchedRef.current = true;
      fetchOrders();
    }
  }, [session]);

  const fetchOrders = async () => {
    try {
      const response = await fetch('/api/buy-orders');
      if (!response.ok) throw new Error('Failed to fetch buy orders');
      const data = await response.json();
      setOrders(data);
    } catch (error) {
      console.error('Error fetching buy orders:', error);
      showSnackbar('Failed to load buy orders', 'error');
    } finally {
      setLoading(false);
    }
  };

  const searchItems = async (query: string) => {
    if (!query || query.length < 2) {
      setItemOptions([]);
      return;
    }

    setItemSearchLoading(true);
    try {
      const response = await fetch(`/api/item-types/search?q=${encodeURIComponent(query)}`);
      if (!response.ok) throw new Error('Failed to search items');
      const data = await response.json();
      setItemOptions(data || []);
    } catch (error) {
      console.error('Error searching items:', error);
      setItemOptions([]);
    } finally {
      setItemSearchLoading(false);
    }
  };

  const handleItemSearch = (value: string) => {
    if (searchTimeoutRef.current) {
      clearTimeout(searchTimeoutRef.current);
    }

    searchTimeoutRef.current = setTimeout(() => {
      searchItems(value);
    }, 300);
  };

  const searchStations = async (query: string) => {
    if (!query || query.length < 2) {
      setStationOptions([]);
      return;
    }

    setStationSearchLoading(true);
    try {
      const response = await fetch(`/api/stations/search?q=${encodeURIComponent(query)}`);
      if (!response.ok) throw new Error('Failed to search stations');
      const data = await response.json();
      setStationOptions(data || []);
    } catch (error) {
      console.error('Error searching stations:', error);
      setStationOptions([]);
    } finally {
      setStationSearchLoading(false);
    }
  };

  const handleStationSearch = (value: string) => {
    if (stationSearchTimeoutRef.current) {
      clearTimeout(stationSearchTimeoutRef.current);
    }

    stationSearchTimeoutRef.current = setTimeout(() => {
      searchStations(value);
    }, 300);
  };

  const handleCreate = () => {
    setSelectedOrder(null);
    setSelectedItem(null);
    setSelectedStation(null);
    setFormData({ quantityDesired: 0, minPricePerUnit: 0, maxPricePerUnit: 0, locationId: 0 });
    setDialogOpen(true);
  };

  const handleEdit = (order: BuyOrder) => {
    setSelectedOrder(order);
    setSelectedItem({
      TypeID: order.typeId,
      TypeName: order.typeName,
      Volume: 0,
    });
    setSelectedStation({
      stationId: order.locationId,
      name: order.locationName,
      solarSystemName: '',
    });
    setFormData({
      typeId: order.typeId,
      typeName: order.typeName,
      locationId: order.locationId,
      quantityDesired: order.quantityDesired,
      minPricePerUnit: order.minPricePerUnit,
      maxPricePerUnit: order.maxPricePerUnit,
      notes: order.notes,
    });
    setDialogOpen(true);
  };

  const handleDelete = async (orderId: number) => {
    if (!confirm('Are you sure you want to cancel this buy order?')) return;

    try {
      const response = await fetch(`/api/buy-orders/${orderId}`, {
        method: 'DELETE',
      });

      if (!response.ok) throw new Error('Failed to delete buy order');

      showSnackbar('Buy order cancelled successfully', 'success');
      fetchOrders();
    } catch (error) {
      console.error('Error deleting buy order:', error);
      showSnackbar('Failed to cancel buy order', 'error');
    }
  };

  const handleSave = async () => {
    if (!formData.typeId || !formData.locationId || !formData.quantityDesired || formData.maxPricePerUnit === undefined) {
      showSnackbar('Please fill in all required fields', 'error');
      return;
    }

    try {
      const url = selectedOrder ? `/api/buy-orders/${selectedOrder.id}` : '/api/buy-orders';
      const method = selectedOrder ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          typeId: formData.typeId,
          locationId: formData.locationId,
          quantityDesired: formData.quantityDesired,
          minPricePerUnit: formData.minPricePerUnit || 0,
          maxPricePerUnit: formData.maxPricePerUnit,
          notes: formData.notes || null,
          ...(selectedOrder ? { isActive: true } : {}),
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to save buy order');
      }

      showSnackbar(
        selectedOrder ? 'Buy order updated successfully' : 'Buy order created successfully',
        'success'
      );
      setDialogOpen(false);
      fetchOrders();
    } catch (error: any) {
      console.error('Error saving buy order:', error);
      showSnackbar(error.message || 'Failed to save buy order', 'error');
    }
  };

  const showSnackbar = (message: string, severity: 'success' | 'error') => {
    setSnackbarMessage(message);
    setSnackbarSeverity(severity);
    setSnackbarOpen(true);
  };

  const formatNumber = (num: number) => num.toLocaleString();
  const formatISK = (isk: number) => `${isk.toLocaleString()} ISK`;
  const formatDate = (dateString: string) => new Date(dateString).toLocaleDateString();

  if (loading) {
    return <Loading />;
  }

  return (
    <Container maxWidth="xl">
      <Box sx={{ my: 4 }}>
        <Card>
          <CardContent>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
              <Typography variant="h5" component="h2">
                My Buy Orders
              </Typography>
              <Button
                variant="contained"
                startIcon={<AddIcon />}
                onClick={handleCreate}
              >
                Create Buy Order
              </Button>
            </Box>

            {orders.length === 0 ? (
              <Typography variant="body1" color="text.secondary" sx={{ textAlign: 'center', py: 4 }}>
                No buy orders yet. Create one to let sellers know what you're looking for!
              </Typography>
            ) : (
              <TableContainer component={Paper}>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>Item</TableCell>
                      <TableCell>Location</TableCell>
                      <TableCell align="right">Quantity Desired</TableCell>
                      <TableCell align="right">Price Range/Unit</TableCell>
                      <TableCell align="right">Total Budget</TableCell>
                      <TableCell>Status</TableCell>
                      <TableCell>Notes</TableCell>
                      <TableCell>Created</TableCell>
                      <TableCell align="right">Actions</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {orders.map((order) => (
                      <TableRow key={order.id}>
                        <TableCell>{order.typeName}</TableCell>
                        <TableCell>{order.locationName || '-'}</TableCell>
                        <TableCell align="right">{formatNumber(order.quantityDesired)}</TableCell>
                        <TableCell align="right">
                          {order.minPricePerUnit > 0
                            ? `${formatISK(order.minPricePerUnit)} - ${formatISK(order.maxPricePerUnit)}`
                            : formatISK(order.maxPricePerUnit)}
                        </TableCell>
                        <TableCell align="right">
                          {formatISK(order.quantityDesired * order.maxPricePerUnit)}
                        </TableCell>
                        <TableCell>
                          <Box sx={{ display: 'flex', gap: 0.5, alignItems: 'center' }}>
                            <Chip
                              label={order.isActive ? 'Active' : 'Inactive'}
                              color={order.isActive ? 'success' : 'default'}
                              size="small"
                            />
                            {order.autoBuyConfigId && (
                              <Chip
                                label="Auto"
                                size="small"
                                sx={{
                                  fontSize: '0.7rem',
                                  fontWeight: 600,
                                  background: 'rgba(245, 158, 11, 0.15)',
                                  color: '#f59e0b',
                                  border: '1px solid rgba(245, 158, 11, 0.3)',
                                }}
                              />
                            )}
                          </Box>
                        </TableCell>
                        <TableCell>{order.notes || '-'}</TableCell>
                        <TableCell>{formatDate(order.createdAt)}</TableCell>
                        <TableCell align="right">
                          {!order.autoBuyConfigId && (
                            <>
                              <IconButton
                                size="small"
                                onClick={() => handleEdit(order)}
                                title="Edit"
                              >
                                <EditIcon />
                              </IconButton>
                              <IconButton
                                size="small"
                                onClick={() => handleDelete(order.id)}
                                title="Cancel"
                              >
                                <DeleteIcon />
                              </IconButton>
                            </>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            )}
          </CardContent>
        </Card>
      </Box>

      {/* Create/Edit Dialog */}
      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>
          {selectedOrder ? 'Edit Buy Order' : 'Create Buy Order'}
        </DialogTitle>
        <DialogContent>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 2 }}>
            <Autocomplete
              value={selectedItem}
              onChange={(_, newValue) => {
                setSelectedItem(newValue);
                if (newValue) {
                  setFormData({
                    ...formData,
                    typeId: newValue.TypeID,
                    typeName: newValue.TypeName
                  });
                } else {
                  setFormData({
                    ...formData,
                    typeId: undefined,
                    typeName: undefined
                  });
                }
              }}
              onInputChange={(_, value) => handleItemSearch(value)}
              options={itemOptions}
              getOptionLabel={(option) => option.TypeName}
              isOptionEqualToValue={(option, value) => option.TypeID === value.TypeID}
              loading={itemSearchLoading}
              disabled={!!selectedOrder}
              renderOption={(props, option) => (
                <Box component="li" {...props} sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
                  <Avatar
                    src={getItemIconUrl(option.TypeID, 32)}
                    alt={option.TypeName}
                    sx={{ width: 32, height: 32 }}
                    variant="square"
                  />
                  <Typography>{option.TypeName}</Typography>
                </Box>
              )}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="Item Name"
                  placeholder="Start typing to search..."
                  required
                  helperText={selectedOrder ? "Cannot change item type" : "Search for an item by name"}
                  InputProps={{
                    ...params.InputProps,
                    startAdornment: selectedItem ? (
                      <>
                        <Avatar
                          src={getItemIconUrl(selectedItem.TypeID, 32)}
                          alt={selectedItem.TypeName}
                          sx={{ width: 24, height: 24, mr: 1 }}
                          variant="square"
                        />
                        {params.InputProps.startAdornment}
                      </>
                    ) : params.InputProps.startAdornment,
                    endAdornment: (
                      <>
                        {itemSearchLoading ? <CircularProgress color="inherit" size={20} /> : null}
                        {params.InputProps.endAdornment}
                      </>
                    ),
                  }}
                />
              )}
              fullWidth
            />
            <Autocomplete
              value={selectedStation}
              onChange={(_, newValue) => {
                setSelectedStation(newValue);
                if (newValue) {
                  setFormData({ ...formData, locationId: newValue.stationId });
                } else {
                  setFormData({ ...formData, locationId: 0 });
                }
              }}
              onInputChange={(_, value) => handleStationSearch(value)}
              options={stationOptions}
              getOptionLabel={(option) => option.name}
              isOptionEqualToValue={(option, value) => option.stationId === value.stationId}
              loading={stationSearchLoading}
              renderOption={(props, option) => (
                <Box component="li" {...props}>
                  <Box>
                    <Typography variant="body2">{option.name}</Typography>
                    <Typography variant="caption" color="text.secondary">{option.solarSystemName}</Typography>
                  </Box>
                </Box>
              )}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="Station"
                  placeholder="Search for a station..."
                  required
                  helperText="Where you want items delivered"
                  InputProps={{
                    ...params.InputProps,
                    endAdornment: (
                      <>
                        {stationSearchLoading ? <CircularProgress color="inherit" size={20} /> : null}
                        {params.InputProps.endAdornment}
                      </>
                    ),
                  }}
                />
              )}
              fullWidth
            />
            <TextField
              label="Quantity Desired"
              type="number"
              value={formData.quantityDesired || ''}
              onChange={(e) => setFormData({ ...formData, quantityDesired: parseInt(e.target.value) })}
              fullWidth
              required
            />
            <Box sx={{ display: 'flex', gap: 2 }}>
              <TextField
                label="Min Price Per Unit (ISK)"
                type="number"
                value={formData.minPricePerUnit || ''}
                onChange={(e) => setFormData({ ...formData, minPricePerUnit: parseFloat(e.target.value) || 0 })}
                fullWidth
                helperText="Floor price for auto-fulfill (optional)"
              />
              <TextField
                label="Max Price Per Unit (ISK)"
                type="number"
                value={formData.maxPricePerUnit || ''}
                onChange={(e) => setFormData({ ...formData, maxPricePerUnit: parseFloat(e.target.value) || 0 })}
                fullWidth
                required
                helperText="Maximum you're willing to pay"
              />
            </Box>
            <TextField
              label="Notes"
              multiline
              rows={3}
              value={formData.notes || ''}
              onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
              fullWidth
              placeholder="Optional notes about this buy order..."
            />
            {formData.quantityDesired && formData.maxPricePerUnit !== undefined && (
              <Typography variant="body2" color="text.secondary">
                Total Budget: {formatISK(formData.quantityDesired * formData.maxPricePerUnit)}
              </Typography>
            )}
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>Cancel</Button>
          <Button onClick={handleSave} variant="contained">
            {selectedOrder ? 'Update' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Snackbar for notifications */}
      <Snackbar
        open={snackbarOpen}
        autoHideDuration={6000}
        onClose={() => setSnackbarOpen(false)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
      >
        <Alert
          onClose={() => setSnackbarOpen(false)}
          severity={snackbarSeverity}
          sx={{ width: '100%' }}
        >
          {snackbarMessage}
        </Alert>
      </Snackbar>
    </Container>
  );
}
