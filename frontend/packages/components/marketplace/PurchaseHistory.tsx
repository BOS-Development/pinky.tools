import { useState, useEffect } from 'react';
import { useSession } from 'next-auth/react';
import { CheckCircle, XCircle, ClipboardList, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { toast } from '@/components/ui/sonner';
import { cn } from '@/lib/utils';

type PurchaseTransaction = {
  id: number;
  forSaleItemId: number;
  buyerUserId: number;
  sellerUserId: number;
  typeId: number;
  typeName: string;
  quantityPurchased: number;
  pricePerUnit: number;
  totalPrice: number;
  status: string;
  transactionNotes?: string;
  buyOrderId?: number;
  isAutoFulfilled: boolean;
  purchasedAt: string;
};

export default function PurchaseHistory() {
  const { data: session } = useSession();
  const [activeTab, setActiveTab] = useState('purchases');
  const [buyerHistory, setBuyerHistory] = useState<PurchaseTransaction[]>([]);
  const [sellerHistory, setSellerHistory] = useState<PurchaseTransaction[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (session) {
      fetchHistory();
    }
  }, [session]);

  const fetchHistory = async () => {
    setLoading(true);
    try {
      const [buyerResponse, sellerResponse] = await Promise.all([
        fetch('/api/purchases/buyer'),
        fetch('/api/purchases/seller'),
      ]);

      if (buyerResponse.ok) {
        const buyerData = await buyerResponse.json();
        setBuyerHistory(buyerData || []);
      }

      if (sellerResponse.ok) {
        const sellerData = await sellerResponse.json();
        setSellerHistory(sellerData || []);
      }
    } catch (error) {
      console.error('Failed to fetch purchase history:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getStatusBadgeClass = (status: string): string => {
    switch (status) {
      case 'pending':
        return 'bg-[rgba(245,158,11,0.15)] text-amber-manufacturing border border-[rgba(245,158,11,0.3)]';
      case 'contract_created':
        return 'bg-[rgba(59,130,246,0.15)] text-blue-science border border-[rgba(59,130,246,0.3)]';
      case 'completed':
        return 'bg-[rgba(16,185,129,0.15)] text-teal-success border border-[rgba(16,185,129,0.3)]';
      case 'cancelled':
        return 'bg-[rgba(239,68,68,0.15)] text-rose-danger border border-[rgba(239,68,68,0.3)]';
      default:
        return 'bg-overlay-subtle text-text-muted border border-overlay-strong';
    }
  };

  const handleMarkContractCreated = async (purchaseId: number) => {
    try {
      const response = await fetch(`/api/purchases/${purchaseId}/mark-contract-created`, {
        method: 'POST',
      });

      if (response.ok) {
        await fetchHistory();
        toast.success('Contract marked as created');
      } else {
        const error = await response.json();
        toast.error(error.error || 'Failed to mark contract created');
      }
    } catch (error) {
      console.error('Failed to mark contract created:', error);
      toast.error('Failed to mark contract created');
    }
  };

  const handleCompletePurchase = async (purchaseId: number) => {
    try {
      const response = await fetch(`/api/purchases/${purchaseId}/complete`, {
        method: 'POST',
      });

      if (response.ok) {
        await fetchHistory();
        toast.success('Purchase marked as completed');
      } else {
        const error = await response.json();
        toast.error(error.error || 'Failed to complete purchase');
      }
    } catch (error) {
      console.error('Failed to complete purchase:', error);
      toast.error('Failed to complete purchase');
    }
  };

  const handleCancelPurchase = async (purchaseId: number) => {
    if (!confirm('Are you sure you want to cancel this purchase? The quantity will be restored to the listing.')) {
      return;
    }

    try {
      const response = await fetch(`/api/purchases/${purchaseId}/cancel`, {
        method: 'POST',
      });

      if (response.ok) {
        await fetchHistory();
        toast.success('Purchase cancelled successfully');
      } else {
        const error = await response.json();
        toast.error(error.error || 'Failed to cancel purchase');
      }
    } catch (error) {
      console.error('Failed to cancel purchase:', error);
      toast.error('Failed to cancel purchase');
    }
  };

  const renderTransactionsTable = (transactions: PurchaseTransaction[], isBuyer: boolean) => {
    if (transactions.length === 0) {
      return (
        <div className="bg-background-panel rounded-sm border border-overlay-subtle p-8 text-center">
          <h3 className="text-lg font-semibold text-text-secondary">No {isBuyer ? 'purchases' : 'sales'} yet</h3>
          <p className="text-sm text-text-muted mt-1">
            {isBuyer
              ? 'Browse the marketplace to make your first purchase.'
              : 'List items for sale to start selling.'}
          </p>
        </div>
      );
    }

    return (
      <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
        <Table>
          <TableHeader>
            <TableRow className="bg-background-void">
              <TableHead>Date</TableHead>
              <TableHead>Item</TableHead>
              <TableHead className="text-right">Quantity</TableHead>
              <TableHead className="text-right">Price per Unit</TableHead>
              <TableHead className="text-right">Total Price</TableHead>
              <TableHead>Status</TableHead>
              {transactions.some(t => t.transactionNotes) && <TableHead>Notes</TableHead>}
              <TableHead className="text-center">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {transactions.map((transaction) => (
              <TableRow key={transaction.id} className="bg-background-panel hover:bg-interactive-hover">
                <TableCell className="text-text-secondary">{formatDate(transaction.purchasedAt)}</TableCell>
                <TableCell className="text-text-emphasis">
                  <div className="flex items-center gap-1.5">
                    {transaction.typeName}
                    {transaction.isAutoFulfilled && (
                      <Badge className="text-[0.65rem] font-semibold h-5 bg-[rgba(16,185,129,0.15)] text-teal-success border border-[rgba(16,185,129,0.3)] hover:bg-[rgba(16,185,129,0.2)] cursor-default">
                        Auto
                      </Badge>
                    )}
                  </div>
                </TableCell>
                <TableCell className="text-right text-text-emphasis">{transaction.quantityPurchased.toLocaleString()}</TableCell>
                <TableCell className="text-right text-text-emphasis">{transaction.pricePerUnit.toLocaleString()} ISK</TableCell>
                <TableCell className="text-right">
                  <span className={cn("font-semibold", isBuyer ? 'text-rose-danger' : 'text-teal-success')}>
                    {isBuyer ? '-' : '+'}
                    {transaction.totalPrice.toLocaleString()} ISK
                  </span>
                </TableCell>
                <TableCell>
                  <Badge className={cn("text-xs hover:opacity-90 cursor-default", getStatusBadgeClass(transaction.status))}>
                    {transaction.status.replace('_', ' ')}
                  </Badge>
                </TableCell>
                {transactions.some(t => t.transactionNotes) && (
                  <TableCell>
                    {transaction.transactionNotes && (
                      <span className="text-xs text-text-muted">{transaction.transactionNotes}</span>
                    )}
                  </TableCell>
                )}
                <TableCell className="text-center">
                  <div className="flex items-center justify-center gap-1">
                    {/* Buyer actions */}
                    {isBuyer && transaction.status === 'contract_created' && (
                      <Button
                        variant="outline"
                        size="sm"
                        className="border-emerald-500 text-emerald-400 hover:bg-emerald-500/10"
                        onClick={() => handleCompletePurchase(transaction.id)}
                      >
                        <CheckCircle className="h-4 w-4 mr-1" />
                        Complete
                      </Button>
                    )}

                    {/* Seller actions */}
                    {!isBuyer && transaction.status === 'pending' && (
                      <Button
                        variant="outline"
                        size="sm"
                        className="border-blue-500 text-blue-400 hover:bg-blue-500/10"
                        onClick={() => handleMarkContractCreated(transaction.id)}
                      >
                        <ClipboardList className="h-4 w-4 mr-1" />
                        Mark Contract Created
                      </Button>
                    )}

                    {/* Cancel action (both parties) */}
                    {(transaction.status === 'pending' || transaction.status === 'contract_created') && (
                      <Button
                        variant="outline"
                        size="sm"
                        className="border-red-500 text-red-400 hover:bg-red-500/10"
                        onClick={() => handleCancelPurchase(transaction.id)}
                      >
                        <XCircle className="h-4 w-4 mr-1" />
                        Cancel
                      </Button>
                    )}

                    {/* No actions for completed/cancelled */}
                    {(transaction.status === 'completed' || transaction.status === 'cancelled') && (
                      <span className="text-xs text-text-muted">-</span>
                    )}
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    );
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div>
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="border-b border-overlay-medium bg-transparent w-full justify-start rounded-none p-0 h-auto mb-6">
          <TabsTrigger
            value="purchases"
            className="text-text-secondary data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none bg-transparent px-4 py-2"
          >
            My Purchases ({buyerHistory.length})
          </TabsTrigger>
          <TabsTrigger
            value="sales"
            className="text-text-secondary data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none bg-transparent px-4 py-2"
          >
            My Sales ({sellerHistory.length})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="purchases">{renderTransactionsTable(buyerHistory, true)}</TabsContent>
        <TabsContent value="sales">{renderTransactionsTable(sellerHistory, false)}</TabsContent>
      </Tabs>
    </div>
  );
}
