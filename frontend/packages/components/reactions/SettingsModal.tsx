import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import { ReactionSettings } from "@industry-tool/pages/reactions";

type Props = {
  open: boolean;
  onClose: () => void;
  settings: ReactionSettings;
  onSettingChange: <K extends keyof ReactionSettings>(key: K, value: ReactionSettings[K]) => void;
};

export default function SettingsModal({ open, onClose, settings, onSettingChange }: Props) {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Advanced Settings</DialogTitle>
      <DialogContent>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: 1 }}>
          <Typography variant="subtitle2" color="text.secondary">Skills & Fees</Typography>

          <TextField
            size="small"
            label="Reactions Skill Level"
            type="number"
            value={settings.reactions_skill}
            onChange={(e) => {
              const val = parseInt(e.target.value);
              if (val >= 0 && val <= 5) onSettingChange('reactions_skill', val);
            }}
            inputProps={{ min: 0, max: 5 }}
            fullWidth
          />

          <TextField
            size="small"
            label="Facility Tax (%)"
            type="number"
            value={settings.facility_tax}
            onChange={(e) => onSettingChange('facility_tax', parseFloat(e.target.value) || 0)}
            inputProps={{ min: 0, step: 0.25 }}
            fullWidth
          />

          <TextField
            size="small"
            label="Broker Fee (%)"
            type="number"
            value={settings.broker_fee}
            onChange={(e) => onSettingChange('broker_fee', parseFloat(e.target.value) || 0)}
            inputProps={{ min: 0, step: 0.1 }}
            fullWidth
          />

          <TextField
            size="small"
            label="Sales Tax (%)"
            type="number"
            value={settings.sales_tax}
            onChange={(e) => onSettingChange('sales_tax', parseFloat(e.target.value) || 0)}
            inputProps={{ min: 0, step: 0.1 }}
            fullWidth
          />

          <Typography variant="subtitle2" color="text.secondary" sx={{ mt: 1 }}>Pricing</Typography>

          <FormControl size="small" fullWidth>
            <InputLabel>Input Price</InputLabel>
            <Select
              value={settings.input_price}
              label="Input Price"
              onChange={(e) => onSettingChange('input_price', e.target.value)}
            >
              <MenuItem value="sell">Jita Sell</MenuItem>
              <MenuItem value="buy">Jita Buy</MenuItem>
              <MenuItem value="split">Split (avg)</MenuItem>
            </Select>
          </FormControl>

          <FormControl size="small" fullWidth>
            <InputLabel>Output Price</InputLabel>
            <Select
              value={settings.output_price}
              label="Output Price"
              onChange={(e) => onSettingChange('output_price', e.target.value)}
            >
              <MenuItem value="sell">Jita Sell</MenuItem>
              <MenuItem value="buy">Jita Buy</MenuItem>
            </Select>
          </FormControl>

          <Typography variant="subtitle2" color="text.secondary" sx={{ mt: 1 }}>Shipping</Typography>

          <FormControlLabel
            control={
              <Switch
                checked={settings.ship_inputs}
                onChange={(e) => onSettingChange('ship_inputs', e.target.checked)}
              />
            }
            label="Ship Inputs"
          />

          <FormControlLabel
            control={
              <Switch
                checked={settings.ship_outputs}
                onChange={(e) => onSettingChange('ship_outputs', e.target.checked)}
              />
            }
            label="Ship Outputs"
          />

          <TextField
            size="small"
            label="Shipping ISK/m3"
            type="number"
            value={settings.shipping_m3}
            onChange={(e) => onSettingChange('shipping_m3', parseFloat(e.target.value) || 0)}
            inputProps={{ min: 0, step: 100 }}
            fullWidth
          />

          <TextField
            size="small"
            label="Collateral (%)"
            type="number"
            value={settings.shipping_collateral * 100}
            onChange={(e) => onSettingChange('shipping_collateral', (parseFloat(e.target.value) || 0) / 100)}
            inputProps={{ min: 0, max: 100, step: 0.5 }}
            fullWidth
          />
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Close</Button>
      </DialogActions>
    </Dialog>
  );
}
