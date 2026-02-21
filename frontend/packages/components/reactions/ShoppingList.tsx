import { useState } from 'react';
import Box from '@mui/material/Box';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableSortLabel from '@mui/material/TableSortLabel';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import CircularProgress from '@mui/material/CircularProgress';
import Autocomplete from '@mui/material/Autocomplete';
import TextField from '@mui/material/TextField';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import InventoryIcon from '@mui/icons-material/Inventory';
import { PlanResponse, AssetsResponse } from "@industry-tool/client/data/models";
import { formatISK, formatNumber } from "@industry-tool/utils/formatting";
import { aggregateAssetsByTypeId, getUniqueOwners } from "@industry-tool/utils/assetAggregation";
import StockpileDialog from './StockpileDialog';

type Props = {
  planData: PlanResponse | null;
  loading: boolean;
  assets: AssetsResponse | null;
  isAuthenticated: boolean;
  stockpileLocationId: number;
  onStockpileLocationChange: (locationId: number) => void;
};

type SortKey = 'name' | 'quantity' | 'inStock' | 'delta' | 'price' | 'cost' | 'volume';
type SortDir = 'asc' | 'desc';

export default function ShoppingList({ planData, loading, assets, isAuthenticated, stockpileLocationId, onStockpileLocationChange }: Props) {
  const [stockpileDialogOpen, setStockpileDialogOpen] = useState(false);
  const [sortKey, setSortKey] = useState<SortKey>('name');
  const [sortDir, setSortDir] = useState<SortDir>('asc');

  const handleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir(prev => prev === 'asc' ? 'desc' : 'asc');
    } else {
      setSortKey(key);
      setSortDir(key === 'name' ? 'asc' : 'desc');
    }
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (!planData || planData.shopping_list.length === 0) {
    return (
      <Typography color="text.secondary" sx={{ py: 4, textAlign: 'center' }}>
        Select reactions in the Pick Reactions tab to generate a shopping list.
      </Typography>
    );
  }

  const locations = assets?.structures ?? [];
  const selectedStructure = locations.find(s => s.id === stockpileLocationId) || null;
  const stockpileMap = selectedStructure ? aggregateAssetsByTypeId(selectedStructure) : null;

  const totalCost = planData.shopping_list.reduce((sum, item) => sum + item.cost, 0);
  const totalVolume = planData.shopping_list.reduce((sum, item) => sum + item.volume, 0);

  const deltaCost = stockpileMap
    ? planData.shopping_list.reduce((sum, item) => {
        const have = stockpileMap.get(item.type_id) || 0;
        const delta = Math.max(0, item.quantity - have);
        return sum + (delta * item.price);
      }, 0)
    : null;

  const copyMultibuy = () => {
    const text = planData.shopping_list
      .map(item => `${item.name} ${item.quantity}`)
      .join('\n');
    navigator.clipboard.writeText(text);
  };

  const copyMultibuyDelta = () => {
    if (!stockpileMap) return;
    const lines = planData.shopping_list
      .map(item => {
        const have = stockpileMap.get(item.type_id) || 0;
        const delta = Math.max(0, item.quantity - have);
        return delta > 0 ? `${item.name} ${delta}` : null;
      })
      .filter(Boolean);
    navigator.clipboard.writeText(lines.join('\n'));
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2, flexWrap: 'wrap', gap: 1 }}>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.5 }}>
          <Typography variant="body2" color="text.secondary">
            {planData.shopping_list.length} items | Total: {formatISK(totalCost)} | Volume: {formatNumber(totalVolume, 1)} m3
          </Typography>
          {deltaCost !== null && (
            <Typography variant="body2" color="success.main">
              Delta Cost: {formatISK(deltaCost)}
            </Typography>
          )}
        </Box>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          {isAuthenticated && (
            <Autocomplete
              size="small"
              sx={{ minWidth: 250 }}
              options={locations}
              getOptionLabel={(option) => option.name}
              value={selectedStructure}
              onChange={(_e, value) => onStockpileLocationChange(value?.id ?? 0)}
              renderInput={(params) => <TextField {...params} label="Stockpile Location" />}
              isOptionEqualToValue={(option, value) => option.id === value.id}
            />
          )}
          <Button
            variant="outlined"
            size="small"
            startIcon={<ContentCopyIcon />}
            onClick={copyMultibuy}
          >
            Copy Multibuy
          </Button>
          {stockpileMap && (
            <Button
              variant="outlined"
              size="small"
              startIcon={<ContentCopyIcon />}
              onClick={copyMultibuyDelta}
              color="success"
            >
              Copy Delta
            </Button>
          )}
          {selectedStructure && (
            <Button
              variant="outlined"
              size="small"
              startIcon={<InventoryIcon />}
              onClick={() => setStockpileDialogOpen(true)}
              color="warning"
            >
              Set Stockpile
            </Button>
          )}
        </Box>
      </Box>

      <TableContainer>
        <Table size="small" sx={{ '& th': { backgroundColor: '#0f1219', fontWeight: 'bold' } }}>
          <TableHead>
            <TableRow>
              <TableCell>
                <TableSortLabel active={sortKey === 'name'} direction={sortKey === 'name' ? sortDir : 'asc'} onClick={() => handleSort('name')}>
                  Material
                </TableSortLabel>
              </TableCell>
              <TableCell align="right">
                <TableSortLabel active={sortKey === 'quantity'} direction={sortKey === 'quantity' ? sortDir : 'desc'} onClick={() => handleSort('quantity')}>
                  Quantity
                </TableSortLabel>
              </TableCell>
              {stockpileMap && (
                <TableCell align="right">
                  <TableSortLabel active={sortKey === 'inStock'} direction={sortKey === 'inStock' ? sortDir : 'desc'} onClick={() => handleSort('inStock')}>
                    In Stock
                  </TableSortLabel>
                </TableCell>
              )}
              {stockpileMap && (
                <TableCell align="right">
                  <TableSortLabel active={sortKey === 'delta'} direction={sortKey === 'delta' ? sortDir : 'desc'} onClick={() => handleSort('delta')}>
                    Delta
                  </TableSortLabel>
                </TableCell>
              )}
              <TableCell align="right">
                <TableSortLabel active={sortKey === 'price'} direction={sortKey === 'price' ? sortDir : 'desc'} onClick={() => handleSort('price')}>
                  Unit Price
                </TableSortLabel>
              </TableCell>
              <TableCell align="right">
                <TableSortLabel active={sortKey === 'cost'} direction={sortKey === 'cost' ? sortDir : 'desc'} onClick={() => handleSort('cost')}>
                  Total Cost
                </TableSortLabel>
              </TableCell>
              <TableCell align="right">
                <TableSortLabel active={sortKey === 'volume'} direction={sortKey === 'volume' ? sortDir : 'desc'} onClick={() => handleSort('volume')}>
                  Volume (m3)
                </TableSortLabel>
              </TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {[...planData.shopping_list]
              .map((item) => {
                const have = stockpileMap ? (stockpileMap.get(item.type_id) || 0) : 0;
                const delta = Math.max(0, item.quantity - have);
                return { item, have, delta };
              })
              .sort((a, b) => {
                let cmp = 0;
                switch (sortKey) {
                  case 'name': cmp = a.item.name.localeCompare(b.item.name); break;
                  case 'quantity': cmp = a.item.quantity - b.item.quantity; break;
                  case 'inStock': cmp = a.have - b.have; break;
                  case 'delta': cmp = a.delta - b.delta; break;
                  case 'price': cmp = a.item.price - b.item.price; break;
                  case 'cost': cmp = a.item.cost - b.item.cost; break;
                  case 'volume': cmp = a.item.volume - b.item.volume; break;
                }
                return sortDir === 'asc' ? cmp : -cmp;
              })
              .map(({ item, have, delta }) => {
              const fulfilled = stockpileMap ? delta === 0 : false;

              return (
                <TableRow
                  key={item.type_id}
                  sx={{ '&:nth-of-type(odd)': { backgroundColor: 'rgba(255,255,255,0.02)' } }}
                >
                  <TableCell>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <img
                        src={`https://images.evetech.net/types/${item.type_id}/icon?size=32`}
                        alt=""
                        width={24}
                        height={24}
                        style={{ borderRadius: 2 }}
                      />
                      {item.name}
                    </Box>
                  </TableCell>
                  <TableCell align="right">{formatNumber(item.quantity)}</TableCell>
                  {stockpileMap && (
                    <TableCell align="right">{formatNumber(have)}</TableCell>
                  )}
                  {stockpileMap && (
                    <TableCell align="right" sx={fulfilled ? { color: '#10b981' } : undefined}>
                      {fulfilled ? (
                        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end', gap: 0.5 }}>
                          <CheckCircleIcon sx={{ fontSize: 16 }} />
                          0
                        </Box>
                      ) : (
                        formatNumber(delta)
                      )}
                    </TableCell>
                  )}
                  <TableCell align="right">{formatISK(item.price)}</TableCell>
                  <TableCell align="right">{formatISK(item.cost)}</TableCell>
                  <TableCell align="right">{formatNumber(item.volume, 1)}</TableCell>
                </TableRow>
              );
            })}
            <TableRow sx={{ '& td': { fontWeight: 'bold', borderTop: '2px solid rgba(255,255,255,0.1)' } }}>
              <TableCell>Total</TableCell>
              <TableCell />
              {stockpileMap && <TableCell />}
              {stockpileMap && <TableCell />}
              <TableCell />
              <TableCell align="right">{formatISK(totalCost)}</TableCell>
              <TableCell align="right">{formatNumber(totalVolume, 1)}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </TableContainer>

      {selectedStructure && (
        <StockpileDialog
          open={stockpileDialogOpen}
          onClose={() => setStockpileDialogOpen(false)}
          shoppingList={planData.shopping_list}
          locationId={selectedStructure.id}
          locationName={selectedStructure.name}
          owners={getUniqueOwners(selectedStructure)}
        />
      )}
    </Box>
  );
}
