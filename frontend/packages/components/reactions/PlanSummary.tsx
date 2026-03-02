import { Loader2, LayoutGrid, Wallet, TrendingUp, BarChart3 } from 'lucide-react';
import { Card, CardContent } from "@/components/ui/card";
import {
  Table, TableHeader, TableBody, TableRow, TableHead, TableCell,
} from "@/components/ui/table";
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
      <div className="flex justify-center py-16">
        <Loader2 className="h-8 w-8 animate-spin text-[var(--color-primary-cyan)]" />
      </div>
    );
  }

  if (!planData || !planData.summary) {
    return (
      <p className="py-8 text-center text-[var(--color-text-secondary)]">
        Select reactions in the Pick Reactions tab to see a plan summary.
      </p>
    );
  }

  const { summary } = planData;

  const statCards = [
    {
      label: 'Total Slots',
      value: `${summary.total_slots}`,
      subtitle: `${summary.intermediate_slots} intermediate + ${summary.complex_slots} complex`,
      icon: <LayoutGrid className="h-5 w-5" />,
      color: '#00d4ff',
    },
    {
      label: 'Investment',
      value: formatISK(summary.investment),
      subtitle: 'Total input + job costs',
      icon: <Wallet className="h-5 w-5" />,
      color: '#ef4444',
    },
    {
      label: 'Revenue',
      value: formatISK(summary.revenue),
      subtitle: 'Total output value',
      icon: <TrendingUp className="h-5 w-5" />,
      color: '#10b981',
    },
    {
      label: 'Profit',
      value: formatISK(summary.profit),
      subtitle: `${summary.margin.toFixed(2)}% margin`,
      icon: <BarChart3 className="h-5 w-5" />,
      color: getValueColor(summary.profit),
    },
  ];

  // Build selected complex reactions list
  const selectedReactions = reactionsData?.reactions.filter(
    r => selections[r.reaction_type_id] > 0
  ) || [];

  return (
    <div>
      <div className="grid grid-cols-[repeat(auto-fit,minmax(220px,1fr))] gap-3 mb-6">
        {statCards.map((card) => (
          <Card
            key={card.label}
            className="bg-[var(--color-bg-panel)]"
            style={{ borderLeft: `3px solid ${card.color}` }}
          >
            <CardContent className="py-3 px-4">
              <div className="flex items-center gap-2 mb-1">
                <span style={{ color: card.color }}>{card.icon}</span>
                <span className="text-xs text-[var(--color-text-secondary)]">{card.label}</span>
              </div>
              <p className="text-lg font-bold" style={{ color: card.color }}>
                {card.value}
              </p>
              <span className="text-xs text-[var(--color-text-secondary)]">{card.subtitle}</span>
            </CardContent>
          </Card>
        ))}
      </div>

      {planData.intermediates.length > 0 && (
        <>
          <h3 className="text-sm font-bold mb-2 text-[var(--color-text-primary)]">Intermediate Reactions</h3>
          <div className="mb-6">
            <Table>
              <TableHeader>
                <TableRow className="bg-[var(--color-bg-panel)]">
                  <TableHead>Intermediate</TableHead>
                  <TableHead className="text-right">Slots</TableHead>
                  <TableHead className="text-right">Runs</TableHead>
                  <TableHead className="text-right">Produced</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {planData.intermediates.map((item) => (
                  <TableRow
                    key={item.type_id}
                    className="odd:bg-white/[0.02]"
                  >
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <img
                          src={`https://images.evetech.net/types/${item.type_id}/icon?size=32`}
                          alt=""
                          width={24}
                          height={24}
                          className="rounded-sm"
                        />
                        {item.name}
                      </div>
                    </TableCell>
                    <TableCell className="text-right">{item.slots}</TableCell>
                    <TableCell className="text-right">{formatNumber(item.runs)}</TableCell>
                    <TableCell className="text-right">{formatNumber(item.produced)}</TableCell>
                  </TableRow>
                ))}
                <TableRow className="[&_td]:font-bold [&_td]:border-t-2 [&_td]:border-white/10">
                  <TableCell>Total</TableCell>
                  <TableCell className="text-right">{summary.intermediate_slots}</TableCell>
                  <TableCell />
                  <TableCell />
                </TableRow>
              </TableBody>
            </Table>
          </div>
        </>
      )}

      {selectedReactions.length > 0 && (
        <>
          <h3 className="text-sm font-bold mb-2 text-[var(--color-text-primary)]">Complex Reactions</h3>
          <Table>
            <TableHeader>
              <TableRow className="bg-[var(--color-bg-panel)]">
                <TableHead>Reaction</TableHead>
                <TableHead className="text-right">Instances</TableHead>
                <TableHead className="text-right">Lines</TableHead>
                <TableHead className="text-right">Runs</TableHead>
                <TableHead className="text-right">Produced</TableHead>
                <TableHead className="text-right">Net Profit</TableHead>
                <TableHead className="text-right">Margin</TableHead>
              </TableRow>
            </TableHeader>
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
                          className="odd:bg-white/[0.02]"
                        >
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <img
                                src={`https://images.evetech.net/types/${r.product_type_id}/icon?size=32`}
                                alt=""
                                width={24}
                                height={24}
                                className="rounded-sm"
                              />
                              {r.product_name}
                            </div>
                          </TableCell>
                          <TableCell className="text-right">{instances}</TableCell>
                          <TableCell className="text-right">{lines}</TableCell>
                          <TableCell className="text-right">{formatNumber(r.runs_per_cycle)}</TableCell>
                          <TableCell className="text-right">{formatNumber(r.product_qty_per_run * r.runs_per_cycle * lines)}</TableCell>
                          <TableCell className="text-right" style={{ color: getValueColor(cycleProfitTotal) }}>
                            {formatISK(cycleProfitTotal)}
                          </TableCell>
                          <TableCell className="text-right" style={{ color: getValueColor(r.margin) }}>
                            {r.margin.toFixed(2)}%
                          </TableCell>
                        </TableRow>
                      );
                    })}
                    {(() => {
                      const totalMargin = totalRevenue > 0 ? (totalNetProfit / totalRevenue) * 100 : 0;
                      return (
                        <TableRow className="[&_td]:font-bold [&_td]:border-t-2 [&_td]:border-white/10">
                          <TableCell>Total</TableCell>
                          <TableCell />
                          <TableCell className="text-right">{summary.complex_slots}</TableCell>
                          <TableCell />
                          <TableCell />
                          <TableCell className="text-right" style={{ color: getValueColor(totalNetProfit) }}>
                            {formatISK(totalNetProfit)}
                          </TableCell>
                          <TableCell className="text-right" style={{ color: getValueColor(totalMargin) }}>
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
        </>
      )}
    </div>
  );
}
