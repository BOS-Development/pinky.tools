import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
import CircularProgress from '@mui/material/CircularProgress';
import ViewModuleIcon from '@mui/icons-material/ViewModule';
import AccountBalanceWalletIcon from '@mui/icons-material/AccountBalanceWallet';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import ShowChartIcon from '@mui/icons-material/ShowChart';
import { PlanResponse, ReactionsResponse } from "@industry-tool/client/data/models";
import { formatISK, formatNumber, getValueColor } from "@industry-tool/utils/formatting";

type Props = {
  planData: PlanResponse | null;
  reactionsData: ReactionsResponse | null;
  selections: Record<number, number>;
  loading: boolean;
};

export default function PlanSummary({ planData, reactionsData, selections, loading }: Props) {
  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (!planData || !planData.summary) {
    return (
      <Typography color="text.secondary" sx={{ py: 4, textAlign: 'center' }}>
        Select reactions in the Pick Reactions tab to see a plan summary.
      </Typography>
    );
  }

  const { summary } = planData;

  const statCards = [
    {
      label: 'Total Slots',
      value: `${summary.total_slots}`,
      subtitle: `${summary.intermediate_slots} intermediate + ${summary.complex_slots} complex`,
      icon: <ViewModuleIcon />,
      color: '#3b82f6',
    },
    {
      label: 'Investment',
      value: formatISK(summary.investment),
      subtitle: 'Total input + job costs',
      icon: <AccountBalanceWalletIcon />,
      color: '#ef4444',
    },
    {
      label: 'Revenue',
      value: formatISK(summary.revenue),
      subtitle: 'Total output value',
      icon: <TrendingUpIcon />,
      color: '#10b981',
    },
    {
      label: 'Profit',
      value: formatISK(summary.profit),
      subtitle: `${summary.margin.toFixed(2)}% margin`,
      icon: <ShowChartIcon />,
      color: getValueColor(summary.profit),
    },
  ];

  // Build selected complex reactions list
  const selectedReactions = reactionsData?.reactions.filter(
    r => selections[r.reaction_type_id] > 0
  ) || [];

  return (
    <Box>
      <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 2, mb: 3 }}>
        {statCards.map((card) => (
          <Card
            key={card.label}
            sx={{
              background: `linear-gradient(135deg, ${card.color}15 0%, ${card.color}05 100%)`,
              borderLeft: `3px solid ${card.color}`,
            }}
          >
            <CardContent sx={{ py: 1.5, '&:last-child': { pb: 1.5 } }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
                <Box sx={{ color: card.color }}>{card.icon}</Box>
                <Typography variant="caption" color="text.secondary">{card.label}</Typography>
              </Box>
              <Typography variant="h6" sx={{ color: card.color, fontWeight: 'bold' }}>
                {card.value}
              </Typography>
              <Typography variant="caption" color="text.secondary">{card.subtitle}</Typography>
            </CardContent>
          </Card>
        ))}
      </Box>

      {planData.intermediates.length > 0 && (
        <>
          <Typography variant="subtitle1" sx={{ mb: 1, fontWeight: 'bold' }}>Intermediate Reactions</Typography>
          <TableContainer sx={{ mb: 3 }}>
            <Table size="small" sx={{ '& th': { backgroundColor: '#0f1219', fontWeight: 'bold' } }}>
              <TableHead>
                <TableRow>
                  <TableCell>Intermediate</TableCell>
                  <TableCell align="right">Slots</TableCell>
                  <TableCell align="right">Runs</TableCell>
                  <TableCell align="right">Produced</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {planData.intermediates.map((item) => (
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
                    <TableCell align="right">{item.slots}</TableCell>
                    <TableCell align="right">{formatNumber(item.runs)}</TableCell>
                    <TableCell align="right">{formatNumber(item.produced)}</TableCell>
                  </TableRow>
                ))}
                <TableRow sx={{ '& td': { fontWeight: 'bold', borderTop: '2px solid rgba(255,255,255,0.1)' } }}>
                  <TableCell>Total</TableCell>
                  <TableCell align="right">{summary.intermediate_slots}</TableCell>
                  <TableCell />
                  <TableCell />
                </TableRow>
              </TableBody>
            </Table>
          </TableContainer>
        </>
      )}

      {selectedReactions.length > 0 && (
        <>
          <Typography variant="subtitle1" sx={{ mb: 1, fontWeight: 'bold' }}>Complex Reactions</Typography>
          <TableContainer>
            <Table size="small" sx={{ '& th': { backgroundColor: '#0f1219', fontWeight: 'bold' } }}>
              <TableHead>
                <TableRow>
                  <TableCell>Reaction</TableCell>
                  <TableCell align="right">Instances</TableCell>
                  <TableCell align="right">Lines</TableCell>
                  <TableCell align="right">Runs</TableCell>
                  <TableCell align="right">Produced</TableCell>
                  <TableCell align="right">Net Profit</TableCell>
                  <TableCell align="right">Margin</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {(() => {
                  let totalNetProfit = 0;
                  let totalRevenue = 0;
                  return (
                    <>
                      {selectedReactions.map((r) => {
                        const instances = selections[r.reaction_type_id] || 0;
                        const lines = instances * r.complex_instances;
                        const cycleProfitTotal = r.profit_per_cycle * lines;
                        const cycleRevenue = r.output_value_per_run * r.runs_per_cycle * lines;
                        totalNetProfit += cycleProfitTotal;
                        totalRevenue += cycleRevenue;
                        return (
                          <TableRow
                            key={r.reaction_type_id}
                            sx={{ '&:nth-of-type(odd)': { backgroundColor: 'rgba(255,255,255,0.02)' } }}
                          >
                            <TableCell>
                              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                <img
                                  src={`https://images.evetech.net/types/${r.product_type_id}/icon?size=32`}
                                  alt=""
                                  width={24}
                                  height={24}
                                  style={{ borderRadius: 2 }}
                                />
                                {r.product_name}
                              </Box>
                            </TableCell>
                            <TableCell align="right">{instances}</TableCell>
                            <TableCell align="right">{lines}</TableCell>
                            <TableCell align="right">{formatNumber(r.runs_per_cycle)}</TableCell>
                            <TableCell align="right">{formatNumber(r.product_qty_per_run * r.runs_per_cycle * lines)}</TableCell>
                            <TableCell align="right" sx={{ color: getValueColor(cycleProfitTotal) }}>
                              {formatISK(cycleProfitTotal)}
                            </TableCell>
                            <TableCell align="right" sx={{ color: getValueColor(r.margin) }}>
                              {r.margin.toFixed(2)}%
                            </TableCell>
                          </TableRow>
                        );
                      })}
                      {(() => {
                        const totalMargin = totalRevenue > 0 ? (totalNetProfit / totalRevenue) * 100 : 0;
                        return (
                          <TableRow sx={{ '& td': { fontWeight: 'bold', borderTop: '2px solid rgba(255,255,255,0.1)' } }}>
                            <TableCell>Total</TableCell>
                            <TableCell />
                            <TableCell align="right">{summary.complex_slots}</TableCell>
                            <TableCell />
                            <TableCell />
                            <TableCell align="right" sx={{ color: getValueColor(totalNetProfit) }}>
                              {formatISK(totalNetProfit)}
                            </TableCell>
                            <TableCell align="right" sx={{ color: getValueColor(totalMargin) }}>
                              {totalMargin.toFixed(2)}%
                            </TableCell>
                          </TableRow>
                        );
                      })()}
                    </>
                  );
                })()}
              </TableBody>
            </Table>
          </TableContainer>
        </>
      )}
    </Box>
  );
}
