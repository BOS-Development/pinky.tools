import { useState, useEffect, useRef } from 'react';
import { useSession } from "next-auth/react";
import Image from 'next/image';
import { getItemIconUrl } from "@industry-tool/utils/eveImages";
import { Plus, Edit, Trash2, Loader2, ChevronDown, Check } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { toast } from '@/components/ui/sonner';
import { cn } from '@/lib/utils';
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
  maxPricePerUnit: number;
  notes?: string;
  autoBuyConfigId?: number;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

type BuyOrderFormData = {
  typeId: number;
  typeName?: string;
  locationId: number;
  quantityDesired: number;
  minPricePerUnit: number;
  maxPricePerUnit: number;
  notes?: string;
};

type ItemType = {
  TypeID: number;
  TypeName: string;
  Volume: number;
  IconID?: number;
};

type StationOption = {
  stationId: number;
  name: string;
  solarSystemName: string;
};

// Async search combobox for items
function ItemSearchCombobox({
  value,
  onChange,
  disabled,
}: {
  value: ItemType | null;
  onChange: (item: ItemType | null) => void;
  disabled?: boolean;
}) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState('');
  const [options, setOptions] = useState<ItemType[]>([]);
  const [loading, setLoading] = useState(false);
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);

  const handleSearch = (val: string) => {
    setSearch(val);
    if (timeoutRef.current) clearTimeout(timeoutRef.current);
    if (!val || val.length < 2) {
      setOptions([]);
      return;
    }
    timeoutRef.current = setTimeout(async () => {
      setLoading(true);
      try {
        const response = await fetch(`/api/item-types/search?q=${encodeURIComponent(val)}`);
        if (response.ok) {
          const data = await response.json();
          setOptions(data || []);
        }
      } catch {
        setOptions([]);
      } finally {
        setLoading(false);
      }
    }, 300);
  };

  return (
    <Popover open={open && !disabled} onOpenChange={(v) => !disabled && setOpen(v)}>
      <PopoverTrigger asChild>
        <button
          role="combobox"
          disabled={disabled}
          className={cn(
            "flex h-9 w-full items-center justify-between whitespace-nowrap rounded-sm border border-overlay-strong bg-background-void px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-primary",
            disabled && "opacity-50 cursor-not-allowed",
            !value && "text-text-muted"
          )}
        >
          <span className="flex items-center gap-2 min-w-0">
            {value && (
              <Image
                src={getItemIconUrl(value.TypeID, 32)}
                alt={value.TypeName}
                width={20}
                height={20}
                className="rounded-sm shrink-0"
              />
            )}
            <span className="truncate">{value ? value.TypeName : 'Start typing to search...'}</span>
          </span>
          <ChevronDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </button>
      </PopoverTrigger>
      <PopoverContent className="p-0 w-[--radix-popover-trigger-width]" align="start">
        <div className="flex flex-col">
          <div className="p-2">
            <Input
              placeholder="Search items..."
              value={search}
              onChange={(e) => handleSearch(e.target.value)}
              className="h-8"
              autoFocus
            />
          </div>
          <div className="max-h-60 overflow-y-auto">
            {loading ? (
              <div className="flex justify-center py-4">
                <Loader2 className="h-4 w-4 animate-spin text-primary" />
              </div>
            ) : options.length === 0 ? (
              <div className="py-6 text-center text-sm text-text-muted">
                {search.length < 2 ? 'Type at least 2 characters' : 'No items found'}
              </div>
            ) : (
              options.map((option) => (
                <button
                  key={option.TypeID}
                  className={cn(
                    "flex w-full items-center gap-2 px-3 py-1.5 text-sm outline-none cursor-pointer hover:bg-status-neutral-tint",
                    value?.TypeID === option.TypeID && "bg-interactive-hover text-primary"
                  )}
                  onClick={() => {
                    onChange(option);
                    setOpen(false);
                    setSearch('');
                  }}
                >
                  <Image
                    src={getItemIconUrl(option.TypeID, 32)}
                    alt={option.TypeName}
                    width={24}
                    height={24}
                    className="rounded-sm shrink-0"
                  />
                  <span>{option.TypeName}</span>
                  {value?.TypeID === option.TypeID && <Check className="h-4 w-4 ml-auto shrink-0" />}
                </button>
              ))
            )}
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}

// Async search combobox for stations
function StationSearchCombobox({
  value,
  onChange,
}: {
  value: StationOption | null;
  onChange: (station: StationOption | null) => void;
}) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState('');
  const [options, setOptions] = useState<StationOption[]>([]);
  const [loading, setLoading] = useState(false);
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);

  const handleSearch = (val: string) => {
    setSearch(val);
    if (timeoutRef.current) clearTimeout(timeoutRef.current);
    if (!val || val.length < 2) {
      setOptions([]);
      return;
    }
    timeoutRef.current = setTimeout(async () => {
      setLoading(true);
      try {
        const response = await fetch(`/api/stations/search?q=${encodeURIComponent(val)}`);
        if (response.ok) {
          const data = await response.json();
          setOptions(data || []);
        }
      } catch {
        setOptions([]);
      } finally {
        setLoading(false);
      }
    }, 300);
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          role="combobox"
          className={cn(
            "flex h-9 w-full items-center justify-between whitespace-nowrap rounded-sm border border-overlay-strong bg-background-void px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-primary",
            !value && "text-text-muted"
          )}
        >
          <span className="truncate">{value ? value.name : 'Search for a station...'}</span>
          <ChevronDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </button>
      </PopoverTrigger>
      <PopoverContent className="p-0 w-[--radix-popover-trigger-width]" align="start">
        <div className="flex flex-col">
          <div className="p-2">
            <Input
              placeholder="Search stations..."
              value={search}
              onChange={(e) => handleSearch(e.target.value)}
              className="h-8"
              autoFocus
            />
          </div>
          <div className="max-h-60 overflow-y-auto">
            {loading ? (
              <div className="flex justify-center py-4">
                <Loader2 className="h-4 w-4 animate-spin text-primary" />
              </div>
            ) : options.length === 0 ? (
              <div className="py-6 text-center text-sm text-text-muted">
                {search.length < 2 ? 'Type at least 2 characters' : 'No stations found'}
              </div>
            ) : (
              options.map((option) => (
                <button
                  key={option.stationId}
                  className={cn(
                    "flex w-full flex-col items-start px-3 py-1.5 text-sm outline-none cursor-pointer hover:bg-status-neutral-tint",
                    value?.stationId === option.stationId && "bg-interactive-hover"
                  )}
                  onClick={() => {
                    onChange(option);
                    setOpen(false);
                    setSearch('');
                  }}
                >
                  <span className="text-text-emphasis">{option.name}</span>
                  <span className="text-xs text-text-muted">{option.solarSystemName}</span>
                </button>
              ))
            )}
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}

export default function BuyOrders() {
  const { data: session } = useSession();
  const [orders, setOrders] = useState<BuyOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedOrder, setSelectedOrder] = useState<BuyOrder | null>(null);
  const [formData, setFormData] = useState<Partial<BuyOrderFormData>>({});
  const [selectedItem, setSelectedItem] = useState<ItemType | null>(null);
  const [selectedStation, setSelectedStation] = useState<StationOption | null>(null);
  const hasFetchedRef = useRef(false);

  useEffect(() => {
    if (session && !hasFetchedRef.current) {
      hasFetchedRef.current = true;
      fetchOrders();
    }
  }, [session]);

  const fetchOrders = async () => {
    try {
      const response = await fetch('/api/buy-orders');
      if (!response.ok) throw new Error('Failed to fetch buy orders');
      const data = await response.json();
      setOrders(data);
    } catch (error) {
      console.error('Error fetching buy orders:', error);
      toast.error('Failed to load buy orders');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setSelectedOrder(null);
    setSelectedItem(null);
    setSelectedStation(null);
    setFormData({ quantityDesired: 0, minPricePerUnit: 0, maxPricePerUnit: 0, locationId: 0 });
    setDialogOpen(true);
  };

  const handleEdit = (order: BuyOrder) => {
    setSelectedOrder(order);
    setSelectedItem({
      TypeID: order.typeId,
      TypeName: order.typeName,
      Volume: 0,
    });
    setSelectedStation({
      stationId: order.locationId,
      name: order.locationName,
      solarSystemName: '',
    });
    setFormData({
      typeId: order.typeId,
      typeName: order.typeName,
      locationId: order.locationId,
      quantityDesired: order.quantityDesired,
      minPricePerUnit: order.minPricePerUnit,
      maxPricePerUnit: order.maxPricePerUnit,
      notes: order.notes,
    });
    setDialogOpen(true);
  };

  const handleDelete = async (orderId: number) => {
    if (!confirm('Are you sure you want to cancel this buy order?')) return;

    try {
      const response = await fetch(`/api/buy-orders/${orderId}`, {
        method: 'DELETE',
      });

      if (!response.ok) throw new Error('Failed to delete buy order');

      toast.success('Buy order cancelled successfully');
      fetchOrders();
    } catch (error) {
      console.error('Error deleting buy order:', error);
      toast.error('Failed to cancel buy order');
    }
  };

  const handleSave = async () => {
    if (!formData.typeId || !formData.locationId || !formData.quantityDesired || formData.maxPricePerUnit === undefined) {
      toast.error('Please fill in all required fields');
      return;
    }

    try {
      const url = selectedOrder ? `/api/buy-orders/${selectedOrder.id}` : '/api/buy-orders';
      const method = selectedOrder ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          typeId: formData.typeId,
          locationId: formData.locationId,
          quantityDesired: formData.quantityDesired,
          minPricePerUnit: formData.minPricePerUnit || 0,
          maxPricePerUnit: formData.maxPricePerUnit,
          notes: formData.notes || null,
          ...(selectedOrder ? { isActive: true } : {}),
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to save buy order');
      }

      toast.success(selectedOrder ? 'Buy order updated successfully' : 'Buy order created successfully');
      setDialogOpen(false);
      fetchOrders();
    } catch (error: unknown) {
      console.error('Error saving buy order:', error);
      toast.error(error instanceof Error ? error.message : 'Failed to save buy order');
    }
  };

  const formatNumber = (num: number) => num.toLocaleString();
  const formatISK = (isk: number) => `${isk.toLocaleString()} ISK`;
  const formatDate = (dateString: string) => new Date(dateString).toLocaleDateString();

  if (loading) {
    return <Loading />;
  }

  return (
    <div className="max-w-[1280px] my-4">
      <Card className="bg-background-panel border-overlay-subtle">
        <CardContent className="p-6">
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-xl font-semibold text-text-emphasis">My Buy Orders</h2>
            <Button onClick={handleCreate}>
              <Plus className="h-4 w-4 mr-1" />
              Create Buy Order
            </Button>
          </div>

          {orders.length === 0 ? (
            <p className="text-text-secondary text-center py-8">
              No buy orders yet. Create one to let sellers know what you&apos;re looking for!
            </p>
          ) : (
            <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
              <Table>
                <TableHeader>
                  <TableRow className="bg-background-void">
                    <TableHead>Item</TableHead>
                    <TableHead>Location</TableHead>
                    <TableHead className="text-right">Quantity Desired</TableHead>
                    <TableHead className="text-right">Price Range/Unit</TableHead>
                    <TableHead className="text-right">Total Budget</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Notes</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {orders.map((order) => (
                    <TableRow key={order.id} className="bg-background-panel hover:bg-interactive-hover">
                      <TableCell className="text-text-emphasis">{order.typeName}</TableCell>
                      <TableCell className="text-text-secondary">{order.locationName || '-'}</TableCell>
                      <TableCell className="text-right text-text-emphasis">{formatNumber(order.quantityDesired)}</TableCell>
                      <TableCell className="text-right text-text-emphasis">
                        {order.minPricePerUnit > 0
                          ? `${formatISK(order.minPricePerUnit)} - ${formatISK(order.maxPricePerUnit)}`
                          : formatISK(order.maxPricePerUnit)}
                      </TableCell>
                      <TableCell className="text-right text-text-emphasis">
                        {formatISK(order.quantityDesired * order.maxPricePerUnit)}
                      </TableCell>
                      <TableCell>
                        <div className="flex gap-1 items-center">
                          <Badge
                            className={cn(
                              "text-xs font-semibold",
                              order.isActive
                                ? "bg-teal-success/15 text-teal-success border border-teal-success/30 hover:bg-teal-success/20"
                                : "bg-overlay-subtle text-text-muted border border-overlay-strong hover:bg-overlay-medium"
                            )}
                          >
                            {order.isActive ? 'Active' : 'Inactive'}
                          </Badge>
                          {order.autoBuyConfigId && (
                            <Badge className="text-[0.7rem] font-semibold bg-amber-manufacturing/15 text-amber-manufacturing border border-amber-manufacturing/30 hover:bg-amber-manufacturing/20 cursor-default">
                              Auto
                            </Badge>
                          )}
                        </div>
                      </TableCell>
                      <TableCell className="text-text-secondary">{order.notes || '-'}</TableCell>
                      <TableCell className="text-text-secondary">{formatDate(order.createdAt)}</TableCell>
                      <TableCell className="text-right">
                        {!order.autoBuyConfigId && (
                          <>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8 text-blue-science hover:text-blue-science hover:bg-blue-science/10"
                              onClick={() => handleEdit(order)}
                              title="Edit"
                            >
                              <Edit className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8 text-rose-danger hover:text-rose-danger hover:bg-rose-danger/10"
                              onClick={() => handleDelete(order.id)}
                              title="Cancel"
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create/Edit Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-sm bg-background-panel border-overlay-medium">
          <DialogHeader>
            <DialogTitle className="text-text-emphasis">
              {selectedOrder ? 'Edit Buy Order' : 'Create Buy Order'}
            </DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-4 mt-2">
            <div>
              <label className="text-sm text-text-secondary mb-1 block">Item Name *</label>
              <ItemSearchCombobox
                value={selectedItem}
                onChange={(item) => {
                  setSelectedItem(item);
                  if (item) {
                    setFormData({ ...formData, typeId: item.TypeID, typeName: item.TypeName });
                  } else {
                    setFormData({ ...formData, typeId: undefined, typeName: undefined });
                  }
                }}
                disabled={!!selectedOrder}
              />
              <p className="text-xs text-text-muted mt-1">
                {selectedOrder ? 'Cannot change item type' : 'Search for an item by name'}
              </p>
            </div>

            <div>
              <label className="text-sm text-text-secondary mb-1 block">Station *</label>
              <StationSearchCombobox
                value={selectedStation}
                onChange={(station) => {
                  setSelectedStation(station);
                  setFormData({ ...formData, locationId: station ? station.stationId : 0 });
                }}
              />
              <p className="text-xs text-text-muted mt-1">Where you want items delivered</p>
            </div>

            <div>
              <label className="text-sm text-text-secondary mb-1 block">Quantity Desired *</label>
              <Input
                type="number"
                value={formData.quantityDesired || ''}
                onChange={(e) => setFormData({ ...formData, quantityDesired: parseInt(e.target.value) })}
              />
            </div>

            <div className="flex gap-3">
              <div className="flex-1">
                <label className="text-sm text-text-secondary mb-1 block">Min Price Per Unit (ISK)</label>
                <Input
                  type="number"
                  value={formData.minPricePerUnit || ''}
                  onChange={(e) => setFormData({ ...formData, minPricePerUnit: parseFloat(e.target.value) || 0 })}
                />
                <p className="text-xs text-text-muted mt-1">Floor price for auto-fulfill (optional)</p>
              </div>
              <div className="flex-1">
                <label className="text-sm text-text-secondary mb-1 block">Max Price Per Unit (ISK) *</label>
                <Input
                  type="number"
                  value={formData.maxPricePerUnit || ''}
                  onChange={(e) => setFormData({ ...formData, maxPricePerUnit: parseFloat(e.target.value) || 0 })}
                />
                <p className="text-xs text-text-muted mt-1">Maximum you&apos;re willing to pay</p>
              </div>
            </div>

            <div>
              <label className="text-sm text-text-secondary mb-1 block">Notes</label>
              <textarea
                rows={3}
                className="w-full rounded-sm border border-overlay-strong bg-background-void text-text-emphasis text-sm px-3 py-2 resize-none focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary"
                value={formData.notes || ''}
                onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                placeholder="Optional notes about this buy order..."
              />
            </div>

            {formData.quantityDesired && formData.maxPricePerUnit !== undefined && (
              <p className="text-sm text-text-secondary">
                Total Budget: {formatISK(formData.quantityDesired * formData.maxPricePerUnit)}
              </p>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>Cancel</Button>
            <Button onClick={handleSave}>
              {selectedOrder ? 'Update' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
