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
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Button from '@mui/material/Button';
import ButtonGroup from '@mui/material/ButtonGroup';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import CancelIcon from '@mui/icons-material/Cancel';
import AssignmentIcon from '@mui/icons-material/Assignment';

type PurchaseTransaction = {
  id: number;
  forSaleItemId: number;
  buyerUserId: number;
  sellerUserId: number;
  typeId: number;
  typeName: string;
  quantityPurchased: number;
  pricePerUnit: number;
  totalPrice: number;
  status: string;
  transactionNotes?: string;
  buyOrderId?: number;
  isAutoFulfilled: boolean;
  purchasedAt: string;
};

export default function PurchaseHistory() {
  const { data: session } = useSession();
  const [activeTab, setActiveTab] = useState(0);
  const [buyerHistory, setBuyerHistory] = useState<PurchaseTransaction[]>([]);
  const [sellerHistory, setSellerHistory] = useState<PurchaseTransaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity?: 'success' | 'error' }>({
    open: false,
    message: '',
    severity: 'success',
  });

  useEffect(() => {
    if (session) {
      fetchHistory();
    }
  }, [session]);

  const fetchHistory = async () => {
    setLoading(true);
    try {
      const [buyerResponse, sellerResponse] = await Promise.all([
        fetch('/api/purchases/buyer'),
        fetch('/api/purchases/seller'),
      ]);

      if (buyerResponse.ok) {
        const buyerData = await buyerResponse.json();
        setBuyerHistory(buyerData || []);
      }

      if (sellerResponse.ok) {
        const sellerData = await sellerResponse.json();
        setSellerHistory(sellerData || []);
      }
    } catch (error) {
      console.error('Failed to fetch purchase history:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getStatusColor = (status: string): 'warning' | 'info' | 'success' | 'error' | 'default' => {
    switch (status) {
      case 'pending':
        return 'warning';
      case 'contract_created':
        return 'info';
      case 'completed':
        return 'success';
      case 'cancelled':
        return 'error';
      default:
        return 'default';
    }
  };

  const handleMarkContractCreated = async (purchaseId: number) => {
    try {
      const response = await fetch(`/api/purchases/${purchaseId}/mark-contract-created`, {
        method: 'POST',
      });

      if (response.ok) {
        await fetchHistory();
        setSnackbar({ open: true, message: 'Contract marked as created', severity: 'success' });
      } else {
        const error = await response.json();
        setSnackbar({ open: true, message: error.error || 'Failed to mark contract created', severity: 'error' });
      }
    } catch (error) {
      console.error('Failed to mark contract created:', error);
      setSnackbar({ open: true, message: 'Failed to mark contract created', severity: 'error' });
    }
  };

  const handleCompletePurchase = async (purchaseId: number) => {
    try {
      const response = await fetch(`/api/purchases/${purchaseId}/complete`, {
        method: 'POST',
      });

      if (response.ok) {
        await fetchHistory();
        setSnackbar({ open: true, message: 'Purchase marked as completed', severity: 'success' });
      } else {
        const error = await response.json();
        setSnackbar({ open: true, message: error.error || 'Failed to complete purchase', severity: 'error' });
      }
    } catch (error) {
      console.error('Failed to complete purchase:', error);
      setSnackbar({ open: true, message: 'Failed to complete purchase', severity: 'error' });
    }
  };

  const handleCancelPurchase = async (purchaseId: number) => {
    if (!confirm('Are you sure you want to cancel this purchase? The quantity will be restored to the listing.')) {
      return;
    }

    try {
      const response = await fetch(`/api/purchases/${purchaseId}/cancel`, {
        method: 'POST',
      });

      if (response.ok) {
        await fetchHistory();
        setSnackbar({ open: true, message: 'Purchase cancelled successfully', severity: 'success' });
      } else {
        const error = await response.json();
        setSnackbar({ open: true, message: error.error || 'Failed to cancel purchase', severity: 'error' });
      }
    } catch (error) {
      console.error('Failed to cancel purchase:', error);
      setSnackbar({ open: true, message: 'Failed to cancel purchase', severity: 'error' });
    }
  };

  const renderTransactionsTable = (transactions: PurchaseTransaction[], isBuyer: boolean) => {
    if (transactions.length === 0) {
      return (
        <Paper sx={{ p: 4, textAlign: 'center' }}>
          <Typography variant="h6" color="text.secondary">
            No {isBuyer ? 'purchases' : 'sales'} yet
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
            {isBuyer
              ? 'Browse the marketplace to make your first purchase.'
              : 'List items for sale to start selling.'}
          </Typography>
        </Paper>
      );
    }

    return (
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Date</TableCell>
              <TableCell>Item</TableCell>
              <TableCell align="right">Quantity</TableCell>
              <TableCell align="right">Price per Unit</TableCell>
              <TableCell align="right">Total Price</TableCell>
              <TableCell>Status</TableCell>
              {transactions.some(t => t.transactionNotes) && <TableCell>Notes</TableCell>}
              <TableCell align="center">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {transactions.map((transaction) => (
              <TableRow key={transaction.id}>
                <TableCell>{formatDate(transaction.purchasedAt)}</TableCell>
                <TableCell>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                    {transaction.typeName}
                    {transaction.isAutoFulfilled && (
                      <Chip
                        label="Auto"
                        size="small"
                        sx={{
                          fontSize: '0.65rem',
                          fontWeight: 600,
                          height: 20,
                          background: 'rgba(16, 185, 129, 0.15)',
                          color: '#10b981',
                          border: '1px solid rgba(16, 185, 129, 0.3)',
                        }}
                      />
                    )}
                  </Box>
                </TableCell>
                <TableCell align="right">{transaction.quantityPurchased.toLocaleString()}</TableCell>
                <TableCell align="right">{transaction.pricePerUnit.toLocaleString()} ISK</TableCell>
                <TableCell align="right">
                  <Typography
                    variant="body2"
                    sx={{
                      fontWeight: 600,
                      color: isBuyer ? 'error.main' : 'success.main',
                    }}
                  >
                    {isBuyer ? '-' : '+'}
                    {transaction.totalPrice.toLocaleString()} ISK
                  </Typography>
                </TableCell>
                <TableCell>
                  <Chip
                    label={transaction.status.replace('_', ' ')}
                    size="small"
                    color={getStatusColor(transaction.status)}
                  />
                </TableCell>
                {transactions.some(t => t.transactionNotes) && (
                  <TableCell>
                    {transaction.transactionNotes && (
                      <Typography variant="caption" color="text.secondary">
                        {transaction.transactionNotes}
                      </Typography>
                    )}
                  </TableCell>
                )}
                <TableCell align="center">
                  <ButtonGroup size="small" variant="outlined">
                    {/* Buyer actions */}
                    {isBuyer && transaction.status === 'contract_created' && (
                      <Button
                        onClick={() => handleCompletePurchase(transaction.id)}
                        color="success"
                        startIcon={<CheckCircleIcon />}
                      >
                        Complete
                      </Button>
                    )}

                    {/* Seller actions */}
                    {!isBuyer && transaction.status === 'pending' && (
                      <Button
                        onClick={() => handleMarkContractCreated(transaction.id)}
                        color="info"
                        startIcon={<AssignmentIcon />}
                      >
                        Mark Contract Created
                      </Button>
                    )}

                    {/* Cancel action (both parties) */}
                    {(transaction.status === 'pending' || transaction.status === 'contract_created') && (
                      <Button
                        onClick={() => handleCancelPurchase(transaction.id)}
                        color="error"
                        startIcon={<CancelIcon />}
                      >
                        Cancel
                      </Button>
                    )}

                    {/* No actions for completed/cancelled */}
                    {(transaction.status === 'completed' || transaction.status === 'cancelled') && (
                      <Typography variant="caption" color="text.secondary">
                        -
                      </Typography>
                    )}
                  </ButtonGroup>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    );
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
      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 3 }}>
        <Tabs value={activeTab} onChange={(_, newValue) => setActiveTab(newValue)}>
          <Tab label={`My Purchases (${buyerHistory.length})`} />
          <Tab label={`My Sales (${sellerHistory.length})`} />
        </Tabs>
      </Box>

      {activeTab === 0 && renderTransactionsTable(buyerHistory, true)}
      {activeTab === 1 && renderTransactionsTable(sellerHistory, false)}

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
