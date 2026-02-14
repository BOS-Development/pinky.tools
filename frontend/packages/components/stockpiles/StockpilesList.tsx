import { useState, useMemo } from 'react';
import { AssetsResponse, Asset } from "@industry-tool/client/data/models";
import { useSession } from "next-auth/react";
import Navbar from "@industry-tool/components/Navbar";
import Container from '@mui/material/Container';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import SearchIcon from '@mui/icons-material/Search';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';

export type StockpilesListProps = {
  assets: AssetsResponse;
};

type StockpileItem = Asset & {
  structureName: string;
  location: string;
  containerName?: string;
};

export default function StockpilesList(props: StockpilesListProps) {
  const { data: session } = useSession();
  const { assets } = props;
  const [searchQuery, setSearchQuery] = useState('');

  // Flatten all assets and filter for items needing replenishment
  const stockpileItems = useMemo(() => {
    const items: StockpileItem[] = [];

    assets.structures.forEach((structure) => {
      const addAssets = (assetList: Asset[], containerName?: string) => {
        assetList.forEach((asset) => {
          if (asset.stockpileDelta !== undefined && asset.stockpileDelta < 0) {
            items.push({
              ...asset,
              structureName: structure.name,
              location: `${structure.solarSystem}, ${structure.region}`,
              containerName,
            });
          }
        });
      };

      // Personal hangar assets
      if (structure.hangarAssets) {
        addAssets(structure.hangarAssets);
      }

      // Container assets
      structure.hangarContainers?.forEach((container) => {
        addAssets(container.assets, container.name);
      });

      // Deliveries
      if (structure.deliveries) {
        addAssets(structure.deliveries, 'Deliveries');
      }

      // Asset safety
      if (structure.assetSafety) {
        addAssets(structure.assetSafety, 'Asset Safety');
      }

      // Corporation hangars
      structure.corporationHangers?.forEach((hanger) => {
        addAssets(hanger.assets, hanger.name);
        hanger.hangarContainers?.forEach((container) => {
          addAssets(container.assets, `${hanger.name} - ${container.name}`);
        });
      });
    });

    return items;
  }, [assets]);

  // Filter items based on search
  const filteredItems = useMemo(() => {
    if (!searchQuery) return stockpileItems;

    const query = searchQuery.toLowerCase();
    return stockpileItems.filter(
      (item) =>
        item.name.toLowerCase().includes(query) ||
        item.structureName.toLowerCase().includes(query) ||
        item.location.toLowerCase().includes(query) ||
        item.containerName?.toLowerCase().includes(query)
    );
  }, [stockpileItems, searchQuery]);

  // Calculate totals
  const totalDeficit = useMemo(() => {
    return filteredItems.reduce((sum, item) => {
      return sum + Math.abs(item.stockpileDelta || 0);
    }, 0);
  }, [filteredItems]);

  const totalVolume = useMemo(() => {
    return filteredItems.reduce((sum, item) => {
      const deficit = Math.abs(item.stockpileDelta || 0);
      // item.volume is total volume (per-unit × quantity), so divide by quantity to get per-unit volume
      const perUnitVolume = item.quantity > 0 ? item.volume / item.quantity : 0;
      return sum + (deficit * perUnitVolume);
    }, 0);
  }, [filteredItems]);

  if (!session) {
    return null;
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
                        <TableCell>{item.location}</TableCell>
                        <TableCell>{item.containerName || '-'}</TableCell>
                        <TableCell align="right">{item.quantity.toLocaleString()}</TableCell>
                        <TableCell align="right">{item.desiredQuantity?.toLocaleString()}</TableCell>
                        <TableCell align="right">
                          <Typography variant="body2" sx={{ color: 'error.main', fontWeight: 600 }}>
                            {item.stockpileDelta?.toLocaleString()}
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
    </>
  );
}
