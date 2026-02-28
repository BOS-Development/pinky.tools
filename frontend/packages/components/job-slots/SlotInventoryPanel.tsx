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
import CircularProgress from '@mui/material/CircularProgress';
import Chip from '@mui/material/Chip';

type ActivitySlotInfo = {
  activityType: string;
  slotsMax: number;
  slotsInUse: number;
  slotsReserved: number;
  slotsAvailable: number;
  slotsListed: number;
};

type CharacterSlotInventory = {
  characterId: number;
  characterName: string;
  slotsByActivity: Record<string, ActivitySlotInfo>;
};

const ACTIVITY_LABELS: Record<string, string> = {
  manufacturing: 'Manufacturing',
  reaction: 'Reactions',
  copying: 'Copying',
  invention: 'Invention',
  me_research: 'ME Research',
  te_research: 'TE Research',
};

export default function SlotInventoryPanel() {
  const { data: session } = useSession();
  const [inventory, setInventory] = useState<CharacterSlotInventory[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (session) {
      fetchInventory();
    }
  }, [session]);

  const fetchInventory = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/job-slots/inventory');
      if (response.ok) {
        const data = await response.json();
        setInventory(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch slot inventory:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (inventory.length === 0) {
    return (
      <Paper sx={{ p: 4, textAlign: 'center' }}>
        <Typography variant="h6" color="text.secondary">
          No slot data available
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
          Make sure you have characters with industry skills synced.
        </Typography>
      </Paper>
    );
  }

  return (
    <TableContainer component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Character</TableCell>
            <TableCell>Activity Type</TableCell>
            <TableCell align="right">Max Slots</TableCell>
            <TableCell align="right">In Use</TableCell>
            <TableCell align="right">Reserved</TableCell>
            <TableCell align="right">Listed</TableCell>
            <TableCell align="right">Available</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {inventory.map((char) => (
            Object.entries(char.slotsByActivity).map(([activityType, slotInfo]) => (
              <TableRow key={`${char.characterId}-${activityType}`} hover>
                <TableCell>
                  <Typography variant="body2" sx={{ fontWeight: 500 }}>
                    {char.characterName}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Chip
                    label={ACTIVITY_LABELS[activityType] || activityType}
                    size="small"
                    sx={{
                      background: 'rgba(59, 130, 246, 0.1)',
                      borderColor: 'rgba(59, 130, 246, 0.3)',
                      color: '#60a5fa',
                    }}
                  />
                </TableCell>
                <TableCell align="right">{slotInfo.slotsMax}</TableCell>
                <TableCell align="right">{slotInfo.slotsInUse}</TableCell>
                <TableCell align="right">{slotInfo.slotsReserved}</TableCell>
                <TableCell align="right">{slotInfo.slotsListed}</TableCell>
                <TableCell align="right">
                  <Typography
                    variant="body2"
                    sx={{
                      fontWeight: 600,
                      color: slotInfo.slotsAvailable > 0 ? '#10b981' : '#ef4444',
                    }}
                  >
                    {slotInfo.slotsAvailable}
                  </Typography>
                </TableCell>
              </TableRow>
            ))
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}
