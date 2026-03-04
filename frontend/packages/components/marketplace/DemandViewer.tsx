import { useState, useEffect, useRef } from 'react';
import { useSession } from "next-auth/react";
import { TrendingUp, Search } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { toast } from '@/components/ui/sonner';
import Loading from "@industry-tool/components/loading";

export type BuyOrder = {
  id: number;
  buyerUserId: number;
  typeId: number;
  typeName: string;
  locationId: number;
  locationName: string;
  quantityDesired: number;
  minPricePerUnit: number;
  notes?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

export default function DemandViewer() {
  const { data: session } = useSession();
  const [demand, setDemand] = useState<BuyOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const hasFetchedRef = useRef(false);

  useEffect(() => {
    if (session && !hasFetchedRef.current) {
      hasFetchedRef.current = true;
      fetchDemand();
    }
  }, [session]);

  const fetchDemand = async () => {
    try {
      const response = await fetch('/api/buy-orders/demand');
      if (!response.ok) throw new Error('Failed to fetch demand');
      const data = await response.json();
      setDemand(data);
    } catch (error) {
      console.error('Error fetching demand:', error);
      toast.error('Failed to load demand data');
    } finally {
      setLoading(false);
    }
  };

  const formatNumber = (num: number) => num.toLocaleString();
  const formatISK = (isk: number) => `${isk.toLocaleString()} ISK`;
  const formatDate = (dateString: string) => new Date(dateString).toLocaleDateString();

  const filteredDemand = demand.filter((order) =>
    order.typeName.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Group orders by item type and aggregate
  const aggregatedDemand = filteredDemand.reduce((acc, order) => {
    const key = order.typeId;
    if (!acc[key]) {
      acc[key] = {
        typeId: order.typeId,
        typeName: order.typeName,
        totalQuantity: 0,
        maxPrice: 0,
        orderCount: 0,
        orders: [] as BuyOrder[],
      };
    }
    acc[key].totalQuantity += order.quantityDesired;
    acc[key].maxPrice = Math.max(acc[key].maxPrice, order.minPricePerUnit);
    acc[key].orderCount += 1;
    acc[key].orders.push(order);
    return acc;
  }, {} as Record<number, {
    typeId: number;
    typeName: string;
    totalQuantity: number;
    maxPrice: number;
    orderCount: number;
    orders: BuyOrder[];
  }>);

  const aggregatedData = Object.values(aggregatedDemand).sort(
    (a, b) => b.totalQuantity - a.totalQuantity
  );

  if (loading) {
    return <Loading />;
  }

  return (
    <div className="max-w-[1280px] my-4">
      <Card className="bg-[#12151f] border-[rgba(148,163,184,0.1)]">
        <CardContent className="p-6">
          <div className="flex justify-between items-center mb-6">
            <div className="flex items-center gap-3">
              <TrendingUp className="h-7 w-7 text-[#00d4ff]" />
              <div>
                <h2 className="text-xl font-semibold text-[#e2e8f0]">Market Demand</h2>
                <p className="text-sm text-[#94a3b8]">Buy orders from your contacts</p>
              </div>
            </div>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-[#64748b]" />
              <Input
                placeholder="Search items..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-9 w-52"
              />
            </div>
          </div>

          {demand.length === 0 ? (
            <p className="text-[#94a3b8] text-center py-8">
              No active buy orders from your contacts yet.
              <br />
              When your contacts create buy orders, they&apos;ll appear here!
            </p>
          ) : (
            <>
              {/* Aggregated Summary */}
              <h3 className="text-lg font-semibold text-[#e2e8f0] mb-3 mt-2">Aggregated Demand</h3>
              <div className="overflow-x-auto rounded-sm border border-[rgba(148,163,184,0.1)] mb-8">
                <Table>
                  <TableHeader>
                    <TableRow className="bg-[#0f1219]">
                      <TableHead>Item</TableHead>
                      <TableHead className="text-right">Total Quantity Wanted</TableHead>
                      <TableHead className="text-right">Highest Floor Price</TableHead>
                      <TableHead className="text-right">Potential Revenue</TableHead>
                      <TableHead className="text-center">Number of Orders</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {aggregatedData.map((item) => (
                      <TableRow key={item.typeId} className="bg-[#12151f] hover:bg-[rgba(0,212,255,0.04)]">
                        <TableCell>
                          <strong className="text-[#e2e8f0]">{item.typeName}</strong>
                        </TableCell>
                        <TableCell className="text-right text-[#e2e8f0]">{formatNumber(item.totalQuantity)}</TableCell>
                        <TableCell className="text-right text-[#e2e8f0]">{formatISK(item.maxPrice)}</TableCell>
                        <TableCell className="text-right">
                          <span className="font-bold text-[#10b981]">
                            {formatISK(item.totalQuantity * item.maxPrice)}
                          </span>
                        </TableCell>
                        <TableCell className="text-center">
                          <Badge className="bg-[rgba(59,130,246,0.15)] text-[#60a5fa] border border-[rgba(59,130,246,0.3)] hover:bg-[rgba(59,130,246,0.2)] cursor-default">
                            {item.orderCount}
                          </Badge>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>

              {/* Detailed Orders */}
              <h3 className="text-lg font-semibold text-[#e2e8f0] mb-3">Individual Buy Orders</h3>
              <div className="overflow-x-auto rounded-sm border border-[rgba(148,163,184,0.1)]">
                <Table>
                  <TableHeader>
                    <TableRow className="bg-[#0f1219]">
                      <TableHead>Item</TableHead>
                      <TableHead>Location</TableHead>
                      <TableHead className="text-right">Quantity</TableHead>
                      <TableHead className="text-right">Min Price/Unit</TableHead>
                      <TableHead className="text-right">Est. Revenue</TableHead>
                      <TableHead>Notes</TableHead>
                      <TableHead>Created</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredDemand.map((order) => (
                      <TableRow key={order.id} className="bg-[#12151f] hover:bg-[rgba(0,212,255,0.04)]">
                        <TableCell className="text-[#e2e8f0]">{order.typeName}</TableCell>
                        <TableCell className="text-[#94a3b8]">{order.locationName || '-'}</TableCell>
                        <TableCell className="text-right text-[#e2e8f0]">{formatNumber(order.quantityDesired)}</TableCell>
                        <TableCell className="text-right text-[#e2e8f0]">{formatISK(order.minPricePerUnit)}</TableCell>
                        <TableCell className="text-right text-[#e2e8f0]">
                          {formatISK(order.quantityDesired * order.minPricePerUnit)}
                        </TableCell>
                        <TableCell className="text-[#94a3b8]">{order.notes || '-'}</TableCell>
                        <TableCell className="text-[#94a3b8]">{formatDate(order.createdAt)}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
