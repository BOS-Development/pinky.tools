import { useState, useEffect } from 'react';
import { useSession } from 'next-auth/react';
import { ShoppingCart, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { toast } from '@/components/ui/sonner';
import { formatISK, formatNumber } from '@industry-tool/utils/formatting';

type ForSaleListing = {
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
};

export default function MarketplaceBrowser() {
  const { data: session } = useSession();
  const [listings, setListings] = useState<ForSaleListing[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [purchaseDialogOpen, setPurchaseDialogOpen] = useState(false);
  const [selectedListing, setSelectedListing] = useState<ForSaleListing | null>(null);
  const [purchaseQuantity, setPurchaseQuantity] = useState('');
  const [submittingPurchase, setSubmittingPurchase] = useState(false);

  useEffect(() => {
    if (session) {
      fetchListings();
    }
  }, [session]);

  const fetchListings = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/for-sale/browse');
      if (response.ok) {
        const data = await response.json();
        setListings(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch marketplace listings:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleOpenPurchaseDialog = (listing: ForSaleListing) => {
    setSelectedListing(listing);
    setPurchaseQuantity(listing.quantityAvailable.toLocaleString());
    setPurchaseDialogOpen(true);
  };

  const handlePurchaseQuantityChange = (value: string) => {
    const numericValue = value.replace(/\D/g, '');
    const formatted = numericValue ? parseInt(numericValue).toLocaleString() : '';
    setPurchaseQuantity(formatted);
  };

  const handlePurchase = async () => {
    if (!selectedListing) return;

    const quantity = parseInt(purchaseQuantity.replace(/,/g, ''));
    if (quantity <= 0 || quantity > selectedListing.quantityAvailable) {
      toast.error('Invalid quantity');
      return;
    }

    setSubmittingPurchase(true);
    try {
      const response = await fetch('/api/purchases', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          forSaleItemId: selectedListing.id,
          quantityPurchased: quantity,
        }),
      });

      if (response.ok) {
        setPurchaseDialogOpen(false);
        await fetchListings();
        toast.success('Purchase successful');
      } else {
        const error = await response.json();
        toast.error(error.error || 'Purchase failed');
      }
    } catch (error) {
      console.error('Purchase failed:', error);
      toast.error('Purchase failed');
    } finally {
      setSubmittingPurchase(false);
    }
  };

  const filteredListings = listings.filter(listing =>
    listing.typeName.toLowerCase().includes(searchQuery.toLowerCase()) ||
    listing.ownerName.toLowerCase().includes(searchQuery.toLowerCase()) ||
    listing.locationName.toLowerCase().includes(searchQuery.toLowerCase())
  );

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div>
      <div className="mb-4">
        <Input
          placeholder="Search by item name, seller, or location"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
        />
      </div>

      {filteredListings.length === 0 ? (
        <div className="bg-background-panel rounded-sm border border-overlay-subtle p-8 text-center">
          <h3 className="text-lg font-semibold text-text-secondary">No listings available</h3>
          <p className="text-sm text-text-muted mt-1">
            {listings.length === 0
              ? "Your contacts haven't listed any items for sale, or they haven't granted you browse permission."
              : "No listings match your search."}
          </p>
        </div>
      ) : (
        <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
          <Table>
            <TableHeader>
              <TableRow className="bg-background-void">
                <TableHead>Item</TableHead>
                <TableHead>Seller</TableHead>
                <TableHead>Location</TableHead>
                <TableHead className="text-right">Quantity</TableHead>
                <TableHead className="text-right">Price per Unit</TableHead>
                <TableHead className="text-right">Total Value</TableHead>
                <TableHead>Notes</TableHead>
                <TableHead className="text-center">Action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredListings.map((listing) => (
                <TableRow key={listing.id} className="bg-background-panel hover:bg-interactive-hover">
                  <TableCell>
                    <span className="font-medium text-text-emphasis">{listing.typeName}</span>
                  </TableCell>
                  <TableCell>
                    <Badge className="bg-interactive-selected border border-border-active text-blue-science hover:bg-interactive-active cursor-default">
                      {listing.ownerName}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <span className="text-sm text-text-secondary">{listing.locationName}</span>
                  </TableCell>
                  <TableCell className="text-right text-text-emphasis">
                    <span className="text-sm">{formatNumber(listing.quantityAvailable)}</span>
                  </TableCell>
                  <TableCell className="text-right">
                    <span className="text-sm font-medium text-text-emphasis">{formatISK(listing.pricePerUnit)}</span>
                  </TableCell>
                  <TableCell className="text-right">
                    <span className="text-sm font-semibold text-teal-success">
                      {formatISK(listing.quantityAvailable * listing.pricePerUnit)}
                    </span>
                  </TableCell>
                  <TableCell>
                    {listing.notes && (
                      <span className="text-xs text-text-secondary">{listing.notes}</span>
                    )}
                  </TableCell>
                  <TableCell className="text-center">
                    <Button
                      size="sm"
                      onClick={() => handleOpenPurchaseDialog(listing)}
                    >
                      <ShoppingCart className="h-4 w-4 mr-1" />
                      Buy
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Purchase Dialog */}
      <Dialog open={purchaseDialogOpen} onOpenChange={setPurchaseDialogOpen}>
        <DialogContent className="max-w-sm bg-background-panel border-overlay-medium">
          <DialogHeader>
            <DialogTitle className="text-text-emphasis">Purchase Item</DialogTitle>
          </DialogHeader>
          {selectedListing && (
            <div className="flex flex-col gap-2 pt-1">
              <p className="text-sm text-text-emphasis">
                <strong>Item:</strong> {selectedListing.typeName}
              </p>
              <p className="text-sm text-text-emphasis">
                <strong>Seller:</strong> {selectedListing.ownerName}
              </p>
              <p className="text-sm text-text-emphasis">
                <strong>Location:</strong> {selectedListing.locationName}
              </p>
              <p className="text-sm text-text-emphasis">
                <strong>Price per Unit:</strong> {selectedListing.pricePerUnit.toLocaleString()} ISK
              </p>
              <p className="text-sm text-text-emphasis mb-2">
                <strong>Available:</strong> {selectedListing.quantityAvailable.toLocaleString()}
              </p>

              <div>
                <label className="text-sm text-text-secondary mb-1 block">Quantity to Purchase</label>
                <Input
                  type="text"
                  value={purchaseQuantity}
                  onChange={(e) => handlePurchaseQuantityChange(e.target.value)}
                  placeholder="0"
                />
                <p className="text-xs text-text-muted mt-1">
                  {purchaseQuantity
                    ? `Total Cost: ${(
                        parseInt(purchaseQuantity.replace(/,/g, '')) * selectedListing.pricePerUnit
                      ).toLocaleString()} ISK`
                    : `Max: ${selectedListing.quantityAvailable.toLocaleString()}`}
                </p>
              </div>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setPurchaseDialogOpen(false)}>Cancel</Button>
            <Button
              onClick={handlePurchase}
              disabled={!purchaseQuantity || submittingPurchase}
            >
              {submittingPurchase ? 'Purchasing...' : 'Confirm Purchase'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
