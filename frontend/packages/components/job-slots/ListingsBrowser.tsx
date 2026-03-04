import { useState, useEffect } from 'react';
import { useSession } from 'next-auth/react';
import { Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { toast } from '@/components/ui/sonner';
import { formatISK } from '@industry-tool/utils/formatting';

type JobSlotRentalListing = {
  id: number;
  userId: number;
  characterId: number;
  characterName: string;
  ownerName: string;
  activityType: string;
  slotsListed: number;
  priceAmount: number;
  pricingUnit: string;
  locationId: number | null;
  locationName: string;
  notes: string | null;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

const ACTIVITY_LABELS: Record<string, string> = {
  manufacturing: 'Manufacturing',
  reaction: 'Reactions',
  copying: 'Copying',
  invention: 'Invention',
  me_research: 'ME Research',
  te_research: 'TE Research',
};

const PRICING_UNIT_LABELS: Record<string, string> = {
  per_slot_day: 'Per Slot/Day',
  per_job: 'Per Job',
  flat_fee: 'Flat Fee',
};

export default function ListingsBrowser() {
  const { data: session } = useSession();
  const [listings, setListings] = useState<JobSlotRentalListing[]>([]);
  const [loading, setLoading] = useState(true);
  const [activityFilter, setActivityFilter] = useState<string>('all');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedListing, setSelectedListing] = useState<JobSlotRentalListing | null>(null);
  const [interestData, setInterestData] = useState<{
    slotsRequested: number;
    durationDays: number | null;
    message: string;
  }>({
    slotsRequested: 1,
    durationDays: null,
    message: '',
  });

  useEffect(() => {
    if (session) {
      fetchListings();
    }
  }, [session]);

  const fetchListings = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/job-slots/listings/browse');
      if (response.ok) {
        const data = await response.json();
        setListings(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch listings:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleExpressInterest = (listing: JobSlotRentalListing) => {
    setSelectedListing(listing);
    setInterestData({ slotsRequested: 1, durationDays: null, message: '' });
    setDialogOpen(true);
  };

  const handleSubmitInterest = async () => {
    if (!selectedListing || interestData.slotsRequested <= 0 || interestData.slotsRequested > selectedListing.slotsListed) {
      toast.error('Invalid slots requested');
      return;
    }

    try {
      const response = await fetch('/api/job-slots/interest', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          listingId: selectedListing.id,
          slotsRequested: interestData.slotsRequested,
          durationDays: interestData.durationDays,
          message: interestData.message || null,
        }),
      });

      if (response.ok) {
        setDialogOpen(false);
        toast.success('Interest submitted successfully');
      } else {
        const error = await response.json();
        toast.error(error.error || 'Failed to submit interest');
      }
    } catch (error) {
      console.error('Submit interest failed:', error);
      toast.error('Failed to submit interest');
    }
  };

  const filteredListings = listings.filter(
    listing => activityFilter === 'all' || listing.activityType === activityFilter
  );

  const uniqueActivities = Array.from(new Set(listings.map(l => l.activityType)));

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-[#00d4ff]" />
      </div>
    );
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-3">
        <h2 className="text-xl font-semibold text-[#e2e8f0]">Browse Listings</h2>
        <div className="min-w-[200px]">
          <Select value={activityFilter} onValueChange={setActivityFilter}>
            <SelectTrigger><SelectValue placeholder="Filter by Activity" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Activities</SelectItem>
              {uniqueActivities.map((activity) => (
                <SelectItem key={activity} value={activity}>
                  {ACTIVITY_LABELS[activity] || activity}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {filteredListings.length === 0 ? (
        <div className="bg-[#12151f] rounded-sm border border-[rgba(148,163,184,0.1)] p-8 text-center">
          <h3 className="text-lg font-semibold text-[#94a3b8]">No listings available</h3>
          <p className="text-sm text-[#64748b] mt-1">
            {listings.length === 0
              ? "Your contacts haven't listed any slots for rent, or they haven't granted you browse permission."
              : "No listings match your selected filter."}
          </p>
        </div>
      ) : (
        <div className="overflow-x-auto rounded-sm border border-[rgba(148,163,184,0.1)]">
          <Table>
            <TableHeader>
              <TableRow className="bg-[#0f1219]">
                <TableHead>Owner</TableHead>
                <TableHead>Character</TableHead>
                <TableHead>Activity</TableHead>
                <TableHead className="text-right">Slots Available</TableHead>
                <TableHead className="text-right">Price</TableHead>
                <TableHead>Pricing Unit</TableHead>
                <TableHead>Location</TableHead>
                <TableHead>Notes</TableHead>
                <TableHead className="text-center">Action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredListings.map((listing) => (
                <TableRow key={listing.id} className="hover:bg-[rgba(0,212,255,0.04)]">
                  <TableCell>
                    <Badge className="bg-[rgba(0,212,255,0.1)] border border-[rgba(0,212,255,0.3)] text-[#60a5fa] hover:bg-[rgba(0,212,255,0.15)] cursor-default">
                      {listing.ownerName}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-[#e2e8f0]">{listing.characterName}</TableCell>
                  <TableCell>
                    <Badge className="bg-[rgba(16,185,129,0.1)] border border-[rgba(16,185,129,0.3)] text-[#10b981] hover:bg-[rgba(16,185,129,0.15)] cursor-default">
                      {ACTIVITY_LABELS[listing.activityType] || listing.activityType}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right text-[#e2e8f0]">{listing.slotsListed}</TableCell>
                  <TableCell className="text-right text-[#e2e8f0]">{formatISK(listing.priceAmount)}</TableCell>
                  <TableCell className="text-[#cbd5e1]">{PRICING_UNIT_LABELS[listing.pricingUnit] || listing.pricingUnit}</TableCell>
                  <TableCell className="text-[#94a3b8]">{listing.locationName || '-'}</TableCell>
                  <TableCell>
                    {listing.notes ? (
                      <span className="text-xs text-[#94a3b8]">{listing.notes}</span>
                    ) : '-'}
                  </TableCell>
                  <TableCell className="text-center">
                    <Button size="sm" onClick={() => handleExpressInterest(listing)}>
                      Express Interest
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-md bg-[#12151f] border-[rgba(148,163,184,0.15)]">
          <DialogHeader>
            <DialogTitle className="text-[#e2e8f0]">Express Interest</DialogTitle>
          </DialogHeader>
          {selectedListing && (
            <div className="flex flex-col gap-2 pt-1">
              <p className="text-sm text-[#e2e8f0]"><strong>Owner:</strong> {selectedListing.ownerName}</p>
              <p className="text-sm text-[#e2e8f0]"><strong>Character:</strong> {selectedListing.characterName}</p>
              <p className="text-sm text-[#e2e8f0]"><strong>Activity:</strong> {ACTIVITY_LABELS[selectedListing.activityType] || selectedListing.activityType}</p>
              <p className="text-sm text-[#e2e8f0]"><strong>Price:</strong> {formatISK(selectedListing.priceAmount)} {PRICING_UNIT_LABELS[selectedListing.pricingUnit]}</p>
              <p className="text-sm text-[#e2e8f0] mb-2"><strong>Slots Available:</strong> {selectedListing.slotsListed}</p>

              <div>
                <Label className="text-sm text-[#94a3b8] mb-1 block">Slots Requested</Label>
                <Input
                  type="number"
                  value={interestData.slotsRequested}
                  onChange={(e) => setInterestData({ ...interestData, slotsRequested: parseInt(e.target.value) || 0 })}
                  min={1}
                  max={selectedListing.slotsListed}
                />
                <span className="text-xs text-[#64748b] mt-0.5 block">Max: {selectedListing.slotsListed}</span>
              </div>

              <div>
                <Label className="text-sm text-[#94a3b8] mb-1 block">Duration (days) - optional</Label>
                <Input
                  type="number"
                  value={interestData.durationDays || ''}
                  onChange={(e) => setInterestData({ ...interestData, durationDays: e.target.value ? parseInt(e.target.value) : null })}
                  min={1}
                />
                <span className="text-xs text-[#64748b] mt-0.5 block">Leave empty if flexible</span>
              </div>

              <div>
                <Label className="text-sm text-[#94a3b8] mb-1 block">Message (optional)</Label>
                <textarea
                  className="flex w-full rounded-sm border border-[var(--color-border-dim)] bg-[var(--color-bg-void)] px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-[var(--color-primary-cyan)]"
                  rows={3}
                  value={interestData.message}
                  onChange={(e) => setInterestData({ ...interestData, message: e.target.value })}
                  placeholder="Additional information for the owner..."
                />
              </div>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>Cancel</Button>
            <Button onClick={handleSubmitInterest}>Submit Interest</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
