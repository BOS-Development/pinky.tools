import Box from '@mui/material/Box';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import CircularProgress from '@mui/material/CircularProgress';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import { PlanResponse } from "@industry-tool/client/data/models";
import { formatISK, formatNumber } from "@industry-tool/utils/formatting";

type Props = {
  planData: PlanResponse | null;
  loading: boolean;
};

export default function ShoppingList({ planData, loading }: Props) {
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

  const totalCost = planData.shopping_list.reduce((sum, item) => sum + item.cost, 0);
  const totalVolume = planData.shopping_list.reduce((sum, item) => sum + item.volume, 0);

  const copyMultibuy = () => {
    const text = planData.shopping_list
      .map(item => `${item.name} ${item.quantity}`)
      .join('\n');
    navigator.clipboard.writeText(text);
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Box sx={{ display: 'flex', gap: 3 }}>
          <Typography variant="body2" color="text.secondary">
            {planData.shopping_list.length} items | Total: {formatISK(totalCost)} | Volume: {formatNumber(totalVolume, 1)} m3
          </Typography>
        </Box>
        <Button
          variant="outlined"
          size="small"
          startIcon={<ContentCopyIcon />}
          onClick={copyMultibuy}
        >
          Copy Multibuy
        </Button>
      </Box>

      <TableContainer>
        <Table size="small" sx={{ '& th': { backgroundColor: '#0f1219', fontWeight: 'bold' } }}>
          <TableHead>
            <TableRow>
              <TableCell>Material</TableCell>
              <TableCell align="right">Quantity</TableCell>
              <TableCell align="right">Unit Price</TableCell>
              <TableCell align="right">Total Cost</TableCell>
              <TableCell align="right">Volume (m3)</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {planData.shopping_list.map((item) => (
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
                <TableCell align="right">{formatISK(item.price)}</TableCell>
                <TableCell align="right">{formatISK(item.cost)}</TableCell>
                <TableCell align="right">{formatNumber(item.volume, 1)}</TableCell>
              </TableRow>
            ))}
            <TableRow sx={{ '& td': { fontWeight: 'bold', borderTop: '2px solid rgba(255,255,255,0.1)' } }}>
              <TableCell>Total</TableCell>
              <TableCell />
              <TableCell />
              <TableCell align="right">{formatISK(totalCost)}</TableCell>
              <TableCell align="right">{formatNumber(totalVolume, 1)}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
}
