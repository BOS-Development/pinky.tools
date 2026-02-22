import { useState, useCallback, useRef } from 'react';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Autocomplete from '@mui/material/Autocomplete';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import CircularProgress from '@mui/material/CircularProgress';
import Alert from '@mui/material/Alert';
import { Asset, StockpileMarker, EveInventoryType } from "@industry-tool/client/data/models";

type Owner = {
  ownerType: string;
  ownerId: number;
  ownerName: string;
};

type Props = {
  open: boolean;
  onClose: () => void;
  onSaved: (asset: Asset) => void;
  locationId: number;
  containerId?: number;
  divisionNumber?: number;
  owners: Owner[];
};

export default function AddStockpileDialog({ open, onClose, onSaved, locationId, containerId, divisionNumber, owners }: Props) {
  const [selectedItem, setSelectedItem] = useState<EveInventoryType | null>(null);
  const [itemOptions, setItemOptions] = useState<EveInventoryType[]>([]);
  const [itemLoading, setItemLoading] = useState(false);
  const [selectedOwner, setSelectedOwner] = useState('');
  const [desiredQuantity, setDesiredQuantity] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Auto-select single owner
  const effectiveOwner = owners.length === 1
    ? `${owners[0].ownerType}:${owners[0].ownerId}`
    : selectedOwner;

  const parseOwnerKey = (key: string) => {
    const [ownerType, ownerId] = key.split(':');
    return { ownerType, ownerId: parseInt(ownerId, 10) };
  };

  const searchItems = useCallback((query: string) => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    if (!query || query.length < 2) {
      setItemOptions([]);
      return;
    }

    debounceRef.current = setTimeout(async () => {
      setItemLoading(true);
      try {
        const res = await fetch(`/api/item-types/search?q=${encodeURIComponent(query)}`);
        if (res.ok) {
          const data: EveInventoryType[] = await res.json();
          setItemOptions(data);
        }
      } finally {
        setItemLoading(false);
      }
    }, 300);
  }, []);

  const handleSave = async () => {
    if (!selectedItem || !effectiveOwner || !desiredQuantity) return;

    const qty = parseInt(desiredQuantity.replace(/,/g, ''), 10);
    if (qty <= 0 || isNaN(qty)) return;

    const { ownerType, ownerId } = parseOwnerKey(effectiveOwner);
    const ownerInfo = owners.find(o => o.ownerType === ownerType && o.ownerId === ownerId);

    setSaving(true);
    setError(null);

    try {
      const marker: StockpileMarker = {
        userId: 0,
        typeId: selectedItem.TypeID,
        ownerType,
        ownerId,
        locationId,
        containerId,
        divisionNumber,
        desiredQuantity: qty,
      };

      const res = await fetch('/api/stockpiles/upsert', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(marker),
      });

      if (!res.ok) {
        throw new Error('Failed to save stockpile marker');
      }

      const delta = -qty;
      const phantomAsset: Asset = {
        name: selectedItem.TypeName,
        typeId: selectedItem.TypeID,
        quantity: 0,
        volume: 0,
        ownerType,
        ownerName: ownerInfo?.ownerName || '',
        ownerId,
        desiredQuantity: qty,
        stockpileDelta: delta,
      };

      onSaved(phantomAsset);
      handleClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save');
    } finally {
      setSaving(false);
    }
  };

  const handleClose = () => {
    setSelectedItem(null);
    setItemOptions([]);
    setSelectedOwner('');
    setDesiredQuantity('');
    setError(null);
    onClose();
  };

  const canSave = selectedItem && effectiveOwner && desiredQuantity && parseInt(desiredQuantity.replace(/,/g, ''), 10) > 0;

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>Add Stockpile Marker</DialogTitle>
      <DialogContent>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: 1 }}>
          <Autocomplete
            size="small"
            options={itemOptions}
            getOptionLabel={(option) => option.TypeName}
            value={selectedItem}
            onChange={(_e, value) => setSelectedItem(value)}
            onInputChange={(_e, value, reason) => {
              if (reason === 'input') searchItems(value);
            }}
            loading={itemLoading}
            filterOptions={(x) => x}
            isOptionEqualToValue={(option, value) => option.TypeID === value.TypeID}
            renderOption={(props, option) => (
              <li {...props} key={option.TypeID}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <img
                    src={`https://images.evetech.net/types/${option.TypeID}/icon?size=32`}
                    alt=""
                    width={24}
                    height={24}
                    style={{ borderRadius: 2 }}
                  />
                  {option.TypeName}
                </Box>
              </li>
            )}
            renderInput={(params) => (
              <TextField
                {...params}
                label="Item Type"
                placeholder="Search for an item..."
                InputProps={{
                  ...params.InputProps,
                  endAdornment: (
                    <>
                      {itemLoading ? <CircularProgress color="inherit" size={20} /> : null}
                      {params.InputProps.endAdornment}
                    </>
                  ),
                }}
              />
            )}
          />

          {owners.length > 1 && (
            <FormControl size="small" fullWidth>
              <InputLabel>Owner</InputLabel>
              <Select
                value={selectedOwner}
                label="Owner"
                onChange={(e) => setSelectedOwner(e.target.value)}
              >
                {owners.map(o => (
                  <MenuItem key={`${o.ownerType}:${o.ownerId}`} value={`${o.ownerType}:${o.ownerId}`}>
                    {o.ownerName} ({o.ownerType})
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          )}

          {owners.length === 1 && (
            <Typography variant="body2" color="text.secondary">
              Owner: {owners[0].ownerName} ({owners[0].ownerType})
            </Typography>
          )}

          <TextField
            size="small"
            label="Desired Quantity"
            type="number"
            value={desiredQuantity}
            onChange={(e) => setDesiredQuantity(e.target.value)}
            inputProps={{ min: 1 }}
            fullWidth
          />

          {error && <Alert severity="error">{error}</Alert>}
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose} disabled={saving}>Cancel</Button>
        <Button
          onClick={handleSave}
          variant="contained"
          disabled={saving || !canSave}
          startIcon={saving ? <CircularProgress size={16} /> : undefined}
        >
          {saving ? 'Saving...' : 'Add Stockpile'}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
