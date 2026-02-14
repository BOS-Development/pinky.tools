import { useState, useMemo, useEffect, useRef } from 'react';
import { useSession } from "next-auth/react";
import Navbar from "@industry-tool/components/Navbar";
import Loading from "@industry-tool/components/loading";
import Container from '@mui/material/Container';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import SearchIcon from '@mui/icons-material/Search';
import AttachMoneyIcon from '@mui/icons-material/AttachMoney';
import Button from '@mui/material/Button';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import OpenInNewIcon from '@mui/icons-material/OpenInNew';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';

export type StockpileItem = {
  name: string;
  typeId: number;
  quantity: number;
  volume: number;
  ownerType: string;
  ownerName: string;
  ownerId: number;
  desiredQuantity: number;
  stockpileDelta: number;
  deficitValue: number;
  structureName: string;
  solarSystem: string;
  region: string;
  containerName?: string;
};

export type StockpilesResponse = {
  items: StockpileItem[];
};

export default function StockpilesList() {
  const { data: session } = useSession();
  const [stockpileItems, setStockpileItems] = useState<StockpileItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [snackbarOpen, setSnackbarOpen] = useState(false);
  const [snackbarMessage, setSnackbarMessage] = useState('');
  const [creatingAppraisal, setCreatingAppraisal] = useState(false);
  const hasFetchedRef = useRef(false);

  // Fetch stockpile deficits on mount (only once)
  useEffect(() => {
    if (session && !hasFetchedRef.current) {
      hasFetchedRef.current = true;
      fetchStockpileDeficits();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [session]);

  const fetchStockpileDeficits = async () => {
    if (!session) return;

    setLoading(true);
    try {
      const response = await fetch('/api/stockpiles/deficits');
      if (response.ok) {
        const data: StockpilesResponse = await response.json();
        setStockpileItems(data.items || []);
      }
    } finally {
      setLoading(false);
    }
  };

  // Filter items based on search
  const filteredItems = useMemo(() => {
    if (!searchQuery) return stockpileItems;

    const query = searchQuery.toLowerCase();
    return stockpileItems.filter(
      (item) =>
        item.name.toLowerCase().includes(query) ||
        item.structureName.toLowerCase().includes(query) ||
        item.solarSystem.toLowerCase().includes(query) ||
        item.region.toLowerCase().includes(query) ||
        item.containerName?.toLowerCase().includes(query)
    );
  }, [stockpileItems, searchQuery]);

  // Calculate totals
  const totalDeficit = useMemo(() => {
    return filteredItems.reduce((sum, item) => {
      return sum + Math.abs(item.stockpileDelta);
    }, 0);
  }, [filteredItems]);

  const totalVolume = useMemo(() => {
    return filteredItems.reduce((sum, item) => {
      const deficit = Math.abs(item.stockpileDelta);
      // item.volume is total volume (per-unit × quantity), so divide by quantity to get per-unit volume
      const perUnitVolume = item.quantity > 0 ? item.volume / item.quantity : 0;
      return sum + (deficit * perUnitVolume);
    }, 0);
  }, [filteredItems]);

  const totalDeficitISK = useMemo(() => {
    return filteredItems.reduce((sum, item) => {
      return sum + item.deficitValue;
    }, 0);
  }, [filteredItems]);

  const handleCopyForJanice = async () => {
    // Format items as "ItemName quantity" for Janice
    const janiceText = filteredItems
      .map((item) => `${item.name} ${Math.abs(item.stockpileDelta)}`)
      .join('\n');

    try {
      await navigator.clipboard.writeText(janiceText);
      setSnackbarMessage('Copied to clipboard! Paste into Janice for appraisal.');
      setSnackbarOpen(true);
    } catch (err) {
      setSnackbarMessage('Failed to copy to clipboard');
      setSnackbarOpen(true);
    }
  };

  const handleOpenJanice = async () => {
    if (!session) return;

    // Format items as "ItemName quantity" for Janice
    const janiceText = filteredItems
      .map((item) => `${item.name} ${Math.abs(item.stockpileDelta)}`)
      .join('\n');

    setCreatingAppraisal(true);
    try {
      // POST to our Next.js API route which proxies to the backend
      const response = await fetch('/api/janice/appraisal', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          items: janiceText,
        }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        console.error('Janice API error response:', response.status, errorText);
        throw new Error(`Failed to create Janice appraisal: ${response.status} - ${errorText}`);
      }

      const data = await response.json();
      console.log('Janice API response:', data);

      // Open the appraisal in a new tab
      if (data.code) {
        window.open(`https://janice.e-351.com/a/${data.code}`, '_blank');
        setSnackbarMessage('Janice appraisal created and opened!');
        setSnackbarOpen(true);
      } else {
        throw new Error('Janice response missing appraisal code');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setSnackbarMessage(`Failed: ${errorMessage}`);
      setSnackbarOpen(true);
      console.error('Janice API error:', err);
    } finally {
      setCreatingAppraisal(false);
    }
  };

  if (!session) {
    return null;
  }

  if (loading) {
    return <Loading />;
  }

  return (
    <>
      <Navbar />
      <Container maxWidth={false} sx={{ mt: 4, mb: 4 }}>
        {/* Sticky Header Section */}
        <Box
          sx={{
            position: 'sticky',
            top: 64,
            zIndex: 100,
            backgroundColor: 'background.default',
            pb: 2,
          }}
        >
          <Typography variant="h4" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1, pt: 0.5 }}>
            <WarningAmberIcon fontSize="large" color="error" />
            Stockpiles Needing Replenishment
          </Typography>

          {/* Summary Stats */}
          <Box sx={{ display: 'flex', gap: 2, mb: 3 }}>
          <Card sx={{ flex: 1 }}>
            <CardContent>
              <Typography variant="h6" color="text.secondary" gutterBottom>
                Items Below Target
              </Typography>
              <Typography variant="h3">{filteredItems.length}</Typography>
            </CardContent>
          </Card>
          <Card sx={{ flex: 1 }}>
            <CardContent>
              <Typography variant="h6" color="text.secondary" gutterBottom>
                Total Deficit
              </Typography>
              <Typography variant="h3" color="error.main">
                {totalDeficit.toLocaleString()}
              </Typography>
            </CardContent>
          </Card>
          <Card sx={{ flex: 1 }}>
            <CardContent>
              <Typography variant="h6" color="text.secondary" gutterBottom>
                Total Volume
              </Typography>
              <Typography variant="h3">
                {totalVolume.toLocaleString(undefined, { maximumFractionDigits: 2 })} m³
              </Typography>
            </CardContent>
          </Card>
          <Card sx={{ flex: 1 }}>
            <CardContent>
              <Typography variant="h6" color="text.secondary" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <AttachMoneyIcon color="success" />
                Total Cost (ISK)
              </Typography>
              <Typography variant="h3" color="error.main">
                {totalDeficitISK.toLocaleString(undefined, { maximumFractionDigits: 0 })}
              </Typography>
            </CardContent>
          </Card>
        </Box>

        {/* Actions */}
        <Box sx={{ display: 'flex', gap: 2, mb: 2 }}>
          <Button
            variant="outlined"
            startIcon={<ContentCopyIcon />}
            onClick={handleCopyForJanice}
            disabled={filteredItems.length === 0}
          >
            Copy for Janice
          </Button>
          <Button
            variant="contained"
            startIcon={<OpenInNewIcon />}
            onClick={handleOpenJanice}
            disabled={filteredItems.length === 0 || creatingAppraisal}
          >
            {creatingAppraisal ? 'Creating...' : 'Create Janice Appraisal'}
          </Button>
        </Box>

        {/* Search */}
        <Box sx={{ mb: 2 }}>
          <TextField
            fullWidth
            size="small"
            placeholder="Search items, structures, or locations..."
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

        {/* Items Table */}
        {filteredItems.length === 0 ? (
          <Card>
            <CardContent>
              <Typography variant="h6" align="center" color="text.secondary">
                {stockpileItems.length === 0
                  ? 'No stockpiles need replenishment! 🎉'
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
                      <TableCell>Structure</TableCell>
                      <TableCell>Location</TableCell>
                      <TableCell>Container</TableCell>
                      <TableCell align="right">Current</TableCell>
                      <TableCell align="right">Target</TableCell>
                      <TableCell align="right">Deficit</TableCell>
                      <TableCell align="right">Cost (ISK)</TableCell>
                      <TableCell>Owner</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {filteredItems.map((item, idx) => (
                      <TableRow
                        key={idx}
                        hover
                        sx={{
                          '&:nth-of-type(odd)': {
                            backgroundColor: 'action.hover',
                          },
                          borderLeft: '4px solid #d32f2f',
                        }}
                      >
                        <TableCell sx={{ fontWeight: 600 }}>{item.name}</TableCell>
                        <TableCell>{item.structureName}</TableCell>
                        <TableCell>{item.solarSystem}, {item.region}</TableCell>
                        <TableCell>{item.containerName || '-'}</TableCell>
                        <TableCell align="right">{item.quantity.toLocaleString()}</TableCell>
                        <TableCell align="right">{item.desiredQuantity.toLocaleString()}</TableCell>
                        <TableCell align="right">
                          <Typography variant="body2" sx={{ color: 'error.main', fontWeight: 600 }}>
                            {item.stockpileDelta.toLocaleString()}
                          </Typography>
                        </TableCell>
                        <TableCell align="right">
                          <Typography variant="body2" sx={{ color: 'error.main', fontWeight: 600 }}>
                            {item.deficitValue.toLocaleString(undefined, { maximumFractionDigits: 0 })}
                          </Typography>
                        </TableCell>
                        <TableCell>{item.ownerName}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            </CardContent>
          </Card>
        )}
      </Container>

      {/* Snackbar for notifications */}
      <Snackbar
        open={snackbarOpen}
        autoHideDuration={3000}
        onClose={() => setSnackbarOpen(false)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert onClose={() => setSnackbarOpen(false)} severity="success" sx={{ width: '100%' }}>
          {snackbarMessage}
        </Alert>
      </Snackbar>
    </>
  );
}
