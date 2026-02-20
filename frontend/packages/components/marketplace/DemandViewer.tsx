import { useState, useEffect, useRef } from 'react';
import { useSession } from "next-auth/react";
import Container from '@mui/material/Container';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import SearchIcon from '@mui/icons-material/Search';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import Chip from '@mui/material/Chip';
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
  notes?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

export default function DemandViewer() {
  const { data: session } = useSession();
  const [demand, setDemand] = useState<BuyOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [snackbarOpen, setSnackbarOpen] = useState(false);
  const [snackbarMessage, setSnackbarMessage] = useState('');
  const [snackbarSeverity, setSnackbarSeverity] = useState<'success' | 'error'>('success');
  const hasFetchedRef = useRef(false);

  useEffect(() => {
    if (session && !hasFetchedRef.current) {
      hasFetchedRef.current = true;
      fetchDemand();
    }
  }, [session]);

  const fetchDemand = async () => {
    try {
      const response = await fetch('/api/buy-orders/demand');
      if (!response.ok) throw new Error('Failed to fetch demand');
      const data = await response.json();
      setDemand(data);
    } catch (error) {
      console.error('Error fetching demand:', error);
      showSnackbar('Failed to load demand data', 'error');
    } finally {
      setLoading(false);
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

  const filteredDemand = demand.filter((order) =>
    order.typeName.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Group orders by item type and aggregate
  const aggregatedDemand = filteredDemand.reduce((acc, order) => {
    const key = order.typeId;
    if (!acc[key]) {
      acc[key] = {
        typeId: order.typeId,
        typeName: order.typeName,
        totalQuantity: 0,
        maxPrice: 0,
        orderCount: 0,
        orders: [],
      };
    }
    acc[key].totalQuantity += order.quantityDesired;
    acc[key].maxPrice = Math.max(acc[key].maxPrice, order.minPricePerUnit);
    acc[key].orderCount += 1;
    acc[key].orders.push(order);
    return acc;
  }, {} as Record<number, {
    typeId: number;
    typeName: string;
    totalQuantity: number;
    maxPrice: number;
    orderCount: number;
    orders: BuyOrder[];
  }>);

  const aggregatedData = Object.values(aggregatedDemand).sort(
    (a, b) => b.totalQuantity - a.totalQuantity
  );

  if (loading) {
    return <Loading />;
  }

  return (
    <Container maxWidth="xl">
      <Box sx={{ my: 4 }}>
        <Card>
          <CardContent>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                <TrendingUpIcon fontSize="large" color="primary" />
                <Box>
                  <Typography variant="h5" component="h2">
                    Market Demand
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Buy orders from your contacts
                  </Typography>
                </Box>
              </Box>
              <TextField
                placeholder="Search items..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                size="small"
                InputProps={{
                  startAdornment: (
                    <InputAdornment position="start">
                      <SearchIcon />
                    </InputAdornment>
                  ),
                }}
              />
            </Box>

            {demand.length === 0 ? (
              <Typography variant="body1" color="text.secondary" sx={{ textAlign: 'center', py: 4 }}>
                No active buy orders from your contacts yet.
                <br />
                When your contacts create buy orders, they'll appear here!
              </Typography>
            ) : (
              <>
                {/* Aggregated Summary */}
                <Typography variant="h6" gutterBottom sx={{ mt: 2 }}>
                  Aggregated Demand
                </Typography>
                <TableContainer component={Paper} sx={{ mb: 4 }}>
                  <Table>
                    <TableHead>
                      <TableRow>
                        <TableCell>Item</TableCell>
                        <TableCell align="right">Total Quantity Wanted</TableCell>
                        <TableCell align="right">Highest Floor Price</TableCell>
                        <TableCell align="right">Potential Revenue</TableCell>
                        <TableCell align="center">Number of Orders</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {aggregatedData.map((item) => (
                        <TableRow key={item.typeId}>
                          <TableCell>
                            <strong>{item.typeName}</strong>
                          </TableCell>
                          <TableCell align="right">{formatNumber(item.totalQuantity)}</TableCell>
                          <TableCell align="right">{formatISK(item.maxPrice)}</TableCell>
                          <TableCell align="right">
                            <Typography variant="body2" color="success.main" fontWeight="bold">
                              {formatISK(item.totalQuantity * item.maxPrice)}
                            </Typography>
                          </TableCell>
                          <TableCell align="center">
                            <Chip label={item.orderCount} size="small" color="primary" />
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>

                {/* Detailed Orders */}
                <Typography variant="h6" gutterBottom>
                  Individual Buy Orders
                </Typography>
                <TableContainer component={Paper}>
                  <Table>
                    <TableHead>
                      <TableRow>
                        <TableCell>Item</TableCell>
                        <TableCell>Location</TableCell>
                        <TableCell align="right">Quantity</TableCell>
                        <TableCell align="right">Min Price/Unit</TableCell>
                        <TableCell align="right">Est. Revenue</TableCell>
                        <TableCell>Notes</TableCell>
                        <TableCell>Created</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {filteredDemand.map((order) => (
                        <TableRow key={order.id}>
                          <TableCell>{order.typeName}</TableCell>
                          <TableCell>{order.locationName || '-'}</TableCell>
                          <TableCell align="right">{formatNumber(order.quantityDesired)}</TableCell>
                          <TableCell align="right">{formatISK(order.minPricePerUnit)}</TableCell>
                          <TableCell align="right">
                            {formatISK(order.quantityDesired * order.minPricePerUnit)}
                          </TableCell>
                          <TableCell>{order.notes || '-'}</TableCell>
                          <TableCell>{formatDate(order.createdAt)}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
              </>
            )}
          </CardContent>
        </Card>
      </Box>

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
