import { useState, useEffect, useRef } from 'react';
import { useSession } from 'next-auth/react';
import { ClipboardList, XCircle, MapPin, User, Copy, ChevronDown, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Collapsible, CollapsibleTrigger, CollapsibleContent } from '@/components/ui/collapsible';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { toast } from '@/components/ui/sonner';
import { cn } from '@/lib/utils';

type PendingSale = {
  id: number;
  forSaleItemId: number;
  buyerUserId: number;
  buyerName: string;
  sellerUserId: number;
  typeId: number;
  typeName: string;
  locationId: number;
  locationName: string;
  quantityPurchased: number;
  pricePerUnit: number;
  totalPrice: number;
  status: string;
  contractKey?: string;
  transactionNotes?: string;
  buyOrderId?: number;
  isAutoFulfilled: boolean;
  purchasedAt: string;
};

type AggregatedItem = {
  typeId: number;
  typeName: string;
  totalQuantity: number;
  totalPrice: number;
  pricePerUnit: number;
  hasAutoFulfilled: boolean;
  notes: string[];
  purchaseIds: number[];
  latestPurchasedAt: string;
};

type GroupedSale = {
  buyerUserId: number;
  buyerName: string;
  locationId: number;
  locationName: string;
  items: PendingSale[];
  aggregatedItems: AggregatedItem[];
  totalValue: number;
  contractKey?: string;
};

export default function PendingSales() {
  const { data: session } = useSession();
  const [pendingSales, setPendingSales] = useState<PendingSale[]>([]);
  const [loading, setLoading] = useState(true);
  const [openGroups, setOpenGroups] = useState<Set<string>>(new Set());
  const contractKeyCache = useRef<Map<string, string>>(new Map());

  useEffect(() => {
    if (session) {
      fetchPendingSales();
    }
  }, [session]);

  const fetchPendingSales = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/purchases/pending-sales');

      if (response.ok) {
        const data = await response.json();
        setPendingSales(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch pending sales:', error);
    } finally {
      setLoading(false);
    }
  };

  const groupSales = (): GroupedSale[] => {
    const groups: Map<string, GroupedSale> = new Map();

    pendingSales.forEach((sale) => {
      const key = `${sale.buyerUserId}-${sale.locationId}`;

      if (!groups.has(key)) {
        let contractKey = sale.contractKey;
        if (!contractKey) {
          if (!contractKeyCache.current.has(key)) {
            contractKeyCache.current.set(key, generateContractKey(sale.buyerUserId, sale.locationId));
          }
          contractKey = contractKeyCache.current.get(key)!;
        }

        groups.set(key, {
          buyerUserId: sale.buyerUserId,
          buyerName: sale.buyerName,
          locationId: sale.locationId,
          locationName: sale.locationName,
          items: [],
          aggregatedItems: [],
          totalValue: 0,
          contractKey: contractKey,
        });
      }

      const group = groups.get(key)!;
      group.items.push(sale);
      group.totalValue += sale.totalPrice;
    });

    // Aggregate items by typeId within each group
    for (const group of groups.values()) {
      const byType = new Map<number, AggregatedItem>();
      for (const sale of group.items) {
        const existing = byType.get(sale.typeId);
        if (existing) {
          existing.totalQuantity += sale.quantityPurchased;
          existing.totalPrice += sale.totalPrice;
          existing.pricePerUnit = existing.totalPrice / existing.totalQuantity;
          existing.hasAutoFulfilled = existing.hasAutoFulfilled || sale.isAutoFulfilled;
          existing.purchaseIds.push(sale.id);
          if (sale.transactionNotes) existing.notes.push(sale.transactionNotes);
          if (sale.purchasedAt > existing.latestPurchasedAt) {
            existing.latestPurchasedAt = sale.purchasedAt;
          }
        } else {
          byType.set(sale.typeId, {
            typeId: sale.typeId,
            typeName: sale.typeName,
            totalQuantity: sale.quantityPurchased,
            totalPrice: sale.totalPrice,
            pricePerUnit: sale.pricePerUnit,
            hasAutoFulfilled: sale.isAutoFulfilled,
            notes: sale.transactionNotes ? [sale.transactionNotes] : [],
            purchaseIds: [sale.id],
            latestPurchasedAt: sale.purchasedAt,
          });
        }
      }
      group.aggregatedItems = Array.from(byType.values()).sort((a, b) =>
        a.typeName.localeCompare(b.typeName)
      );
    }

    return Array.from(groups.values()).sort((a, b) => {
      if (a.locationName !== b.locationName) {
        return a.locationName.localeCompare(b.locationName);
      }
      return a.buyerName.localeCompare(b.buyerName);
    });
  };

  const generateContractKey = (buyerUserId: number, locationId: number): string => {
    const timestamp = Date.now();
    return `PT-${buyerUserId}-${locationId}-${timestamp}`;
  };

  const handleMarkGroupContractCreated = async (group: GroupedSale) => {
    const contractKey = group.contractKey!;

    try {
      await Promise.all(
        group.items.map(item =>
          fetch(`/api/purchases/${item.id}/mark-contract-created`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ contractKey }),
          })
        )
      );

      await fetchPendingSales();
      toast.success(`Marked ${group.items.length} contract${group.items.length !== 1 ? 's' : ''} as created`);
    } catch (error) {
      console.error('Failed to mark contracts created:', error);
      toast.error('Failed to mark contracts created');
    }
  };

  const handleCancel = async (purchaseIds: number[]) => {
    const count = purchaseIds.length;
    const message = count === 1
      ? 'Are you sure you want to cancel this sale? The quantity will be restored to the listing.'
      : `Are you sure you want to cancel ${count} sales of this item? The quantities will be restored to listings.`;
    if (!confirm(message)) {
      return;
    }

    try {
      const results = await Promise.all(
        purchaseIds.map(id =>
          fetch(`/api/purchases/${id}/cancel`, { method: 'POST' })
        )
      );

      const failed = results.filter(r => !r.ok).length;
      await fetchPendingSales();
      if (failed === 0) {
        toast.success(count === 1 ? 'Sale cancelled successfully' : `${count} sales cancelled successfully`);
      } else {
        toast.error(`${failed} of ${count} cancellations failed`);
      }
    } catch (error) {
      console.error('Failed to cancel sale:', error);
      toast.error('Failed to cancel sale');
    }
  };

  const handleCopyBuyerName = async (buyerName: string) => {
    try {
      await navigator.clipboard.writeText(buyerName);
      toast.success(`Copied "${buyerName}" to clipboard`);
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
      toast.error('Failed to copy to clipboard');
    }
  };

  const handleCopyContractKey = async (contractKey: string) => {
    try {
      await navigator.clipboard.writeText(contractKey);
      toast.success(`Copied "${contractKey}" to clipboard`);
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
      toast.error('Failed to copy to clipboard');
    }
  };

  const handleCopyTotal = async (totalValue: number) => {
    try {
      await navigator.clipboard.writeText(totalValue.toString());
      toast.success(`Copied "${totalValue.toLocaleString()} ISK" to clipboard`);
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
      toast.error('Failed to copy to clipboard');
    }
  };

  const toggleGroup = (key: string) => {
    setOpenGroups(prev => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key); else next.add(key);
      return next;
    });
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (pendingSales.length === 0) {
    return (
      <div className="bg-background-panel rounded-sm border border-overlay-subtle p-8 text-center">
        <h3 className="text-lg font-semibold text-text-secondary">No pending sales</h3>
        <p className="text-sm text-text-muted mt-1">
          When buyers request to purchase your items, they will appear here.
        </p>
      </div>
    );
  }

  const groupedSales = groupSales();

  return (
    <div>
      <h3 className="text-lg font-semibold text-text-emphasis mb-1">
        Pending Sales ({pendingSales.length} items in {groupedSales.length} groups)
      </h3>
      <p className="text-sm text-text-secondary mb-4">
        Sales are grouped by purchaser and station. Copy the contract key, create the in-game contract with the key in the description, then mark as &quot;Contract Created&quot;.
      </p>

      {groupedSales.map((group, index) => {
        const key = `${group.buyerUserId}-${group.locationId}`;
        const isOpen = openGroups.has(key) || (index === 0 && !openGroups.size);
        return (
          <Collapsible
            key={key}
            open={isOpen}
            onOpenChange={() => toggleGroup(key)}
            className="border border-overlay-subtle rounded-sm mb-2 bg-background-panel"
          >
            <CollapsibleTrigger className="flex items-center justify-between w-full px-4 py-3 hover:bg-interactive-hover text-left">
              <div className="flex items-center gap-3 flex-1 flex-wrap">
                <User className="h-4 w-4 text-text-muted shrink-0" />
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <button
                        className="flex items-center gap-1 group"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleCopyBuyerName(group.buyerName);
                        }}
                      >
                        <span className="font-semibold text-text-emphasis group-hover:text-primary group-hover:underline">
                          {group.buyerName}
                        </span>
                        <Copy className="h-3.5 w-3.5 text-text-muted group-hover:text-primary" />
                      </button>
                    </TooltipTrigger>
                    <TooltipContent>Click to copy buyer name</TooltipContent>
                  </Tooltip>
                </TooltipProvider>
                <span className="text-text-muted">•</span>
                <MapPin className="h-4 w-4 text-text-muted shrink-0" />
                <span className="text-text-emphasis">{group.locationName}</span>
                <span className="text-text-muted">•</span>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <button
                        className="flex items-center gap-1 group"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleCopyTotal(group.totalValue);
                        }}
                      >
                        <span className="text-lg font-semibold text-teal-success group-hover:text-teal-success group-hover:underline">
                          {group.totalValue.toLocaleString()} ISK
                        </span>
                        <Copy className="h-3.5 w-3.5 text-teal-success group-hover:text-teal-success" />
                      </button>
                    </TooltipTrigger>
                    <TooltipContent>Click to copy total value</TooltipContent>
                  </Tooltip>
                </TooltipProvider>
                <span className="text-sm text-text-muted">
                  {group.aggregatedItems.length} item{group.aggregatedItems.length !== 1 ? 's' : ''}
                </span>
              </div>
              <ChevronDown className={cn("h-4 w-4 text-text-muted transition-transform shrink-0 ml-2", isOpen && "rotate-180")} />
            </CollapsibleTrigger>
            <CollapsibleContent>
              <div className="px-4 pb-4">
                <div className="mb-4 flex gap-4 items-start flex-wrap">
                  <div className="flex flex-col gap-1">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-text-secondary">Contract Key:</span>
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button
                              className="flex items-center gap-1 px-3 py-1.5 rounded-sm bg-primary text-background-void hover:bg-primary-muted font-mono font-bold text-sm"
                              onClick={() => handleCopyContractKey(group.contractKey!)}
                            >
                              {group.contractKey}
                              <Copy className="h-3.5 w-3.5" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent>Click to copy contract key</TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </div>
                    <span className="text-xs text-text-muted italic ml-1">
                      Copy this and paste into the in-game contract description
                    </span>
                  </div>
                  <Button
                    onClick={() => handleMarkGroupContractCreated(group)}
                    className="bg-emerald-600 hover:bg-emerald-700 mt-0.5"
                    size="sm"
                  >
                    <ClipboardList className="h-4 w-4 mr-1" />
                    Mark All as Contract Created
                  </Button>
                </div>

                <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
                  <Table>
                    <TableHeader>
                      <TableRow className="bg-background-void">
                        <TableHead>Item</TableHead>
                        <TableHead className="text-right">Quantity</TableHead>
                        <TableHead className="text-right">Price/Unit</TableHead>
                        <TableHead className="text-right">Total</TableHead>
                        {group.aggregatedItems.some(a => a.notes.length > 0) && <TableHead>Notes</TableHead>}
                        <TableHead className="text-center">Actions</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {group.aggregatedItems.map((agg) => (
                        <TableRow key={agg.typeId} className="bg-background-panel hover:bg-interactive-hover">
                          <TableCell className="text-text-emphasis">
                            <div className="flex items-center gap-1.5">
                              {agg.typeName}
                              {agg.hasAutoFulfilled && (
                                <Badge className="text-[0.65rem] font-semibold h-5 bg-teal-success/15 text-teal-success border border-teal-success/30 hover:bg-teal-success/20 cursor-default">
                                  Auto
                                </Badge>
                              )}
                            </div>
                          </TableCell>
                          <TableCell className="text-right text-text-emphasis">{agg.totalQuantity.toLocaleString()}</TableCell>
                          <TableCell className="text-right text-text-emphasis">{agg.pricePerUnit.toLocaleString()} ISK</TableCell>
                          <TableCell className="text-right">
                            <span className="font-semibold text-teal-success">
                              {agg.totalPrice.toLocaleString()} ISK
                            </span>
                          </TableCell>
                          {group.aggregatedItems.some(a => a.notes.length > 0) && (
                            <TableCell>
                              {agg.notes.length > 0 && (
                                <span className="text-xs text-text-muted">{agg.notes.join('; ')}</span>
                              )}
                            </TableCell>
                          )}
                          <TableCell className="text-center">
                            <Button
                              onClick={() => handleCancel(agg.purchaseIds)}
                              variant="outline"
                              size="sm"
                              className="border-red-500 text-red-400 hover:bg-red-500/10"
                            >
                              <XCircle className="h-4 w-4 mr-1" />
                              Cancel
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              </div>
            </CollapsibleContent>
          </Collapsible>
        );
      })}
    </div>
  );
}
