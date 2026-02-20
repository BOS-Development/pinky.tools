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
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Accordion from '@mui/material/Accordion';
import AccordionSummary from '@mui/material/AccordionSummary';
import AccordionDetails from '@mui/material/AccordionDetails';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import AssignmentIcon from '@mui/icons-material/Assignment';
import CancelIcon from '@mui/icons-material/Cancel';
import LocationOnIcon from '@mui/icons-material/LocationOn';
import PersonIcon from '@mui/icons-material/Person';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import Tooltip from '@mui/material/Tooltip';

type PendingSale = {
  id: number;
  forSaleItemId: number;
  buyerUserId: number;
  buyerName: string;
  sellerUserId: number;
  typeId: number;
  typeName: string;
  locationId: number;
  locationName: string;
  quantityPurchased: number;
  pricePerUnit: number;
  totalPrice: number;
  status: string;
  contractKey?: string;
  transactionNotes?: string;
  buyOrderId?: number;
  isAutoFulfilled: boolean;
  purchasedAt: string;
};

type GroupedSale = {
  buyerUserId: number;
  buyerName: string;
  locationId: number;
  locationName: string;
  items: PendingSale[];
  totalValue: number;
  contractKey?: string;
};

export default function PendingSales() {
  const { data: session } = useSession();
  const [pendingSales, setPendingSales] = useState<PendingSale[]>([]);
  const [loading, setLoading] = useState(true);
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity?: 'success' | 'error' }>({
    open: false,
    message: '',
    severity: 'success',
  });

  useEffect(() => {
    if (session) {
      fetchPendingSales();
    }
  }, [session]);

  const fetchPendingSales = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/purchases/pending-sales');

      if (response.ok) {
        const data = await response.json();
        setPendingSales(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch pending sales:', error);
    } finally {
      setLoading(false);
    }
  };

  const groupSales = (): GroupedSale[] => {
    const groups: Map<string, GroupedSale> = new Map();

    pendingSales.forEach((sale) => {
      const key = `${sale.buyerUserId}-${sale.locationId}`;

      if (!groups.has(key)) {
        // Generate contract key immediately for this group
        const contractKey = sale.contractKey || generateContractKey(sale.buyerUserId, sale.locationId);

        groups.set(key, {
          buyerUserId: sale.buyerUserId,
          buyerName: sale.buyerName,
          locationId: sale.locationId,
          locationName: sale.locationName,
          items: [],
          totalValue: 0,
          contractKey: contractKey,
        });
      }

      const group = groups.get(key)!;
      group.items.push(sale);
      group.totalValue += sale.totalPrice;
    });

    return Array.from(groups.values()).sort((a, b) => {
      // Sort by location first, then by buyer name
      if (a.locationName !== b.locationName) {
        return a.locationName.localeCompare(b.locationName);
      }
      return a.buyerName.localeCompare(b.buyerName);
    });
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const generateContractKey = (buyerUserId: number, locationId: number): string => {
    const timestamp = Date.now();
    return `PT-${buyerUserId}-${locationId}-${timestamp}`;
  };

  const handleMarkContractCreated = async (purchaseId: number, contractKey?: string) => {
    try {
      const response = await fetch(`/api/purchases/${purchaseId}/mark-contract-created`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ contractKey }),
      });

      if (response.ok) {
        await fetchPendingSales();
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

  const handleMarkGroupContractCreated = async (group: GroupedSale) => {
    // Use the group's contract key (already generated)
    const contractKey = group.contractKey!;

    try {
      await Promise.all(
        group.items.map(item =>
          fetch(`/api/purchases/${item.id}/mark-contract-created`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ contractKey }),
          })
        )
      );

      await fetchPendingSales();
      setSnackbar({
        open: true,
        message: `Marked ${group.items.length} contract${group.items.length !== 1 ? 's' : ''} as created`,
        severity: 'success'
      });
    } catch (error) {
      console.error('Failed to mark contracts created:', error);
      setSnackbar({ open: true, message: 'Failed to mark contracts created', severity: 'error' });
    }
  };

  const handleCancel = async (purchaseId: number) => {
    if (!confirm('Are you sure you want to cancel this sale? The quantity will be restored to the listing.')) {
      return;
    }

    try {
      const response = await fetch(`/api/purchases/${purchaseId}/cancel`, {
        method: 'POST',
      });

      if (response.ok) {
        await fetchPendingSales();
        setSnackbar({ open: true, message: 'Sale cancelled successfully', severity: 'success' });
      } else {
        const error = await response.json();
        setSnackbar({ open: true, message: error.error || 'Failed to cancel sale', severity: 'error' });
      }
    } catch (error) {
      console.error('Failed to cancel sale:', error);
      setSnackbar({ open: true, message: 'Failed to cancel sale', severity: 'error' });
    }
  };

  const handleCopyBuyerName = async (buyerName: string) => {
    try {
      await navigator.clipboard.writeText(buyerName);
      setSnackbar({ open: true, message: `Copied "${buyerName}" to clipboard`, severity: 'success' });
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
      setSnackbar({ open: true, message: 'Failed to copy to clipboard', severity: 'error' });
    }
  };

  const handleCopyContractKey = async (contractKey: string) => {
    try {
      await navigator.clipboard.writeText(contractKey);
      setSnackbar({ open: true, message: `Copied "${contractKey}" to clipboard`, severity: 'success' });
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
      setSnackbar({ open: true, message: 'Failed to copy to clipboard', severity: 'error' });
    }
  };

  const handleCopyTotal = async (totalValue: number) => {
    try {
      await navigator.clipboard.writeText(totalValue.toString());
      setSnackbar({ open: true, message: `Copied "${totalValue.toLocaleString()} ISK" to clipboard`, severity: 'success' });
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
      setSnackbar({ open: true, message: 'Failed to copy to clipboard', severity: 'error' });
    }
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (pendingSales.length === 0) {
    return (
      <Paper sx={{ p: 4, textAlign: 'center' }}>
        <Typography variant="h6" color="text.secondary">
          No pending sales
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
          When buyers request to purchase your items, they will appear here.
        </Typography>
      </Paper>
    );
  }

  const groupedSales = groupSales();

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Pending Sales ({pendingSales.length} items in {groupedSales.length} groups)
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
        Sales are grouped by purchaser and station. Copy the contract key, create the in-game contract with the key in the description, then mark as "Contract Created".
      </Typography>

      {groupedSales.map((group, index) => (
        <Accordion key={`${group.buyerUserId}-${group.locationId}`} defaultExpanded={index === 0}>
          <AccordionSummary expandIcon={<ExpandMoreIcon />}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, width: '100%' }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, flex: 1, flexWrap: 'wrap' }}>
                <PersonIcon fontSize="small" color="action" />
                <Tooltip title="Click to copy buyer name">
                  <Box
                    sx={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 0.5,
                      cursor: 'pointer',
                      '&:hover .buyer-name': {
                        color: 'primary.main',
                        textDecoration: 'underline',
                      },
                      '&:hover .copy-icon': {
                        color: 'primary.main',
                      },
                    }}
                    onClick={(e) => {
                      e.stopPropagation();
                      handleCopyBuyerName(group.buyerName);
                    }}
                  >
                    <Typography
                      className="buyer-name"
                      variant="subtitle1"
                      sx={{ fontWeight: 600 }}
                    >
                      {group.buyerName}
                    </Typography>
                    <ContentCopyIcon
                      className="copy-icon"
                      fontSize="small"
                      sx={{ fontSize: '1rem', color: 'action.active' }}
                    />
                  </Box>
                </Tooltip>
                <Box sx={{ mx: 1 }}>•</Box>
                <LocationOnIcon fontSize="small" color="action" />
                <Typography variant="subtitle1">
                  {group.locationName}
                </Typography>
                <Box sx={{ mx: 1 }}>•</Box>
                <Tooltip title="Click to copy total value">
                  <Box
                    sx={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 0.5,
                      cursor: 'pointer',
                      '&:hover .total-value': {
                        color: 'success.dark',
                        textDecoration: 'underline',
                      },
                      '&:hover .copy-icon': {
                        color: 'success.dark',
                      },
                    }}
                    onClick={(e) => {
                      e.stopPropagation();
                      handleCopyTotal(group.totalValue);
                    }}
                  >
                    <Typography
                      className="total-value"
                      variant="h6"
                      sx={{ color: 'success.main', fontWeight: 600 }}
                    >
                      {group.totalValue.toLocaleString()} ISK
                    </Typography>
                    <ContentCopyIcon
                      className="copy-icon"
                      fontSize="small"
                      sx={{ fontSize: '1rem', color: 'success.main' }}
                    />
                  </Box>
                </Tooltip>
              </Box>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                <Typography variant="body2" color="text.secondary">
                  {group.items.length} item{group.items.length !== 1 ? 's' : ''}
                </Typography>
              </Box>
            </Box>
          </AccordionSummary>
          <AccordionDetails>
            <Box sx={{ mb: 2, display: 'flex', gap: 2, alignItems: 'flex-start', flexWrap: 'wrap' }}>
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.5 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <Typography variant="body2" color="text.secondary" sx={{ fontWeight: 500 }}>
                    Contract Key:
                  </Typography>
                  <Tooltip title="Click to copy contract key">
                    <Box
                      sx={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 0.5,
                        cursor: 'pointer',
                        px: 1.5,
                        py: 0.75,
                        borderRadius: 1,
                        bgcolor: 'primary.main',
                        color: 'primary.contrastText',
                        '&:hover': {
                          bgcolor: 'primary.dark',
                        },
                      }}
                      onClick={() => handleCopyContractKey(group.contractKey!)}
                    >
                      <Typography variant="body2" sx={{ fontFamily: 'monospace', fontWeight: 700 }}>
                        {group.contractKey}
                      </Typography>
                      <ContentCopyIcon sx={{ fontSize: '1rem' }} />
                    </Box>
                  </Tooltip>
                </Box>
                <Typography variant="caption" color="text.secondary" sx={{ fontStyle: 'italic', ml: 1 }}>
                  Copy this and paste into the in-game contract description
                </Typography>
              </Box>
              <Button
                onClick={() => handleMarkGroupContractCreated(group)}
                variant="contained"
                color="success"
                size="small"
                startIcon={<AssignmentIcon />}
                sx={{ mt: 0.5 }}
              >
                Mark All as Contract Created
              </Button>
            </Box>

            <TableContainer>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Requested</TableCell>
                    <TableCell>Item</TableCell>
                    <TableCell align="right">Quantity</TableCell>
                    <TableCell align="right">Price/Unit</TableCell>
                    <TableCell align="right">Total</TableCell>
                    <TableCell>Status</TableCell>
                    {group.items.some(s => s.transactionNotes) && <TableCell>Notes</TableCell>}
                    <TableCell align="center">Actions</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {group.items.map((sale) => (
                    <TableRow key={sale.id}>
                      <TableCell>{formatDate(sale.purchasedAt)}</TableCell>
                      <TableCell>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                          {sale.typeName}
                          {sale.isAutoFulfilled && (
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
                      <TableCell align="right">{sale.quantityPurchased.toLocaleString()}</TableCell>
                      <TableCell align="right">{sale.pricePerUnit.toLocaleString()} ISK</TableCell>
                      <TableCell align="right">
                        <Typography variant="body2" sx={{ fontWeight: 600, color: 'success.main' }}>
                          {sale.totalPrice.toLocaleString()} ISK
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={sale.status.replace('_', ' ')}
                          size="small"
                          color={sale.status === 'pending' ? 'warning' : 'info'}
                        />
                      </TableCell>
                      {group.items.some(s => s.transactionNotes) && (
                        <TableCell>
                          {sale.transactionNotes && (
                            <Typography variant="caption" color="text.secondary">
                              {sale.transactionNotes}
                            </Typography>
                          )}
                        </TableCell>
                      )}
                      <TableCell align="center">
                        <Button
                          onClick={() => handleCancel(sale.id)}
                          variant="outlined"
                          color="error"
                          size="small"
                          startIcon={<CancelIcon />}
                        >
                          Cancel
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </AccordionDetails>
        </Accordion>
      ))}

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
