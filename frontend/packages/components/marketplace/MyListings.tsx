import { useState, useEffect, useRef, useMemo } from 'react';
import { useSession } from "next-auth/react";
import { Edit, Trash2, Search, Repeat, Tag, ShoppingBag } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { toast } from '@/components/ui/sonner';
import { cn } from '@/lib/utils';
import Loading from "@industry-tool/components/loading";

const AutoSellIcon = () => (
  <span className="relative inline-flex">
    <Tag className="h-4 w-4" />
    <Repeat className="h-2.5 w-2.5 absolute -top-0.5 -right-0.5" />
  </span>
);

export type ForSaleItem = {
  id: number;
  userId: number;
  typeId: number;
  typeName: string;
  ownerType: string;
  ownerId: number;
  ownerName: string;
  locationId: number;
  locationName: string;
  containerId?: number;
  divisionNumber?: number;
  quantityAvailable: number;
  pricePerUnit: number;
  notes?: string;
  autoSellContainerId?: number;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

type ListingFormData = {
  typeId: number;
  ownerType: string;
  ownerId: number;
  locationId: number;
  containerId?: number;
  divisionNumber?: number;
  quantityAvailable: number;
  pricePerUnit: number;
  notes?: string;
};

export default function MyListings() {
  const { data: session } = useSession();
  const [listings, setListings] = useState<ForSaleItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [selectedListing, setSelectedListing] = useState<ForSaleItem | null>(null);
  const [formData, setFormData] = useState<Partial<ListingFormData>>({});
  const hasFetchedRef = useRef(false);

  useEffect(() => {
    if (session && !hasFetchedRef.current) {
      hasFetchedRef.current = true;
      fetchListings();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [session]);

  const fetchListings = async () => {
    if (!session) return;

    setLoading(true);
    try {
      const response = await fetch('/api/for-sale');
      if (response.ok) {
        const data: ForSaleItem[] = await response.json();
        setListings(data || []);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleEditClick = (listing: ForSaleItem) => {
    setSelectedListing(listing);
    setFormData({
      quantityAvailable: listing.quantityAvailable,
      pricePerUnit: listing.pricePerUnit,
      notes: listing.notes,
    });
    setEditDialogOpen(true);
  };

  const handleEditSave = async () => {
    if (!selectedListing || !session) return;

    try {
      const response = await fetch(`/api/for-sale/${selectedListing.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      });

      if (response.ok) {
        toast.success('Listing updated successfully');
        setEditDialogOpen(false);
        setSelectedListing(null);
        setFormData({});
        await fetchListings();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Failed to update listing');
      }
    } catch {
      toast.error('Failed to update listing');
    }
  };

  const handleDelete = async (listingId: number) => {
    if (!confirm('Are you sure you want to delete this listing?')) return;

    try {
      const response = await fetch(`/api/for-sale/${listingId}`, {
        method: 'DELETE',
      });

      if (response.ok) {
        toast.success('Listing deleted successfully');
        await fetchListings();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Failed to delete listing');
      }
    } catch {
      toast.error('Failed to delete listing');
    }
  };

  // Filter listings based on search
  const filteredListings = useMemo(() => {
    if (!searchQuery) return listings;

    const query = searchQuery.toLowerCase();
    return listings.filter(
      (item) =>
        item.typeName.toLowerCase().includes(query) ||
        item.ownerName.toLowerCase().includes(query) ||
        item.locationName.toLowerCase().includes(query)
    );
  }, [listings, searchQuery]);

  // Calculate totals
  const totalValue = useMemo(() => {
    return filteredListings.reduce((sum, item) => {
      return sum + (item.quantityAvailable * item.pricePerUnit);
    }, 0);
  }, [filteredListings]);

  if (!session) {
    return null;
  }

  if (loading) {
    return <Loading />;
  }

  return (
    <div className="w-full mt-4 mb-4">
      <div className="mb-3">
        <h1 className="text-2xl font-bold text-text-emphasis mb-4">My Listings</h1>

        {/* Summary Stats */}
        <div className="flex gap-4 mb-6">
          <Card className="flex-1 bg-background-panel border-overlay-subtle">
            <CardContent className="p-4">
              <h3 className="text-lg font-semibold text-text-secondary mb-1">Active Listings</h3>
              <p className="text-3xl font-bold text-text-emphasis">{filteredListings.length}</p>
            </CardContent>
          </Card>
          <Card className="flex-1 bg-background-panel border-overlay-subtle">
            <CardContent className="p-4">
              <h3 className="text-lg font-semibold text-text-secondary mb-1">Total Value</h3>
              <p className="text-3xl font-bold text-text-emphasis">
                {totalValue.toLocaleString(undefined, { maximumFractionDigits: 0 })} ISK
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Search */}
        <div className="mb-4 relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-text-muted" />
          <Input
            className="pl-9"
            placeholder="Search items, owners, or locations..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
      </div>

      {/* Listings Table */}
      {filteredListings.length === 0 ? (
        <div className="empty-state">
          <ShoppingBag className="empty-state-icon" />
          <p className="empty-state-title">
            {listings.length === 0
              ? 'No active listings. Create your first listing to get started!'
              : 'No items match your search.'}
          </p>
        </div>
      ) : (
        <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
          <Table>
            <TableHeader>
              <TableRow className="bg-background-void">
                <TableHead>Item</TableHead>
                <TableHead>Owner</TableHead>
                <TableHead>Location</TableHead>
                <TableHead className="text-right">Quantity</TableHead>
                <TableHead className="text-right">Price/Unit</TableHead>
                <TableHead className="text-right">Total Value</TableHead>
                <TableHead>Notes</TableHead>
                <TableHead className="text-center">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredListings.map((item, idx) => (
                <TableRow
                  key={item.id}
                  className={cn(idx % 2 === 0 ? 'bg-background-panel' : 'bg-background-void', 'hover:bg-interactive-hover')}
                >
                  <TableCell className="font-semibold text-text-emphasis">
                    <div className="flex items-center gap-2">
                      {item.typeName}
                      {item.autoSellContainerId && (
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Badge
                                className="text-[0.65rem] font-semibold h-[22px] bg-interactive-active text-primary border border-border-active hover:bg-interactive-active cursor-default flex items-center gap-1"
                              >
                                <AutoSellIcon />
                                Auto
                              </Badge>
                            </TooltipTrigger>
                            <TooltipContent>Auto-managed listing — changes will be overwritten on next sync</TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      )}
                    </div>
                  </TableCell>
                  <TableCell className="text-text-secondary">{item.ownerName}</TableCell>
                  <TableCell className="text-text-secondary">{item.locationName}</TableCell>
                  <TableCell className="text-right text-text-emphasis">{item.quantityAvailable.toLocaleString()}</TableCell>
                  <TableCell className="text-right text-text-emphasis">
                    {item.pricePerUnit.toLocaleString(undefined, { maximumFractionDigits: 2 })}
                  </TableCell>
                  <TableCell className="text-right text-text-emphasis">
                    {(item.quantityAvailable * item.pricePerUnit).toLocaleString(undefined, { maximumFractionDigits: 0 })}
                  </TableCell>
                  <TableCell className="text-text-secondary">{item.notes || '-'}</TableCell>
                  <TableCell className="text-center">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 text-blue-science hover:text-blue-science hover:bg-blue-science/10"
                      onClick={() => handleEditClick(item)}
                      aria-label="edit"
                    >
                      <Edit className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 text-rose-danger hover:text-rose-danger hover:bg-rose-danger/10"
                      onClick={() => handleDelete(item.id)}
                      aria-label="delete"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Edit Dialog */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent className="max-w-sm bg-background-panel border-overlay-medium">
          <DialogHeader>
            <DialogTitle className="text-text-emphasis">Edit Listing</DialogTitle>
          </DialogHeader>
          {selectedListing && (
            <div className="flex flex-col gap-4 mt-1">
              <p className="text-sm text-text-secondary">
                Item: <strong className="text-text-emphasis">{selectedListing.typeName}</strong>
              </p>
              <p className="text-sm text-text-secondary">
                Location: <strong className="text-text-emphasis">{selectedListing.locationName}</strong>
              </p>

              <div>
                <label className="text-sm text-text-secondary mb-1 block">Quantity Available</label>
                <Input
                  type="number"
                  value={formData.quantityAvailable || ''}
                  onChange={(e) => setFormData({ ...formData, quantityAvailable: parseInt(e.target.value) })}
                  min={1}
                />
              </div>

              <div>
                <label className="text-sm text-text-secondary mb-1 block">Price Per Unit (ISK)</label>
                <Input
                  type="number"
                  value={formData.pricePerUnit || ''}
                  onChange={(e) => setFormData({ ...formData, pricePerUnit: parseInt(e.target.value) })}
                  min={0}
                />
              </div>

              <div>
                <label className="text-sm text-text-secondary mb-1 block">Notes (optional)</label>
                <textarea
                  rows={3}
                  className="w-full rounded-sm border border-overlay-strong bg-background-void text-text-emphasis text-sm px-3 py-2 resize-none focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary"
                  value={formData.notes || ''}
                  onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                />
              </div>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditDialogOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleEditSave}
              disabled={!formData.quantityAvailable || formData.quantityAvailable <= 0 || formData.pricePerUnit === undefined || formData.pricePerUnit < 0}
            >
              Save Changes
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
