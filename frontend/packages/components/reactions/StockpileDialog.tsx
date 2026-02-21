import { useState } from 'react';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TextField from '@mui/material/TextField';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import CircularProgress from '@mui/material/CircularProgress';
import Alert from '@mui/material/Alert';
import { ShoppingItem, StockpileMarker } from "@industry-tool/client/data/models";
import { formatNumber } from "@industry-tool/utils/formatting";
import { AssetOwner } from "@industry-tool/utils/assetAggregation";

type Props = {
  open: boolean;
  onClose: () => void;
  shoppingList: ShoppingItem[];
  locationId: number;
  locationName: string;
  owners: AssetOwner[];
};

export default function StockpileDialog({ open, onClose, shoppingList, locationId, locationName, owners }: Props) {
  const [selectedOwner, setSelectedOwner] = useState('');
  const [quantities, setQuantities] = useState<Record<number, string>>(() => {
    const initial: Record<number, string> = {};
    for (const item of shoppingList) {
      initial[item.type_id] = item.quantity.toString();
    }
    return initial;
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const ownerKey = (o: AssetOwner) => `${o.ownerType}:${o.ownerId}`;
  const parseOwnerKey = (key: string) => {
    const [ownerType, ownerId] = key.split(':');
    return { ownerType, ownerId: parseInt(ownerId, 10) };
  };

  const handleQuantityChange = (typeId: number, value: string) => {
    setQuantities(prev => ({ ...prev, [typeId]: value }));
  };

  const handleSave = async () => {
    if (!selectedOwner) return;

    const { ownerType, ownerId } = parseOwnerKey(selectedOwner);
    setSaving(true);
    setError(null);

    try {
      for (const item of shoppingList) {
        const qty = parseInt(quantities[item.type_id]?.replace(/,/g, '') || '0', 10);
        if (qty <= 0) continue;

        const marker: StockpileMarker = {
          userId: 0,
          typeId: item.type_id,
          ownerType,
          ownerId,
          locationId,
          desiredQuantity: qty,
        };

        const res = await fetch('/api/stockpiles/upsert', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(marker),
        });

        if (!res.ok) {
          throw new Error(`Failed to set stockpile for ${item.name}`);
        }
      }

      setSuccess(true);
      setTimeout(() => {
        onClose();
        setSuccess(false);
      }, 1000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save stockpile targets');
    } finally {
      setSaving(false);
    }
  };

  const itemsWithQty = shoppingList.filter(item => {
    const qty = parseInt(quantities[item.type_id]?.replace(/,/g, '') || '0', 10);
    return qty > 0;
  });

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>Set Stockpile Targets</DialogTitle>
      <DialogContent>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: 1 }}>
          <Typography variant="body2" color="text.secondary">
            Set desired stockpile quantities at <strong>{locationName}</strong>. Items with quantity 0 will be skipped.
          </Typography>

          <FormControl size="small" fullWidth>
            <InputLabel>Owner</InputLabel>
            <Select
              value={selectedOwner}
              label="Owner"
              onChange={(e) => setSelectedOwner(e.target.value)}
            >
              {owners.map(o => (
                <MenuItem key={ownerKey(o)} value={ownerKey(o)}>
                  {o.ownerName} ({o.ownerType})
                </MenuItem>
              ))}
            </Select>
          </FormControl>

          <Box sx={{ display: 'flex', gap: 1 }}>
            <Button
              size="small"
              onClick={() => {
                const cleared: Record<number, string> = {};
                for (const item of shoppingList) cleared[item.type_id] = '0';
                setQuantities(cleared);
              }}
            >
              Clear All
            </Button>
            <Button
              size="small"
              onClick={() => {
                const reset: Record<number, string> = {};
                for (const item of shoppingList) reset[item.type_id] = item.quantity.toString();
                setQuantities(reset);
              }}
            >
              Reset All
            </Button>
          </Box>

          {error && <Alert severity="error">{error}</Alert>}
          {success && <Alert severity="success">Stockpile targets saved!</Alert>}

          <TableContainer sx={{ maxHeight: 400 }}>
            <Table size="small" stickyHeader sx={{ '& th': { backgroundColor: '#0f1219', fontWeight: 'bold' } }}>
              <TableHead>
                <TableRow>
                  <TableCell>Material</TableCell>
                  <TableCell align="right">Shopping List</TableCell>
                  <TableCell align="right">Desired Quantity</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {shoppingList.map((item) => (
                  <TableRow key={item.type_id}>
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
                    <TableCell align="right" sx={{ width: 160 }}>
                      <TextField
                        size="small"
                        type="number"
                        value={quantities[item.type_id] || ''}
                        onChange={(e) => handleQuantityChange(item.type_id, e.target.value)}
                        inputProps={{ min: 0 }}
                        sx={{ width: 140 }}
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </Box>
      </DialogContent>
      <DialogActions>
        <Typography variant="body2" color="text.secondary" sx={{ mr: 'auto', ml: 1 }}>
          {itemsWithQty.length} of {shoppingList.length} items will be set
        </Typography>
        <Button onClick={onClose} disabled={saving}>Cancel</Button>
        <Button
          onClick={handleSave}
          variant="contained"
          disabled={saving || !selectedOwner}
          startIcon={saving ? <CircularProgress size={16} /> : undefined}
        >
          {saving ? 'Saving...' : 'Set Targets'}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
