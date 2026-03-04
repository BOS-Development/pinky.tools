import { useState, useEffect } from 'react';
import { useSession } from 'next-auth/react';
import { Loader2, Plus, Pencil, Trash2 } from 'lucide-react';
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

type CharacterSlotInventory = {
  characterId: number;
  characterName: string;
  slotsByActivity: Record<string, { slotsAvailable: number }>;
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

export default function MyListings() {
  const { data: session } = useSession();
  const [listings, setListings] = useState<JobSlotRentalListing[]>([]);
  const [inventory, setInventory] = useState<CharacterSlotInventory[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedListing, setSelectedListing] = useState<JobSlotRentalListing | null>(null);
  const [formData, setFormData] = useState<{
    characterId: number;
    activityType: string;
    slotsListed: number;
    priceAmount: number;
    pricingUnit: string;
    locationName: string;
    notes: string;
  }>({
    characterId: 0,
    activityType: '',
    slotsListed: 1,
    priceAmount: 0,
    pricingUnit: 'per_slot_day',
    locationName: '',
    notes: '',
  });

  useEffect(() => {
    if (session) {
      fetchListings();
      fetchInventory();
    }
  }, [session]);

  const fetchListings = async () => {
    try {
      const response = await fetch('/api/job-slots/listings');
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

  const fetchInventory = async () => {
    try {
      const response = await fetch('/api/job-slots/inventory');
      if (response.ok) {
        const data = await response.json();
        setInventory(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch inventory:', error);
    }
  };

  const handleCreate = () => {
    setSelectedListing(null);
    setFormData({
      characterId: 0,
      activityType: '',
      slotsListed: 1,
      priceAmount: 0,
      pricingUnit: 'per_slot_day',
      locationName: '',
      notes: '',
    });
    setDialogOpen(true);
  };

  const handleEdit = (listing: JobSlotRentalListing) => {
    setSelectedListing(listing);
    setFormData({
      characterId: listing.characterId,
      activityType: listing.activityType,
      slotsListed: listing.slotsListed,
      priceAmount: listing.priceAmount,
      pricingUnit: listing.pricingUnit,
      locationName: listing.locationName || '',
      notes: listing.notes || '',
    });
    setDialogOpen(true);
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this listing?')) return;

    try {
      const response = await fetch(`/api/job-slots/listings/${id}`, { method: 'DELETE' });
      if (response.ok) {
        toast.success('Listing deleted successfully');
        fetchListings();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Failed to delete listing');
      }
    } catch (error) {
      console.error('Delete failed:', error);
      toast.error('Failed to delete listing');
    }
  };

  const handleSave = async () => {
    if (!formData.characterId || !formData.activityType || formData.slotsListed <= 0 || formData.priceAmount < 0) {
      toast.error('Please fill in all required fields');
      return;
    }

    try {
      const url = selectedListing ? `/api/job-slots/listings/${selectedListing.id}` : '/api/job-slots/listings';
      const method = selectedListing ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          characterId: formData.characterId,
          activityType: formData.activityType,
          slotsListed: formData.slotsListed,
          priceAmount: formData.priceAmount,
          pricingUnit: formData.pricingUnit,
          locationName: formData.locationName || null,
          notes: formData.notes || null,
        }),
      });

      if (response.ok) {
        setDialogOpen(false);
        toast.success(selectedListing ? 'Listing updated' : 'Listing created');
        fetchListings();
        fetchInventory();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Failed to save listing');
      }
    } catch (error) {
      console.error('Save failed:', error);
      toast.error('Failed to save listing');
    }
  };

  const getAvailableActivities = () => {
    if (!formData.characterId) return [];
    const char = inventory.find(c => c.characterId === formData.characterId);
    if (!char) return [];
    return Object.entries(char.slotsByActivity)
      .filter(([_, info]) => info.slotsAvailable > 0)
      .map(([activityType]) => activityType);
  };

  const getMaxSlots = () => {
    if (!formData.characterId || !formData.activityType) return 0;
    const char = inventory.find(c => c.characterId === formData.characterId);
    if (!char) return 0;
    return char.slotsByActivity[formData.activityType]?.slotsAvailable || 0;
  };

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
        <h2 className="text-xl font-semibold text-[#e2e8f0]">My Slot Listings</h2>
        <Button onClick={handleCreate}>
          <Plus className="h-4 w-4 mr-1" />
          Create Listing
        </Button>
      </div>

      {listings.length === 0 ? (
        <div className="bg-[#12151f] rounded-sm border border-[rgba(148,163,184,0.1)] p-8 text-center">
          <h3 className="text-lg font-semibold text-[#94a3b8]">No listings yet</h3>
          <p className="text-sm text-[#64748b] mt-1">
            Create your first listing to rent out idle job slots.
          </p>
        </div>
      ) : (
        <div className="overflow-x-auto rounded-sm border border-[rgba(148,163,184,0.1)]">
          <Table>
            <TableHeader>
              <TableRow className="bg-[#0f1219]">
                <TableHead>Character</TableHead>
                <TableHead>Activity</TableHead>
                <TableHead className="text-right">Slots Listed</TableHead>
                <TableHead className="text-right">Price</TableHead>
                <TableHead>Pricing Unit</TableHead>
                <TableHead>Location</TableHead>
                <TableHead>Notes</TableHead>
                <TableHead className="text-center">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {listings.map((listing) => (
                <TableRow key={listing.id} className="hover:bg-[rgba(0,212,255,0.04)]">
                  <TableCell className="text-[#e2e8f0]">{listing.characterName}</TableCell>
                  <TableCell>
                    <Badge className="bg-[rgba(0,212,255,0.1)] border border-[rgba(0,212,255,0.3)] text-[#60a5fa] hover:bg-[rgba(0,212,255,0.15)] cursor-default">
                      {ACTIVITY_LABELS[listing.activityType] || listing.activityType}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right text-[#e2e8f0]">{listing.slotsListed}</TableCell>
                  <TableCell className="text-right text-[#e2e8f0]">{formatISK(listing.priceAmount)}</TableCell>
                  <TableCell className="text-[#cbd5e1]">{PRICING_UNIT_LABELS[listing.pricingUnit] || listing.pricingUnit}</TableCell>
                  <TableCell className="text-[#94a3b8]">{listing.locationName || '-'}</TableCell>
                  <TableCell className="text-[#94a3b8]">{listing.notes || '-'}</TableCell>
                  <TableCell className="text-center">
                    <button className="p-1 rounded hover:bg-[rgba(0,212,255,0.1)] text-[#00d4ff]" onClick={() => handleEdit(listing)} aria-label="Edit">
                      <Pencil className="h-4 w-4" />
                    </button>
                    <button className="p-1 rounded hover:bg-[rgba(239,68,68,0.1)] text-[#ef4444]" onClick={() => handleDelete(listing.id)} aria-label="Delete">
                      <Trash2 className="h-4 w-4" />
                    </button>
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
            <DialogTitle className="text-[#e2e8f0]">{selectedListing ? 'Edit Listing' : 'Create Listing'}</DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-3 pt-1">
            <div>
              <Label className="text-sm text-[#94a3b8] mb-1 block">Character</Label>
              <Select
                value={formData.characterId ? String(formData.characterId) : ""}
                onValueChange={(val) => setFormData({ ...formData, characterId: parseInt(val), activityType: '' })}
                disabled={!!selectedListing}
              >
                <SelectTrigger><SelectValue placeholder="Select a character" /></SelectTrigger>
                <SelectContent>
                  {inventory.map((char) => (
                    <SelectItem key={char.characterId} value={String(char.characterId)}>
                      {char.characterName}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label className="text-sm text-[#94a3b8] mb-1 block">Activity Type</Label>
              <Select
                value={formData.activityType}
                onValueChange={(val) => setFormData({ ...formData, activityType: val })}
                disabled={!formData.characterId || !!selectedListing}
              >
                <SelectTrigger><SelectValue placeholder="Select an activity" /></SelectTrigger>
                <SelectContent>
                  {getAvailableActivities().map((activity) => (
                    <SelectItem key={activity} value={activity}>
                      {ACTIVITY_LABELS[activity] || activity}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label className="text-sm text-[#94a3b8] mb-1 block">Slots to List</Label>
              <Input
                type="number"
                value={formData.slotsListed}
                onChange={(e) => setFormData({ ...formData, slotsListed: parseInt(e.target.value) || 0 })}
                min={1}
                max={getMaxSlots()}
                disabled={!formData.activityType}
              />
              <span className="text-xs text-[#64748b] mt-0.5 block">Max available: {getMaxSlots()}</span>
            </div>

            <div>
              <Label className="text-sm text-[#94a3b8] mb-1 block">Price Amount (ISK)</Label>
              <Input
                type="number"
                value={formData.priceAmount}
                onChange={(e) => setFormData({ ...formData, priceAmount: parseFloat(e.target.value) || 0 })}
                min={0}
              />
            </div>

            <div>
              <Label className="text-sm text-[#94a3b8] mb-1 block">Pricing Unit</Label>
              <Select value={formData.pricingUnit} onValueChange={(val) => setFormData({ ...formData, pricingUnit: val })}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {Object.entries(PRICING_UNIT_LABELS).map(([value, label]) => (
                    <SelectItem key={value} value={value}>{label}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label className="text-sm text-[#94a3b8] mb-1 block">Location (optional)</Label>
              <Input
                value={formData.locationName}
                onChange={(e) => setFormData({ ...formData, locationName: e.target.value })}
                placeholder="e.g., Jita 4-4"
              />
            </div>

            <div>
              <Label className="text-sm text-[#94a3b8] mb-1 block">Notes (optional)</Label>
              <textarea
                className="flex w-full rounded-sm border border-[var(--color-border-dim)] bg-[var(--color-bg-void)] px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-[var(--color-primary-cyan)]"
                rows={3}
                value={formData.notes}
                onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                placeholder="Additional information about this listing..."
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>Cancel</Button>
            <Button onClick={handleSave}>
              {selectedListing ? 'Update' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
