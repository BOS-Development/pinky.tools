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
      <Card className="bg-background-panel border-overlay-subtle">
        <CardContent className="p-6">
          <div className="flex justify-between items-center mb-6">
            <div className="flex items-center gap-3">
              <TrendingUp className="h-7 w-7 text-primary" />
              <div>
                <h2 className="text-xl font-semibold text-text-emphasis">Market Demand</h2>
                <p className="text-sm text-text-secondary">Buy orders from your contacts</p>
              </div>
            </div>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-text-muted" />
              <Input
                placeholder="Search items..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-9 w-52"
              />
            </div>
          </div>

          {demand.length === 0 ? (
            <p className="text-text-secondary text-center py-8">
              No active buy orders from your contacts yet.
              <br />
              When your contacts create buy orders, they&apos;ll appear here!
            </p>
          ) : (
            <>
              {/* Aggregated Summary */}
              <h3 className="text-lg font-semibold text-text-emphasis mb-3 mt-2">Aggregated Demand</h3>
              <div className="overflow-x-auto rounded-sm border border-overlay-subtle mb-8">
                <Table>
                  <TableHeader>
                    <TableRow className="bg-background-void">
                      <TableHead>Item</TableHead>
                      <TableHead className="text-right">Total Quantity Wanted</TableHead>
                      <TableHead className="text-right">Highest Floor Price</TableHead>
                      <TableHead className="text-right">Potential Revenue</TableHead>
                      <TableHead className="text-center">Number of Orders</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {aggregatedData.map((item) => (
                      <TableRow key={item.typeId} className="bg-background-panel hover:bg-interactive-hover">
                        <TableCell>
                          <strong className="text-text-emphasis">{item.typeName}</strong>
                        </TableCell>
                        <TableCell className="text-right text-text-emphasis">{formatNumber(item.totalQuantity)}</TableCell>
                        <TableCell className="text-right text-text-emphasis">{formatISK(item.maxPrice)}</TableCell>
                        <TableCell className="text-right">
                          <span className="font-bold text-teal-success">
                            {formatISK(item.totalQuantity * item.maxPrice)}
                          </span>
                        </TableCell>
                        <TableCell className="text-center">
                          <Badge className="bg-accent-blue-muted text-blue-science border border-accent-blue/30 hover:bg-accent-blue/20 cursor-default">
                            {item.orderCount}
                          </Badge>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>

              {/* Detailed Orders */}
              <h3 className="text-lg font-semibold text-text-emphasis mb-3">Individual Buy Orders</h3>
              <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
                <Table>
                  <TableHeader>
                    <TableRow className="bg-background-void">
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
                      <TableRow key={order.id} className="bg-background-panel hover:bg-interactive-hover">
                        <TableCell className="text-text-emphasis">{order.typeName}</TableCell>
                        <TableCell className="text-text-secondary">{order.locationName || '-'}</TableCell>
                        <TableCell className="text-right text-text-emphasis">{formatNumber(order.quantityDesired)}</TableCell>
                        <TableCell className="text-right text-text-emphasis">{formatISK(order.minPricePerUnit)}</TableCell>
                        <TableCell className="text-right text-text-emphasis">
                          {formatISK(order.quantityDesired * order.minPricePerUnit)}
                        </TableCell>
                        <TableCell className="text-text-secondary">{order.notes || '-'}</TableCell>
                        <TableCell className="text-text-secondary">{formatDate(order.createdAt)}</TableCell>
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
