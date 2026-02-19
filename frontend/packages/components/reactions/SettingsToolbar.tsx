import { useState } from 'react';
import Box from '@mui/material/Box';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import IconButton from '@mui/material/IconButton';
import Autocomplete from '@mui/material/Autocomplete';
import TuneIcon from '@mui/icons-material/Tune';
import { ReactionSettings } from "@industry-tool/pages/reactions";
import { ReactionSystem } from "@industry-tool/client/data/models";
import SettingsModal from './SettingsModal';

type Props = {
  settings: ReactionSettings;
  systems: ReactionSystem[];
  onSettingChange: <K extends keyof ReactionSettings>(key: K, value: ReactionSettings[K]) => void;
};

export default function SettingsToolbar({ settings, systems, onSettingChange }: Props) {
  const [modalOpen, setModalOpen] = useState(false);

  const selectedSystem = systems.find(s => s.system_id === settings.system_id) || null;

  return (
    <>
      <Box sx={{
        display: 'flex',
        flexWrap: 'wrap',
        gap: 1.5,
        alignItems: 'center',
        py: 2,
        px: 1,
      }}>
        <Autocomplete
          size="small"
          options={systems}
          getOptionLabel={(option) => option.name}
          value={selectedSystem}
          onChange={(_, newValue) => onSettingChange('system_id', newValue?.system_id || 0)}
          renderInput={(params) => (
            <TextField {...params} label="System" placeholder="Search systems..." />
          )}
          sx={{ minWidth: 220 }}
          isOptionEqualToValue={(option, value) => option.system_id === value.system_id}
        />

        <FormControl size="small" sx={{ minWidth: 120 }}>
          <InputLabel>Structure</InputLabel>
          <Select
            value={settings.structure}
            label="Structure"
            onChange={(e) => onSettingChange('structure', e.target.value)}
          >
            <MenuItem value="tatara">Tatara</MenuItem>
            <MenuItem value="athanor">Athanor</MenuItem>
          </Select>
        </FormControl>

        <FormControl size="small" sx={{ minWidth: 90 }}>
          <InputLabel>Rig</InputLabel>
          <Select
            value={settings.rig}
            label="Rig"
            onChange={(e) => onSettingChange('rig', e.target.value)}
          >
            <MenuItem value="t2">T2</MenuItem>
            <MenuItem value="t1">T1</MenuItem>
            <MenuItem value="none">None</MenuItem>
          </Select>
        </FormControl>

        <FormControl size="small" sx={{ minWidth: 110 }}>
          <InputLabel>Security</InputLabel>
          <Select
            value={settings.security}
            label="Security"
            onChange={(e) => onSettingChange('security', e.target.value)}
          >
            <MenuItem value="null">Null / WH</MenuItem>
            <MenuItem value="low">Lowsec</MenuItem>
            <MenuItem value="high">Highsec</MenuItem>
          </Select>
        </FormControl>

        <TextField
          size="small"
          label="Cycle Days"
          type="number"
          value={settings.cycle_days}
          onChange={(e) => {
            const val = parseInt(e.target.value);
            if (val >= 1 && val <= 30) onSettingChange('cycle_days', val);
          }}
          sx={{ width: 100 }}
          inputProps={{ min: 1, max: 30 }}
        />

        <IconButton
          onClick={() => setModalOpen(true)}
          sx={{ color: 'primary.main' }}
          title="Advanced Settings"
        >
          <TuneIcon />
        </IconButton>
      </Box>

      <SettingsModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        settings={settings}
        onSettingChange={onSettingChange}
      />
    </>
  );
}
