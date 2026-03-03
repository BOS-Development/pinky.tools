import { useState, useMemo, useEffect, useRef } from 'react';
import { AssetsResponse, Asset, AssetContainer, AssetStructure, CorporationHanger, StockpileMarker } from "@industry-tool/client/data/models";
import AddStockpileDialog from './AddStockpileDialog';
import { useSession } from "next-auth/react";
import Navbar from "@industry-tool/components/Navbar";
import { formatISK, formatNumber, formatCompact } from '@industry-tool/utils/formatting';
import {
  Search,
  MapPin,
  Package,
  Layers,
  RefreshCw,
  DollarSign,
  AlertTriangle,
  ChevronUp,
  ChevronDown,
  EyeOff,
  Eye,
  Pin,
  Loader2,
  Plus,
  Pencil,
  Trash2,
  Tag,
  ShoppingCart,
  RefreshCcw,
  CheckCircle2,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent } from '@/components/ui/card';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { Separator } from '@/components/ui/separator';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';

// Combined sell + sync icon for auto-sell features
const AutoSellIcon = ({ className }: { className?: string }) => (
  <span className={cn('relative inline-flex items-center', className)}>
    <Tag className="h-4 w-4" />
    <RefreshCcw className="absolute -top-0.5 -right-0.5 h-2.5 w-2.5" />
  </span>
);

// Combined cart + sync icon for auto-buy features
const AutoBuyIcon = ({ className }: { className?: string }) => (
  <span className={cn('relative inline-flex items-center', className)}>
    <ShoppingCart className="h-4 w-4" />
    <RefreshCcw className="absolute -top-0.5 -right-0.5 h-2.5 w-2.5" />
  </span>
);

type ForSaleListing = {
  id: number;
  typeId: number;
  ownerId: number;
  locationId: number;
  containerId?: number;
  divisionNumber?: number;
  quantityAvailable: number;
  pricePerUnit: number;
  notes?: string;
  autoSellContainerId?: number;
};

type AutoSellConfig = {
  id: number;
  userId: number;
  ownerType: string;
  ownerId: number;
  locationId: number;
  containerId?: number;
  divisionNumber?: number;
  pricePercentage: number;
  priceSource: string;
  isActive: boolean;
};

type AutoBuyConfig = {
  id: number;
  userId: number;
  ownerType: string;
  ownerId: number;
  locationId: number;
  containerId?: number;
  divisionNumber?: number;
  minPricePercentage: number;
  maxPricePercentage: number;
  priceSource: string;
  isActive: boolean;
};

const PRICE_SOURCE_OPTIONS = [
  { value: 'jita_buy', label: 'Jita Buy (Best Bid)', abbrev: 'JBV' },
  { value: 'jita_sell', label: 'Jita Sell (Lowest Ask)', abbrev: 'JSV' },
  { value: 'jita_split', label: 'Jita Split (Buy+Sell avg)', abbrev: 'JSplit' },
] as const;

const getPriceSourceAbbrev = (source: string): string => {
  return PRICE_SOURCE_OPTIONS.find(o => o.value === source)?.abbrev || 'JBV';
};

const getPriceSourceLabel = (source: string): string => {
  return PRICE_SOURCE_OPTIONS.find(o => o.value === source)?.label || 'Jita Buy (Best Bid)';
};

function formatRelativeTime(date: Date): string {
  const now = new Date();
  const diffMs = date.getTime() - now.getTime();
  const absDiffMs = Math.abs(diffMs);
  const minutes = Math.round(absDiffMs / 60000);
  const hours = Math.round(absDiffMs / 3600000);

  if (minutes < 1) return diffMs < 0 ? 'just now' : 'in less than a minute';
  if (minutes < 60) return diffMs < 0 ? `${minutes} minute${minutes !== 1 ? 's' : ''} ago` : `in ${minutes} minute${minutes !== 1 ? 's' : ''}`;
  return diffMs < 0 ? `${hours} hour${hours !== 1 ? 's' : ''} ago` : `in ${hours} hour${hours !== 1 ? 's' : ''}`;
}

export type AssetsListProps = {
  assets?: AssetsResponse;
};

export default function AssetsList(props: AssetsListProps) {
  const { data: session } = useSession();
  const [assets, setAssets] = useState<AssetsResponse>(props.assets ?? { structures: [] });
  const [loading, setLoading] = useState(!props.assets);
  const [searchInput, setSearchInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(() => {
    // Load expanded nodes from localStorage on initial render
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('assetsList-expandedNodes');
      if (saved) {
        try {
          return new Set(JSON.parse(saved));
        } catch (e) {
          console.error('Failed to parse expanded nodes from localStorage', e);
        }
      }
    }
    return new Set();
  });
  const [hiddenStructures, setHiddenStructures] = useState<Set<number>>(() => {
    // Load hidden structures from localStorage on initial render
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('assetsList-hiddenStructures');
      if (saved) {
        try {
          return new Set(JSON.parse(saved));
        } catch (e) {
          console.error('Failed to parse hidden structures from localStorage', e);
        }
      }
    }
    return new Set();
  });
  const [pinnedStructures, setPinnedStructures] = useState<Set<number>>(() => {
    // Load pinned structures from localStorage on initial render
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('assetsList-pinnedStructures');
      if (saved) {
        try {
          return new Set(JSON.parse(saved));
        } catch (e) {
          console.error('Failed to parse pinned structures from localStorage', e);
        }
      }
    }
    return new Set();
  });
  const [showBelowTargetOnly, setShowBelowTargetOnly] = useState(false);
  const [stockpileModalOpen, setStockpileModalOpen] = useState(false);
  const [selectedAsset, setSelectedAsset] = useState<{
    asset: Asset;
    locationId: number;
    containerId?: number;
    divisionNumber?: number;
  } | null>(null);
  const [desiredQuantity, setDesiredQuantity] = useState('');
  const [notes, setNotes] = useState('');
  const [autoProductionEnabled, setAutoProductionEnabled] = useState(false);
  const [selectedPlanId, setSelectedPlanId] = useState<number | null>(null);
  const [availablePlans, setAvailablePlans] = useState<{id: number; name: string; productName?: string}[]>([]);
  const [plansLoading, setPlansLoading] = useState(false);
  const [parallelism, setParallelism] = useState(0);
  const [addStockpileDialogOpen, setAddStockpileDialogOpen] = useState(false);
  const [addStockpileContext, setAddStockpileContext] = useState<{
    locationId: number;
    containerId?: number;
    divisionNumber?: number;
    owners: { ownerType: string; ownerId: number; ownerName: string }[];
  } | null>(null);
  const desiredQuantityInputRef = useRef<HTMLInputElement>(null);
  const [refreshingPrices, setRefreshingPrices] = useState(false);
  const [assetStatus, setAssetStatus] = useState<{ lastUpdatedAt: string | null; nextUpdateAt: string | null } | null>(null);

  // For-sale listing state
  const [listingDialogOpen, setListingDialogOpen] = useState(false);
  const [listingAsset, setListingAsset] = useState<{
    asset: Asset;
    locationId: number;
    containerId?: number;
    divisionNumber?: number;
  } | null>(null);
  const [submittingListing, setSubmittingListing] = useState(false);
  const [editingListingId, setEditingListingId] = useState<number | null>(null);
  const listingQuantityRef = useRef<HTMLInputElement>(null);
  const listingPriceRef = useRef<HTMLInputElement>(null);
  const listingNotesRef = useRef<HTMLInputElement>(null);
  const [listingTotalValue, setListingTotalValue] = useState<string>('');

  // Track active for-sale listings
  const [forSaleListings, setForSaleListings] = useState<ForSaleListing[]>([]);

  // Auto-sell state
  const [autoSellConfigs, setAutoSellConfigs] = useState<AutoSellConfig[]>([]);
  const [autoSellDialogOpen, setAutoSellDialogOpen] = useState(false);
  const [autoSellContainer, setAutoSellContainer] = useState<{
    ownerType: string;
    ownerId: number;
    locationId: number;
    containerId?: number;
    divisionNumber?: number;
    containerName: string;
  } | null>(null);
  const [autoSellPercentage, setAutoSellPercentage] = useState('90');
  const [autoSellPriceSource, setAutoSellPriceSource] = useState('jita_buy');
  const [submittingAutoSell, setSubmittingAutoSell] = useState(false);

  // Auto-buy state
  const [autoBuyConfigs, setAutoBuyConfigs] = useState<AutoBuyConfig[]>([]);
  const [autoBuyDialogOpen, setAutoBuyDialogOpen] = useState(false);
  const [autoBuyContainer, setAutoBuyContainer] = useState<{
    ownerType: string;
    ownerId: number;
    locationId: number;
    containerId?: number;
    divisionNumber?: number;
    containerName: string;
  } | null>(null);
  const [autoBuyMinPercentage, setAutoBuyMinPercentage] = useState('0');
  const [autoBuyMaxPercentage, setAutoBuyMaxPercentage] = useState('100');
  const [autoBuyPriceSource, setAutoBuyPriceSource] = useState('jita_sell');
  const [submittingAutoBuy, setSubmittingAutoBuy] = useState(false);

  useEffect(() => {
    if (!selectedAsset || !stockpileModalOpen) {
      return;
    }
    const fetchPlans = async () => {
      setPlansLoading(true);
      try {
        const res = await fetch(`/api/industry/plans/by-product/${selectedAsset.asset.typeId}`);
        if (res.ok) {
          const data = await res.json();
          setAvailablePlans(data || []);
        }
      } finally {
        setPlansLoading(false);
      }
    };
    fetchPlans();
  }, [selectedAsset, stockpileModalOpen]);

  const handleQuantityChange = (value: string) => {
    // Remove all non-digit characters
    const numericValue = value.replace(/\D/g, '');
    // Format with commas
    const formatted = numericValue ? parseInt(numericValue).toLocaleString() : '';
    setDesiredQuantity(formatted);
  };

  const refetchAssets = async () => {
    if (!session) return;

    setLoading(true);
    try {
      const response = await fetch('/api/assets/get');
      if (response.ok) {
        const data = await response.json();
        setAssets(data);
      }
    } finally {
      setLoading(false);
    }
  };

  const fetchAssetStatus = async () => {
    if (!session) return;

    try {
      const response = await fetch('/api/assets/status');
      if (response.ok) {
        const data = await response.json();
        setAssetStatus(data);
      }
    } catch (error) {
      console.error('Failed to fetch asset status:', error);
    }
  };

  const fetchForSaleListings = async () => {
    if (!session) return;

    try {
      const response = await fetch('/api/for-sale');
      if (response.ok) {
        const data = await response.json();
        setForSaleListings(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch for-sale listings:', error);
    }
  };

  const fetchAutoSellConfigs = async () => {
    if (!session) return;

    try {
      const response = await fetch('/api/auto-sell');
      if (response.ok) {
        const data = await response.json();
        setAutoSellConfigs(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch auto-sell configs:', error);
    }
  };

  const getAutoSellForContainer = (containerId: number | undefined, ownerType: string, ownerId: number, locationId: number, divisionNumber?: number): AutoSellConfig | undefined => {
    return autoSellConfigs.find(config =>
      (config.containerId || 0) === (containerId || 0) &&
      config.ownerType === ownerType &&
      config.ownerId === ownerId &&
      config.locationId === locationId &&
      (config.divisionNumber || 0) === (divisionNumber || 0)
    );
  };

  const handleOpenAutoSellDialog = (ownerType: string, ownerId: number, locationId: number, containerId: number | undefined, containerName: string, divisionNumber?: number) => {
    const existing = getAutoSellForContainer(containerId, ownerType, ownerId, locationId, divisionNumber);
    setAutoSellContainer({ ownerType, ownerId, locationId, containerId, divisionNumber, containerName });
    setAutoSellPercentage(existing ? existing.pricePercentage.toString() : '90');
    setAutoSellPriceSource(existing ? existing.priceSource : 'jita_buy');
    setAutoSellDialogOpen(true);
  };

  const handleSaveAutoSell = async () => {
    if (!autoSellContainer || !session) return;

    setSubmittingAutoSell(true);
    try {
      const response = await fetch('/api/auto-sell', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ownerType: autoSellContainer.ownerType,
          ownerId: autoSellContainer.ownerId,
          locationId: autoSellContainer.locationId,
          containerId: autoSellContainer.containerId,
          divisionNumber: autoSellContainer.divisionNumber,
          pricePercentage: parseFloat(autoSellPercentage),
          priceSource: autoSellPriceSource,
        }),
      });

      if (response.ok) {
        setAutoSellDialogOpen(false);
        await fetchAutoSellConfigs();
        await fetchForSaleListings();
      }
    } finally {
      setSubmittingAutoSell(false);
    }
  };

  const handleDisableAutoSell = async () => {
    if (!autoSellContainer || !session) return;

    const existing = getAutoSellForContainer(
      autoSellContainer.containerId,
      autoSellContainer.ownerType,
      autoSellContainer.ownerId,
      autoSellContainer.locationId,
      autoSellContainer.divisionNumber
    );
    if (!existing) return;

    setSubmittingAutoSell(true);
    try {
      const response = await fetch(`/api/auto-sell/${existing.id}`, {
        method: 'DELETE',
      });

      if (response.ok) {
        setAutoSellDialogOpen(false);
        await fetchAutoSellConfigs();
        await fetchForSaleListings();
      }
    } finally {
      setSubmittingAutoSell(false);
    }
  };

  const fetchAutoBuyConfigs = async () => {
    if (!session) return;

    try {
      const response = await fetch('/api/auto-buy');
      if (response.ok) {
        const data = await response.json();
        setAutoBuyConfigs(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch auto-buy configs:', error);
    }
  };

  const getAutoBuyForContainer = (containerId: number | undefined, ownerType: string, ownerId: number, locationId: number, divisionNumber?: number): AutoBuyConfig | undefined => {
    return autoBuyConfigs.find(config =>
      (config.containerId || 0) === (containerId || 0) &&
      config.ownerType === ownerType &&
      config.ownerId === ownerId &&
      config.locationId === locationId &&
      (config.divisionNumber || 0) === (divisionNumber || 0)
    );
  };

  const handleOpenAutoBuyDialog = (ownerType: string, ownerId: number, locationId: number, containerId: number | undefined, containerName: string, divisionNumber?: number) => {
    const existing = getAutoBuyForContainer(containerId, ownerType, ownerId, locationId, divisionNumber);
    setAutoBuyContainer({ ownerType, ownerId, locationId, containerId, divisionNumber, containerName });
    setAutoBuyMinPercentage(existing ? existing.minPricePercentage.toString() : '0');
    setAutoBuyMaxPercentage(existing ? existing.maxPricePercentage.toString() : '100');
    setAutoBuyPriceSource(existing ? existing.priceSource : 'jita_sell');
    setAutoBuyDialogOpen(true);
  };

  const handleSaveAutoBuy = async () => {
    if (!autoBuyContainer || !session) return;

    setSubmittingAutoBuy(true);
    try {
      const existing = getAutoBuyForContainer(
        autoBuyContainer.containerId,
        autoBuyContainer.ownerType,
        autoBuyContainer.ownerId,
        autoBuyContainer.locationId,
        autoBuyContainer.divisionNumber
      );

      if (existing) {
        const response = await fetch(`/api/auto-buy/${existing.id}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            minPricePercentage: parseFloat(autoBuyMinPercentage),
            maxPricePercentage: parseFloat(autoBuyMaxPercentage),
            priceSource: autoBuyPriceSource,
          }),
        });
        if (response.ok) {
          setAutoBuyDialogOpen(false);
          await fetchAutoBuyConfigs();
        }
      } else {
        const response = await fetch('/api/auto-buy', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            ownerType: autoBuyContainer.ownerType,
            ownerId: autoBuyContainer.ownerId,
            locationId: autoBuyContainer.locationId,
            containerId: autoBuyContainer.containerId,
            divisionNumber: autoBuyContainer.divisionNumber,
            minPricePercentage: parseFloat(autoBuyMinPercentage),
            maxPricePercentage: parseFloat(autoBuyMaxPercentage),
            priceSource: autoBuyPriceSource,
          }),
        });
        if (response.ok) {
          setAutoBuyDialogOpen(false);
          await fetchAutoBuyConfigs();
        }
      }
    } finally {
      setSubmittingAutoBuy(false);
    }
  };

  const handleDisableAutoBuy = async () => {
    if (!autoBuyContainer || !session) return;

    const existing = getAutoBuyForContainer(
      autoBuyContainer.containerId,
      autoBuyContainer.ownerType,
      autoBuyContainer.ownerId,
      autoBuyContainer.locationId,
      autoBuyContainer.divisionNumber
    );
    if (!existing) return;

    setSubmittingAutoBuy(true);
    try {
      const response = await fetch(`/api/auto-buy/${existing.id}`, {
        method: 'DELETE',
      });

      if (response.ok) {
        setAutoBuyDialogOpen(false);
        await fetchAutoBuyConfigs();
      }
    } finally {
      setSubmittingAutoBuy(false);
    }
  };

  const getListingForAsset = (asset: Asset, locationId: number, containerId?: number, divisionNumber?: number): ForSaleListing | undefined => {
    return forSaleListings.find(listing =>
      listing.typeId === asset.typeId &&
      listing.ownerId === asset.ownerId &&
      listing.locationId === locationId &&
      (listing.containerId || 0) === (containerId || 0) &&
      (listing.divisionNumber || 0) === (divisionNumber || 0)
    );
  };

  const handleRefreshPrices = async () => {
    if (!session) return;

    setRefreshingPrices(true);
    try {
      await fetch('/api/market-prices/update', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });
      // Refetch assets to show updated prices
      await refetchAssets();
    } finally {
      setRefreshingPrices(false);
    }
  };

  // Load assets on mount if not provided via props
  useEffect(() => {
    if (!props.assets && session) {
      refetchAssets();
    }

    if (session) {
      fetchForSaleListings();
      fetchAssetStatus();
      fetchAutoSellConfigs();
      fetchAutoBuyConfigs();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      setSearchQuery(searchInput);
    }, 300);

    return () => clearTimeout(timer);
  }, [searchInput]);

  const toggleNode = (nodeId: string) => {
    setExpandedNodes((prev) => {
      const next = new Set(prev);
      if (next.has(nodeId)) {
        next.delete(nodeId);
      } else {
        next.add(nodeId);
      }
      return next;
    });
  };

  const toggleHideStructure = (structureId: number) => {
    setHiddenStructures((prev) => {
      const next = new Set(prev);
      if (next.has(structureId)) {
        next.delete(structureId);
      } else {
        next.add(structureId);
      }
      return next;
    });
  };

  const togglePinStructure = (structureId: number) => {
    setPinnedStructures((prev) => {
      const next = new Set(prev);
      if (next.has(structureId)) {
        next.delete(structureId);
      } else {
        next.add(structureId);
      }
      return next;
    });
  };

  // Save hidden structures to localStorage whenever it changes
  useEffect(() => {
    if (typeof window !== 'undefined') {
      localStorage.setItem('assetsList-hiddenStructures', JSON.stringify(Array.from(hiddenStructures)));
    }
  }, [hiddenStructures]);

  // Save pinned structures to localStorage whenever it changes
  useEffect(() => {
    if (typeof window !== 'undefined') {
      localStorage.setItem('assetsList-pinnedStructures', JSON.stringify(Array.from(pinnedStructures)));
    }
  }, [pinnedStructures]);

  // Save expanded nodes to localStorage whenever it changes
  useEffect(() => {
    if (typeof window !== 'undefined') {
      localStorage.setItem('assetsList-expandedNodes', JSON.stringify(Array.from(expandedNodes)));
    }
  }, [expandedNodes]);

  const { totalItems, totalVolume, uniqueTypes, totalValue, totalDeficit, filteredStructures } = useMemo(() => {
    // Return empty values if no assets
    if (!assets?.structures || assets.structures.length === 0) {
      return {
        totalItems: 0,
        totalVolume: 0,
        uniqueTypes: 0,
        totalValue: 0,
        totalDeficit: 0,
        filteredStructures: []
      };
    }

    let items = 0;
    let volume = 0;
    let value = 0;
    let deficit = 0;
    const types = new Set<string>();

    const countAssets = (assets: Asset[]) => {
      assets.forEach((asset) => {
        items += 1;
        volume += asset.volume;
        types.add(asset.name);
        if (asset.totalValue) value += asset.totalValue;
        if (asset.deficitValue) deficit += asset.deficitValue;
      });
    };

    const filtered = assets.structures.map((structure) => {
      const filteredStructure = { ...structure };

      if (searchQuery) {
        const query = searchQuery.toLowerCase();

        // Check if the query matches the structure/location name
        const structureMatches = structure.name.toLowerCase().includes(query) ||
                                 structure.solarSystem.toLowerCase().includes(query) ||
                                 structure.region.toLowerCase().includes(query);

        // If structure matches, show all assets; otherwise filter by asset names
        if (structureMatches) {
          // Keep all assets when structure name matches
          filteredStructure.hangarAssets = structure.hangarAssets || [];
          filteredStructure.hangarContainers = structure.hangarContainers || [];
          filteredStructure.deliveries = structure.deliveries || [];
          filteredStructure.assetSafety = structure.assetSafety || [];
          filteredStructure.corporationHangers = structure.corporationHangers || [];
        } else {
          // Filter by asset names only
          filteredStructure.hangarAssets = structure.hangarAssets?.filter(a => a.name.toLowerCase().includes(query)) || [];
          filteredStructure.hangarContainers = structure.hangarContainers?.map(c => ({
            ...c,
            assets: c.assets.filter(a => a.name.toLowerCase().includes(query))
          })).filter(c => c.assets.length > 0) || [];
          filteredStructure.deliveries = structure.deliveries?.filter(a => a.name.toLowerCase().includes(query)) || [];
          filteredStructure.assetSafety = structure.assetSafety?.filter(a => a.name.toLowerCase().includes(query)) || [];
          filteredStructure.corporationHangers = structure.corporationHangers?.map(h => ({
            ...h,
            assets: h.assets.filter(a => a.name.toLowerCase().includes(query)),
            hangarContainers: h.hangarContainers?.map(c => ({
              ...c,
              assets: c.assets.filter(a => a.name.toLowerCase().includes(query))
            })).filter(c => c.assets.length > 0) || []
          })).filter(h => h.assets.length > 0 || h.hangarContainers.length > 0) || [];
        }
      }

      if (showBelowTargetOnly) {
        filteredStructure.hangarAssets = filteredStructure.hangarAssets?.filter(a => a.stockpileDelta && a.stockpileDelta < 0) || [];
        filteredStructure.hangarContainers = filteredStructure.hangarContainers?.map(c => ({
          ...c,
          assets: c.assets.filter(a => a.stockpileDelta && a.stockpileDelta < 0)
        })).filter(c => c.assets.length > 0) || [];
        filteredStructure.deliveries = filteredStructure.deliveries?.filter(a => a.stockpileDelta && a.stockpileDelta < 0) || [];
        filteredStructure.assetSafety = filteredStructure.assetSafety?.filter(a => a.stockpileDelta && a.stockpileDelta < 0) || [];
        filteredStructure.corporationHangers = filteredStructure.corporationHangers?.map(h => ({
          ...h,
          assets: h.assets.filter(a => a.stockpileDelta && a.stockpileDelta < 0),
          hangarContainers: h.hangarContainers?.map(c => ({
            ...c,
            assets: c.assets.filter(a => a.stockpileDelta && a.stockpileDelta < 0)
          })).filter(c => c.assets.length > 0) || []
        })).filter(h => h.assets.length > 0 || h.hangarContainers.length > 0) || [];
      }

      return filteredStructure;
    }).filter(s =>
      s.hangarAssets?.length > 0 ||
      s.hangarContainers?.length > 0 ||
      s.deliveries?.length > 0 ||
      s.assetSafety?.length > 0 ||
      s.corporationHangers?.length > 0
    );

    // Count totals from original (unfiltered) data
    assets.structures.forEach((structure) => {
      if (structure.hangarAssets) countAssets(structure.hangarAssets);
      if (structure.deliveries) countAssets(structure.deliveries);
      if (structure.assetSafety) countAssets(structure.assetSafety);
      structure.hangarContainers?.forEach((c) => countAssets(c.assets));
      structure.corporationHangers?.forEach((h) => {
        countAssets(h.assets);
        h.hangarContainers?.forEach((c) => countAssets(c.assets));
      });
    });

    return {
      totalItems: items,
      totalVolume: volume,
      uniqueTypes: types.size,
      totalValue: value,
      totalDeficit: deficit,
      filteredStructures: filtered
    };
  }, [assets, searchQuery, showBelowTargetOnly]);

  // Split structures into visible and hidden, sort pinned to top
  const { visibleStructures, hiddenStructuresList } = useMemo(() => {
    const visible = filteredStructures
      .filter(s => !hiddenStructures.has(s.id))
      .sort((a, b) => {
        const aIsPinned = pinnedStructures.has(a.id);
        const bIsPinned = pinnedStructures.has(b.id);
        if (aIsPinned && !bIsPinned) return -1;
        if (!aIsPinned && bIsPinned) return 1;
        return 0;
      });
    const hidden = filteredStructures.filter(s => hiddenStructures.has(s.id));
    return { visibleStructures: visible, hiddenStructuresList: hidden };
  }, [filteredStructures, hiddenStructures, pinnedStructures]);

  // Auto-expand nodes when searching
  useEffect(() => {
    if (!searchQuery) {
      return;
    }

    const nodesToExpand = new Set<string>();
    const query = searchQuery.toLowerCase();

    filteredStructures.forEach((structure) => {
      const structureId = `structure-${structure.id}`;

      // Check if the structure name/location matches (if so, don't auto-expand)
      const structureMatches = structure.name.toLowerCase().includes(query) ||
                               structure.solarSystem.toLowerCase().includes(query) ||
                               structure.region.toLowerCase().includes(query);

      // Only auto-expand if structure doesn't match (meaning items matched instead)
      if (!structureMatches) {
        // Expand structure if it has any results
        if (structure.hangarAssets?.length > 0 ||
            structure.hangarContainers?.length > 0 ||
            structure.deliveries?.length > 0 ||
            structure.assetSafety?.length > 0 ||
            structure.corporationHangers?.length > 0) {
          nodesToExpand.add(structureId);
        }

        // Expand personal hangar if it has results
        if (structure.hangarAssets?.length > 0) {
          nodesToExpand.add(`structure-${structure.id}-hangar`);
        }

        // Expand containers with results
        structure.hangarContainers?.forEach((container) => {
          if (container.assets.length > 0) {
            nodesToExpand.add(`structure-${structure.id}-container-${container.id}`);
          }
        });

        // Expand deliveries if it has results
        if (structure.deliveries?.length > 0) {
          nodesToExpand.add(`structure-${structure.id}-deliveries`);
        }

        // Expand asset safety if it has results
        if (structure.assetSafety?.length > 0) {
          nodesToExpand.add(`structure-${structure.id}-safety`);
        }

        // Expand corporation hangars and their containers
        structure.corporationHangers?.forEach((hanger) => {
          const hangerNodeId = `structure-${structure.id}-corp-${hanger.id}`;
          if (hanger.assets.length > 0 || hanger.hangarContainers?.length > 0) {
            nodesToExpand.add(hangerNodeId);
          }

          hanger.hangarContainers?.forEach((container) => {
            if (container.assets.length > 0) {
              nodesToExpand.add(`${hangerNodeId}-container-${container.id}`);
            }
          });
        });
      }
    });

    setExpandedNodes(nodesToExpand);
  }, [searchQuery]); // eslint-disable-line react-hooks/exhaustive-deps

  const handleOpenStockpileModal = (asset: Asset, locationId: number, containerId?: number, divisionNumber?: number) => {
    setSelectedAsset({ asset, locationId, containerId, divisionNumber });
    setDesiredQuantity(asset.desiredQuantity?.toLocaleString() || '');
    setNotes('');
    // Reset auto-production state (will be populated by marker fetch)
    setAutoProductionEnabled(false);
    setSelectedPlanId(null);
    setParallelism(0);

    // Fetch existing marker data to get auto-production settings
    if (asset.desiredQuantity) {
      fetch('/api/stockpiles')
        .then(res => res.ok ? res.json() : [])
        .then((markers: StockpileMarker[]) => {
          const marker = markers.find(m =>
            m.typeId === asset.typeId &&
            m.ownerId === asset.ownerId &&
            m.locationId === locationId &&
            (containerId ? m.containerId === containerId : !m.containerId) &&
            (divisionNumber ? m.divisionNumber === divisionNumber : !m.divisionNumber)
          );
          if (marker) {
            setAutoProductionEnabled(marker.autoProductionEnabled || false);
            setSelectedPlanId(marker.planId || null);
            setParallelism(marker.autoProductionParallelism || 0);
          }
        })
        .catch(() => {});
    }

    setStockpileModalOpen(true);
  };

  const handleSaveStockpile = async () => {
    if (!selectedAsset || !session) return;

    const desiredQty = parseInt(desiredQuantity.replace(/,/g, ''));
    const marker: StockpileMarker = {
      userId: 0, // Will be set by backend
      typeId: selectedAsset.asset.typeId,
      ownerType: selectedAsset.asset.ownerType,
      ownerId: selectedAsset.asset.ownerId,
      locationId: selectedAsset.locationId,
      containerId: selectedAsset.containerId,
      divisionNumber: selectedAsset.divisionNumber,
      desiredQuantity: desiredQty,
      notes: notes || undefined,
      autoProductionEnabled,
      planId: autoProductionEnabled && selectedPlanId ? selectedPlanId : undefined,
      autoProductionParallelism: autoProductionEnabled ? parallelism : undefined,
    };

    await fetch('/api/stockpiles/upsert', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(marker),
    });

    // Update local state instead of refetching
    setAssets(prev => {
      const updated = { ...prev };
      // Find and update the asset in the nested structure
      for (const structure of updated.structures) {
        // Update hangar assets
        if (structure.hangarAssets) {
          const asset = structure.hangarAssets.find(a =>
            a.typeId === selectedAsset.asset.typeId &&
            a.ownerId === selectedAsset.asset.ownerId
          );
          if (asset) {
            asset.desiredQuantity = desiredQty;
            asset.stockpileDelta = asset.quantity - desiredQty;
            asset.deficitValue = asset.stockpileDelta < 0
              ? Math.abs(asset.stockpileDelta) * (asset.unitPrice || 0)
              : 0;
          }
        }
        // Update hangar containers
        if (structure.hangarContainers) {
          for (const container of structure.hangarContainers) {
            const asset = container.assets.find(a =>
              a.typeId === selectedAsset.asset.typeId &&
              a.ownerId === selectedAsset.asset.ownerId
            );
            if (asset) {
              asset.desiredQuantity = desiredQty;
              asset.stockpileDelta = asset.quantity - desiredQty;
              asset.deficitValue = asset.stockpileDelta < 0
                ? Math.abs(asset.stockpileDelta) * (asset.unitPrice || 0)
                : 0;
            }
          }
        }
        // Update corporation hangers
        if (structure.corporationHangers) {
          for (const hanger of structure.corporationHangers) {
            const asset = hanger.assets.find(a =>
              a.typeId === selectedAsset.asset.typeId &&
              a.ownerId === selectedAsset.asset.ownerId
            );
            if (asset) {
              asset.desiredQuantity = desiredQty;
              asset.stockpileDelta = asset.quantity - desiredQty;
              asset.deficitValue = asset.stockpileDelta < 0
                ? Math.abs(asset.stockpileDelta) * (asset.unitPrice || 0)
                : 0;
            }
            // Check hanger containers
            if (hanger.hangarContainers) {
              for (const container of hanger.hangarContainers) {
                const asset = container.assets.find(a =>
                  a.typeId === selectedAsset.asset.typeId &&
                  a.ownerId === selectedAsset.asset.ownerId
                );
                if (asset) {
                  asset.desiredQuantity = desiredQty;
                  asset.stockpileDelta = asset.quantity - desiredQty;
                  asset.deficitValue = asset.stockpileDelta < 0
                    ? Math.abs(asset.stockpileDelta) * (asset.unitPrice || 0)
                    : 0;
                }
              }
            }
          }
        }
      }
      return updated;
    });

    setStockpileModalOpen(false);
  };

  const handleDeleteStockpile = async (asset: Asset, locationId: number, containerId?: number, divisionNumber?: number) => {
    if (!confirm('Remove stockpile marker?') || !session) return;

    const marker: StockpileMarker = {
      userId: 0,
      typeId: asset.typeId,
      ownerType: asset.ownerType,
      ownerId: asset.ownerId,
      locationId: locationId,
      containerId: containerId,
      divisionNumber: divisionNumber,
      desiredQuantity: 0,
    };

    await fetch('/api/stockpiles/delete', {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(marker),
    });

    // Update local state instead of refetching
    setAssets(prev => {
      const updated = { ...prev };
      // Find and update the asset in the nested structure
      for (const structure of updated.structures) {
        // Update hangar assets
        if (structure.hangarAssets) {
          const foundAsset = structure.hangarAssets.find(a =>
            a.typeId === asset.typeId &&
            a.ownerId === asset.ownerId
          );
          if (foundAsset) {
            foundAsset.desiredQuantity = undefined;
            foundAsset.stockpileDelta = undefined;
            foundAsset.deficitValue = undefined;
          }
        }
        // Update hangar containers
        if (structure.hangarContainers) {
          for (const container of structure.hangarContainers) {
            const foundAsset = container.assets.find(a =>
              a.typeId === asset.typeId &&
              a.ownerId === asset.ownerId
            );
            if (foundAsset) {
              foundAsset.desiredQuantity = undefined;
              foundAsset.stockpileDelta = undefined;
              foundAsset.deficitValue = undefined;
            }
          }
        }
        // Update corporation hangers
        if (structure.corporationHangers) {
          for (const hanger of structure.corporationHangers) {
            const foundAsset = hanger.assets.find(a =>
              a.typeId === asset.typeId &&
              a.ownerId === asset.ownerId
            );
            if (foundAsset) {
              foundAsset.desiredQuantity = undefined;
              foundAsset.stockpileDelta = undefined;
              foundAsset.deficitValue = undefined;
            }
            // Check hanger containers
            if (hanger.hangarContainers) {
              for (const container of hanger.hangarContainers) {
                const foundAsset = container.assets.find(a =>
                  a.typeId === asset.typeId &&
                  a.ownerId === asset.ownerId
                );
                if (foundAsset) {
                  foundAsset.desiredQuantity = undefined;
                  foundAsset.stockpileDelta = undefined;
                  foundAsset.deficitValue = undefined;
                }
              }
            }
          }
        }
      }
      return updated;
    });
  };

  const handleOpenAddStockpileDialog = (locationId: number, owners: { ownerType: string; ownerId: number; ownerName: string }[], containerId?: number, divisionNumber?: number) => {
    setAddStockpileContext({ locationId, containerId, divisionNumber, owners });
    setAddStockpileDialogOpen(true);
  };

  const handleAddStockpileSaved = (phantomAsset: Asset) => {
    if (!addStockpileContext) return;
    const { locationId, containerId, divisionNumber } = addStockpileContext;

    // Update existing asset row if it matches, otherwise append phantom row
    const upsertAsset = (assets: Asset[]): Asset[] => {
      const idx = assets.findIndex(a => a.typeId === phantomAsset.typeId && a.ownerId === phantomAsset.ownerId);
      if (idx >= 0) {
        const copy = [...assets];
        copy[idx] = {
          ...copy[idx],
          desiredQuantity: phantomAsset.desiredQuantity,
          stockpileDelta: copy[idx].quantity - (phantomAsset.desiredQuantity ?? 0),
        };
        return copy;
      }
      return [...assets, phantomAsset];
    };

    setAssets(prev => ({
      ...prev,
      structures: prev.structures.map(structure => {
        if (structure.id !== locationId) return structure;

        if (containerId) {
          if (divisionNumber) {
            return {
              ...structure,
              corporationHangers: structure.corporationHangers?.map(h =>
                h.id === divisionNumber
                  ? { ...h, hangarContainers: h.hangarContainers?.map(c =>
                      c.id === containerId ? { ...c, assets: upsertAsset(c.assets) } : c
                    )}
                  : h
              ),
            };
          }
          return {
            ...structure,
            hangarContainers: structure.hangarContainers?.map(c =>
              c.id === containerId ? { ...c, assets: upsertAsset(c.assets) } : c
            ),
          };
        } else if (divisionNumber) {
          return {
            ...structure,
            corporationHangers: structure.corporationHangers?.map(h =>
              h.id === divisionNumber ? { ...h, assets: upsertAsset(h.assets) } : h
            ),
          };
        } else {
          return {
            ...structure,
            hangarAssets: upsertAsset(structure.hangarAssets || []),
          };
        }
      }),
    }));
  };

  const handleOpenListingDialog = (asset: Asset, locationId: number, containerId?: number, divisionNumber?: number, existingListing?: ForSaleListing) => {
    setListingAsset({ asset, locationId, containerId, divisionNumber });

    if (existingListing) {
      // Editing existing listing
      setEditingListingId(existingListing.id);
      // Set all fields via refs after dialog opens
      setTimeout(() => {
        if (listingQuantityRef.current) {
          listingQuantityRef.current.value = existingListing.quantityAvailable.toLocaleString();
        }
        if (listingPriceRef.current) {
          listingPriceRef.current.value = existingListing.pricePerUnit.toLocaleString();
        }
        if (listingNotesRef.current) {
          listingNotesRef.current.value = existingListing.notes || '';
        }
        updateTotalValue();
      }, 0);
    } else {
      // Creating new listing
      setEditingListingId(null);
      // Set initial values via refs after dialog opens
      setTimeout(() => {
        if (listingQuantityRef.current) {
          listingQuantityRef.current.value = asset.quantity.toLocaleString();
        }
        if (listingPriceRef.current) {
          listingPriceRef.current.value = '';
        }
        if (listingNotesRef.current) {
          listingNotesRef.current.value = '';
        }
        updateTotalValue();
      }, 0);
    }

    setListingDialogOpen(true);
  };

  const updateTotalValue = () => {
    const quantity = listingQuantityRef.current?.value.replace(/,/g, '') || '0';
    const price = listingPriceRef.current?.value.replace(/,/g, '') || '0';
    const quantityNum = parseInt(quantity) || 0;
    const priceNum = parseFloat(price) || 0;
    const total = quantityNum * priceNum;
    setListingTotalValue(total > 0 ? total.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 2 }) : '');
  };

  const handleListingInputChange = (ref: React.RefObject<HTMLInputElement | null>, allowDecimals = false) => {
    if (!ref.current) return;
    const stripped = ref.current.value.replace(/,/g, '');
    if (allowDecimals) {
      const num = parseFloat(stripped);
      if (isNaN(num) || num <= 0) {
        ref.current.value = '';
      } else {
        ref.current.value = num.toLocaleString(undefined, {
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
        });
      }
    } else {
      const numericValue = stripped.replace(/\D/g, '');
      const formatted = numericValue ? parseInt(numericValue).toLocaleString() : '';
      ref.current.value = formatted;
    }
    updateTotalValue();
  };

  const handleCreateListing = async () => {
    if (!listingAsset || !session) return;

    const quantityValue = listingQuantityRef.current?.value.replace(/,/g, '') || '0';
    const priceValue = listingPriceRef.current?.value.replace(/,/g, '') || '0';

    const quantity = parseInt(quantityValue);
    const price = parseFloat(priceValue);
    const notes = listingNotesRef.current?.value || '';

    if (!quantity || !price) return;

    setSubmittingListing(true);
    try {
      const url = editingListingId ? `/api/for-sale/${editingListingId}` : '/api/for-sale';
      const method = editingListingId ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          typeId: listingAsset.asset.typeId,
          ownerType: listingAsset.asset.ownerType,
          ownerId: listingAsset.asset.ownerId,
          locationId: listingAsset.locationId,
          containerId: listingAsset.containerId,
          divisionNumber: listingAsset.divisionNumber,
          quantityAvailable: quantity,
          pricePerUnit: price,
          notes: notes || undefined,
        }),
      });

      if (response.ok) {
        setListingDialogOpen(false);
        // Refresh listings to show the updated listing indicator
        await fetchForSaleListings();
      }
    } finally {
      setSubmittingListing(false);
    }
  };

  const handleDeleteListing = async () => {
    if (!editingListingId || !session) return;

    if (!confirm('Are you sure you want to delete this listing?')) return;

    setSubmittingListing(true);
    try {
      const response = await fetch(`/api/for-sale/${editingListingId}`, {
        method: 'DELETE',
      });

      if (response.ok) {
        setListingDialogOpen(false);
        // Refresh listings to remove the deleted listing indicator
        await fetchForSaleListings();
      }
    } finally {
      setSubmittingListing(false);
    }
  };

  // Show loading state first, before checking if assets are empty
  if (loading) {
    return (
      <>
        <Navbar />
        <div className="mt-2 mb-2 px-6">
          <div className="flex justify-center items-center min-h-[60vh]">
            <p className="text-lg text-[#94a3b8]">Loading assets...</p>
          </div>
        </div>
      </>
    );
  }

  if (!assets?.structures || assets.structures.length === 0) {
    return (
      <>
        <Navbar />
        <div className="max-w-screen-xl mx-auto px-4 mt-4">
          <div className="flex flex-col items-center justify-center min-h-[60vh] text-center">
            <h2 className="text-3xl font-semibold text-[#e2e8f0] mb-2">No Assets Found</h2>
            <p className="text-sm text-[#94a3b8]">
              You don&apos;t have any assets yet, or they haven&apos;t been synced.
            </p>
          </div>
        </div>
      </>
    );
  }

  const renderAssetsTable = (assetsToRender: Asset[], showOwner: boolean, locationId: number, containerId?: number, divisionNumber?: number) => (
    <div className="px-2 pb-1">
      <div className="overflow-x-auto rounded border border-[rgba(148,163,184,0.1)]">
        <Table>
          <TableHeader>
            <TableRow className="bg-[#0f1219] border-b border-[rgba(148,163,184,0.1)] hover:bg-[#0f1219]">
              <TableHead className="text-[#94a3b8] font-semibold text-xs">Item</TableHead>
              <TableHead className="text-right text-[#94a3b8] font-semibold text-xs">Quantity</TableHead>
              <TableHead className="text-right text-[#94a3b8] font-semibold text-xs">Stockpile</TableHead>
              <TableHead className="text-right text-[#94a3b8] font-semibold text-xs">Volume (m³)</TableHead>
              <TableHead className="text-right text-[#94a3b8] font-semibold text-xs">Unit Price</TableHead>
              <TableHead className="text-right text-[#94a3b8] font-semibold text-xs">Total Value</TableHead>
              <TableHead className="text-right text-[#94a3b8] font-semibold text-xs">Deficit Cost</TableHead>
              {showOwner && <TableHead className="text-[#94a3b8] font-semibold text-xs">Owner</TableHead>}
              <TableHead className="text-center text-[#94a3b8] font-semibold text-xs">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {assetsToRender.map((asset, idx) => (
              <TableRow
                key={idx}
                className={cn(
                  'border-b border-[rgba(148,163,184,0.05)] hover:bg-[rgba(148,163,184,0.04)]',
                  idx % 2 === 0 ? 'bg-[#12151f]' : 'bg-[#0f1219]',
                  asset.stockpileDelta !== undefined && asset.stockpileDelta < 0 && 'border-l-4 border-l-[#ef4444] font-semibold'
                )}
              >
                <TableCell className="py-1.5">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="text-sm text-[#e2e8f0]">{asset.name}</span>
                    {(() => {
                      const listing = getListingForAsset(asset, locationId, containerId, divisionNumber);
                      if (listing) {
                        return (
                          <button
                            onClick={() => handleOpenListingDialog(asset, locationId, containerId, divisionNumber, listing)}
                            className={cn(
                              'inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-semibold border cursor-pointer',
                              listing.autoSellContainerId
                                ? 'bg-[rgba(0,212,255,0.15)] text-[#00d4ff] border-[rgba(0,212,255,0.3)]'
                                : 'bg-[rgba(16,185,129,0.15)] text-[#10b981] border-[rgba(16,185,129,0.3)]'
                            )}
                          >
                            {listing.autoSellContainerId ? (
                              <AutoSellIcon className="text-[#00d4ff]" />
                            ) : (
                              <CheckCircle2 className="h-3.5 w-3.5" />
                            )}
                            {listing.autoSellContainerId ? 'Auto · ' : ''}{formatNumber(listing.quantityAvailable)} @ {formatISK(listing.pricePerUnit)}
                          </button>
                        );
                      }
                      return null;
                    })()}
                  </div>
                </TableCell>
                <TableCell className="text-right py-1.5">
                  <span className="text-sm text-[#e2e8f0]">{formatNumber(asset.quantity)}</span>
                </TableCell>
                <TableCell className="text-right py-1.5">
                  {asset.desiredQuantity ? (
                    <span className="text-sm font-medium">
                      <span className={cn('font-semibold text-sm', asset.stockpileDelta! >= 0 ? 'text-[#10b981]' : 'text-[#ef4444]')}>
                        {asset.stockpileDelta! >= 0 ? '+' : ''}{formatNumber(asset.stockpileDelta!)}
                      </span>
                      <span className="text-[#64748b] mx-1">/</span>
                      <span className="text-[#94a3b8]">{formatNumber(asset.desiredQuantity)}</span>
                    </span>
                  ) : (
                    <span className="text-xs text-[#64748b]">-</span>
                  )}
                </TableCell>
                <TableCell className="text-right py-1.5">
                  <span className="text-sm text-[#94a3b8]">{formatNumber(asset.volume, 2)}</span>
                </TableCell>
                <TableCell className="text-right py-1.5">
                  {asset.unitPrice ? (
                    <span className="text-sm text-[#e2e8f0]">{formatISK(asset.unitPrice)}</span>
                  ) : (
                    <span className="text-xs text-[#64748b]">-</span>
                  )}
                </TableCell>
                <TableCell className="text-right py-1.5">
                  {asset.totalValue ? (
                    <span className="text-sm font-semibold text-[#10b981]">{formatISK(asset.totalValue)}</span>
                  ) : (
                    <span className="text-xs text-[#64748b]">-</span>
                  )}
                </TableCell>
                <TableCell className="text-right py-1.5">
                  {asset.deficitValue && asset.deficitValue > 0 ? (
                    <span className="text-sm font-semibold text-[#ef4444]">{formatISK(asset.deficitValue)}</span>
                  ) : (
                    <span className="text-xs text-[#64748b]">-</span>
                  )}
                </TableCell>
                {showOwner && (
                  <TableCell className="py-1.5">
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border border-[rgba(148,163,184,0.2)] text-[#94a3b8]">
                      {asset.ownerName}
                    </span>
                  </TableCell>
                )}
                <TableCell className="text-center py-1.5">
                  <div className="flex items-center justify-center gap-0.5">
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7 text-[#94a3b8] hover:text-[#e2e8f0] hover:bg-[rgba(148,163,184,0.1)]"
                            onClick={() => handleOpenStockpileModal(asset, locationId, containerId, divisionNumber)}
                          >
                            {asset.desiredQuantity ? <Pencil className="h-3.5 w-3.5" /> : <Plus className="h-3.5 w-3.5" />}
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>Set stockpile target</TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                    {asset.desiredQuantity && (
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-7 w-7 text-[#94a3b8] hover:text-[#ef4444] hover:bg-[rgba(239,68,68,0.1)]"
                              onClick={() => handleDeleteStockpile(asset, locationId, containerId, divisionNumber)}
                            >
                              <Trash2 className="h-3.5 w-3.5" />
                            </Button>
                          </TooltipTrigger>
                          <TooltipContent>Remove stockpile target</TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    )}
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7 text-[#3b82f6] hover:text-[#60a5fa] hover:bg-[rgba(59,130,246,0.1)]"
                            onClick={() => handleOpenListingDialog(asset, locationId, containerId, divisionNumber)}
                          >
                            <Tag className="h-3.5 w-3.5" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>List for sale</TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );

  const renderContainer = (container: AssetContainer, parentId: string, showOwner: boolean, locationId: number, divisionNumber?: number) => {
    const nodeId = `${parentId}-container-${container.id}`;
    const isExpanded = expandedNodes.has(nodeId);
    const autoSellConfig = getAutoSellForContainer(container.id, container.ownerType, container.ownerId, locationId, divisionNumber);
    const autoBuyConfig = getAutoBuyForContainer(container.id, container.ownerType, container.ownerId, locationId, divisionNumber);

    return (
      <div key={container.id}>
        <button
          onClick={() => toggleNode(nodeId)}
          className="w-full flex items-center gap-2 pl-8 pr-3 py-2 hover:bg-[rgba(148,163,184,0.04)] text-left"
        >
          {isExpanded ? <ChevronUp className="h-4 w-4 text-[#94a3b8] flex-shrink-0" /> : <ChevronDown className="h-4 w-4 text-[#94a3b8] flex-shrink-0" />}
          <div className="flex-1 flex items-center gap-2 flex-wrap min-w-0">
            <span className="text-sm font-medium text-[#e2e8f0]">{`📦 ${container.name}`}</span>
            {autoSellConfig && (
              <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-semibold bg-[rgba(0,212,255,0.15)] text-[#00d4ff] border border-[rgba(0,212,255,0.3)]">
                <AutoSellIcon />
                {`Auto-Sell @ ${autoSellConfig.pricePercentage}% ${getPriceSourceAbbrev(autoSellConfig.priceSource)}`}
              </span>
            )}
            {autoBuyConfig && (
              <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-semibold bg-[rgba(245,158,11,0.15)] text-[#f59e0b] border border-[rgba(245,158,11,0.3)]">
                <AutoBuyIcon />
                {`Auto-Buy @ ${autoBuyConfig.minPricePercentage > 0 ? `${autoBuyConfig.minPricePercentage}-` : ''}${autoBuyConfig.maxPricePercentage}% ${getPriceSourceAbbrev(autoBuyConfig.priceSource)}`}
              </span>
            )}
            <span className="text-xs text-[#64748b]">{container.assets.length} items</span>
          </div>
          <div className="flex items-center gap-0.5 flex-shrink-0" onClick={(e) => e.stopPropagation()}>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7 text-[#94a3b8] hover:text-[#e2e8f0] hover:bg-[rgba(148,163,184,0.1)]"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleOpenAddStockpileDialog(locationId, [{ ownerType: container.ownerType, ownerId: container.ownerId, ownerName: container.ownerName }], container.id, divisionNumber);
                    }}
                  >
                    <Package className="h-3.5 w-3.5" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Add Stockpile</TooltipContent>
              </Tooltip>
            </TooltipProvider>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label={autoBuyConfig ? 'Edit Auto-Buy' : 'Enable Auto-Buy'}
                    className={cn('h-7 w-7 hover:bg-[rgba(245,158,11,0.1)]', autoBuyConfig ? 'text-[#f59e0b]' : 'text-[#94a3b8] hover:text-[#f59e0b]')}
                    onClick={(e) => {
                      e.stopPropagation();
                      handleOpenAutoBuyDialog(container.ownerType, container.ownerId, locationId, container.id, container.name, divisionNumber);
                    }}
                  >
                    <AutoBuyIcon />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{autoBuyConfig ? 'Edit Auto-Buy' : 'Enable Auto-Buy'}</TooltipContent>
              </Tooltip>
            </TooltipProvider>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label={autoSellConfig ? 'Edit Auto-Sell' : 'Enable Auto-Sell'}
                    className={cn('h-7 w-7 hover:bg-[rgba(0,212,255,0.1)]', autoSellConfig ? 'text-[#00d4ff]' : 'text-[#94a3b8] hover:text-[#00d4ff]')}
                    onClick={(e) => {
                      e.stopPropagation();
                      handleOpenAutoSellDialog(container.ownerType, container.ownerId, locationId, container.id, container.name, divisionNumber);
                    }}
                  >
                    <AutoSellIcon />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{autoSellConfig ? 'Edit Auto-Sell' : 'Enable Auto-Sell'}</TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        </button>
        {isExpanded && (
          <div>
            {renderAssetsTable(container.assets, showOwner, locationId, container.id, divisionNumber)}
          </div>
        )}
      </div>
    );
  };

  const renderCorporationHanger = (hanger: CorporationHanger, structureId: number) => {
    const nodeId = `structure-${structureId}-corp-${hanger.id}`;
    const isExpanded = expandedNodes.has(nodeId);
    const autoSellConfig = getAutoSellForContainer(undefined, 'corporation', hanger.corporationId, structureId, hanger.id);
    const autoBuyConfig = getAutoBuyForContainer(undefined, 'corporation', hanger.corporationId, structureId, hanger.id);

    return (
      <div key={hanger.id}>
        <button
          onClick={() => toggleNode(nodeId)}
          className="w-full flex items-center gap-2 pl-6 pr-3 py-2 hover:bg-[rgba(148,163,184,0.04)] text-left"
        >
          {isExpanded ? <ChevronUp className="h-4 w-4 text-[#94a3b8] flex-shrink-0" /> : <ChevronDown className="h-4 w-4 text-[#94a3b8] flex-shrink-0" />}
          <div className="flex-1 flex items-center gap-2 flex-wrap min-w-0">
            <span className="text-sm font-medium text-[#e2e8f0]">{`${hanger.corporationName} - ${hanger.name}`}</span>
            {autoSellConfig && (
              <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-semibold bg-[rgba(0,212,255,0.15)] text-[#00d4ff] border border-[rgba(0,212,255,0.3)]">
                <AutoSellIcon />
                {`Auto-Sell @ ${autoSellConfig.pricePercentage}% ${getPriceSourceAbbrev(autoSellConfig.priceSource)}`}
              </span>
            )}
            {autoBuyConfig && (
              <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-semibold bg-[rgba(245,158,11,0.15)] text-[#f59e0b] border border-[rgba(245,158,11,0.3)]">
                <AutoBuyIcon />
                {`Auto-Buy @ ${autoBuyConfig.minPricePercentage > 0 ? `${autoBuyConfig.minPricePercentage}-` : ''}${autoBuyConfig.maxPricePercentage}% ${getPriceSourceAbbrev(autoBuyConfig.priceSource)}`}
              </span>
            )}
            <span className="text-xs text-[#64748b]">{hanger.assets.length} items, {hanger.hangarContainers?.length || 0} containers</span>
          </div>
          <div className="flex items-center gap-0.5 flex-shrink-0" onClick={(e) => e.stopPropagation()}>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7 text-[#94a3b8] hover:text-[#e2e8f0] hover:bg-[rgba(148,163,184,0.1)]"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleOpenAddStockpileDialog(structureId, [{ ownerType: 'corporation', ownerId: hanger.corporationId, ownerName: hanger.corporationName }], undefined, hanger.id);
                    }}
                  >
                    <Package className="h-3.5 w-3.5" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Add Stockpile</TooltipContent>
              </Tooltip>
            </TooltipProvider>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label={autoBuyConfig ? 'Edit Auto-Buy' : 'Enable Auto-Buy'}
                    className={cn('h-7 w-7 hover:bg-[rgba(245,158,11,0.1)]', autoBuyConfig ? 'text-[#f59e0b]' : 'text-[#94a3b8] hover:text-[#f59e0b]')}
                    onClick={(e) => {
                      e.stopPropagation();
                      handleOpenAutoBuyDialog('corporation', hanger.corporationId, structureId, undefined, hanger.name, hanger.id);
                    }}
                  >
                    <AutoBuyIcon />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{autoBuyConfig ? 'Edit Auto-Buy' : 'Enable Auto-Buy'}</TooltipContent>
              </Tooltip>
            </TooltipProvider>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label={autoSellConfig ? 'Edit Auto-Sell' : 'Enable Auto-Sell'}
                    className={cn('h-7 w-7 hover:bg-[rgba(0,212,255,0.1)]', autoSellConfig ? 'text-[#00d4ff]' : 'text-[#94a3b8] hover:text-[#00d4ff]')}
                    onClick={(e) => {
                      e.stopPropagation();
                      handleOpenAutoSellDialog('corporation', hanger.corporationId, structureId, undefined, hanger.name, hanger.id);
                    }}
                  >
                    <AutoSellIcon />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{autoSellConfig ? 'Edit Auto-Sell' : 'Enable Auto-Sell'}</TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        </button>
        {isExpanded && (
          <div>
            {hanger.assets.length > 0 && renderAssetsTable(hanger.assets, false, structureId, undefined, hanger.id)}
            {hanger.hangarContainers?.map((container) =>
              renderContainer(container, nodeId, false, structureId, hanger.id)
            )}
          </div>
        )}
      </div>
    );
  };

  const renderStructureTree = (structure: AssetStructure, isHidden: boolean) => {
    const structureNodeId = `structure-${structure.id}`;
    const isStructureExpanded = expandedNodes.has(structureNodeId);

    return (
      <div key={structure.id}>
        {/* Station/Structure Node */}
        <button
          onClick={() => toggleNode(structureNodeId)}
          className={cn(
            'w-full flex items-center gap-2 px-3 py-2.5 hover:bg-[rgba(148,163,184,0.04)] text-left',
            isHidden && 'opacity-60'
          )}
        >
          {isStructureExpanded ? <ChevronUp className="h-4 w-4 text-[#94a3b8] flex-shrink-0" /> : <ChevronDown className="h-4 w-4 text-[#94a3b8] flex-shrink-0" />}
          <MapPin className="h-4 w-4 text-[#3b82f6] flex-shrink-0" />
          <div className="flex-1 min-w-0">
            <p className="text-sm font-bold text-[#e2e8f0] truncate">{structure.name}</p>
            <p className="text-xs text-[#64748b]">{structure.solarSystem} · {structure.region}</p>
          </div>
          <div className="flex items-center gap-0.5 flex-shrink-0" onClick={(e) => e.stopPropagation()}>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className={cn('h-7 w-7', pinnedStructures.has(structure.id) ? 'text-[#3b82f6]' : 'text-[#94a3b8] hover:text-[#e2e8f0]')}
                    onClick={(e) => {
                      e.stopPropagation();
                      togglePinStructure(structure.id);
                    }}
                  >
                    <Pin className={cn('h-3.5 w-3.5', pinnedStructures.has(structure.id) && 'fill-current')} />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{pinnedStructures.has(structure.id) ? 'Unpin' : 'Pin to top'}</TooltipContent>
              </Tooltip>
            </TooltipProvider>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7 text-[#94a3b8] hover:text-[#e2e8f0]"
                    onClick={(e) => {
                      e.stopPropagation();
                      toggleHideStructure(structure.id);
                    }}
                  >
                    {isHidden ? <Eye className="h-3.5 w-3.5" /> : <EyeOff className="h-3.5 w-3.5" />}
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{isHidden ? 'Show' : 'Hide'}</TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        </button>

        {isStructureExpanded && (
          <div>
            {/* Personal Hangar */}
            {structure.hangarAssets && structure.hangarAssets.length > 0 && (
              <div>
                <button
                  onClick={() => toggleNode(`structure-${structure.id}-hangar`)}
                  className="w-full flex items-center gap-2 pl-6 pr-3 py-2 hover:bg-[rgba(148,163,184,0.04)] text-left"
                >
                  {expandedNodes.has(`structure-${structure.id}-hangar`) ? <ChevronUp className="h-4 w-4 text-[#94a3b8] flex-shrink-0" /> : <ChevronDown className="h-4 w-4 text-[#94a3b8] flex-shrink-0" />}
                  <div className="flex-1">
                    <span className="text-sm font-semibold text-[#e2e8f0]">Personal Hangar</span>
                    <span className="text-xs text-[#64748b] ml-2">{structure.hangarAssets.length} items</span>
                  </div>
                  {!isHidden && (
                    <div onClick={(e) => e.stopPropagation()}>
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-7 w-7 text-[#94a3b8] hover:text-[#e2e8f0] hover:bg-[rgba(148,163,184,0.1)]"
                              onClick={(e) => {
                                e.stopPropagation();
                                const seen = new Map<string, { ownerType: string; ownerId: number; ownerName: string }>();
                                for (const a of structure.hangarAssets) {
                                  const key = `${a.ownerType}:${a.ownerId}`;
                                  if (!seen.has(key)) seen.set(key, { ownerType: a.ownerType, ownerId: a.ownerId, ownerName: a.ownerName });
                                }
                                handleOpenAddStockpileDialog(structure.id, Array.from(seen.values()));
                              }}
                            >
                              <Package className="h-3.5 w-3.5" />
                            </Button>
                          </TooltipTrigger>
                          <TooltipContent>Add Stockpile</TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </div>
                  )}
                </button>
                {expandedNodes.has(`structure-${structure.id}-hangar`) && (
                  <div>
                    {renderAssetsTable(structure.hangarAssets, true, structure.id)}
                  </div>
                )}
              </div>
            )}

            {/* Personal Hangar Containers */}
            {structure.hangarContainers && structure.hangarContainers.length > 0 &&
              structure.hangarContainers.map((container) =>
                renderContainer(container, `structure-${structure.id}`, true, structure.id)
              )}

            {/* Deliveries */}
            {structure.deliveries && structure.deliveries.length > 0 && (
              <div>
                <button
                  onClick={() => toggleNode(`structure-${structure.id}-deliveries`)}
                  className="w-full flex items-center gap-2 pl-6 pr-3 py-2 hover:bg-[rgba(148,163,184,0.04)] text-left"
                >
                  {expandedNodes.has(`structure-${structure.id}-deliveries`) ? <ChevronUp className="h-4 w-4 text-[#94a3b8] flex-shrink-0" /> : <ChevronDown className="h-4 w-4 text-[#94a3b8] flex-shrink-0" />}
                  <div className="flex-1">
                    <span className="text-sm font-semibold text-[#e2e8f0]">📬 Deliveries</span>
                    <span className="text-xs text-[#64748b] ml-2">{structure.deliveries.length} items</span>
                  </div>
                </button>
                {expandedNodes.has(`structure-${structure.id}-deliveries`) && (
                  <div>{renderAssetsTable(structure.deliveries, true, structure.id)}</div>
                )}
              </div>
            )}

            {/* Asset Safety */}
            {structure.assetSafety && structure.assetSafety.length > 0 && (
              <div>
                <button
                  onClick={() => toggleNode(`structure-${structure.id}-safety`)}
                  className="w-full flex items-center gap-2 pl-6 pr-3 py-2 hover:bg-[rgba(148,163,184,0.04)] text-left"
                >
                  {expandedNodes.has(`structure-${structure.id}-safety`) ? <ChevronUp className="h-4 w-4 text-[#94a3b8] flex-shrink-0" /> : <ChevronDown className="h-4 w-4 text-[#94a3b8] flex-shrink-0" />}
                  <div className="flex-1">
                    <span className="text-sm font-semibold text-[#e2e8f0]">🛡️ Asset Safety</span>
                    <span className="text-xs text-[#64748b] ml-2">{structure.assetSafety.length} items</span>
                  </div>
                </button>
                {expandedNodes.has(`structure-${structure.id}-safety`) && (
                  <div>{renderAssetsTable(structure.assetSafety, true, structure.id)}</div>
                )}
              </div>
            )}

            {/* Corporation Hangars */}
            {structure.corporationHangers && structure.corporationHangers.length > 0 &&
              structure.corporationHangers
                .sort((a, b) => a.id - b.id)
                .map((hanger) => renderCorporationHanger(hanger, structure.id))}
          </div>
        )}
      </div>
    );
  };

  return (
    <>
      <Navbar />
      <div className="mt-2 mb-2 px-6">
        {/* Sticky Header Section */}
        <div className="sticky top-16 z-[100] bg-[#0a0e1a] pb-3 mb-3">
          <div className="flex items-center justify-between mb-3 pt-1">
            <div>
              <h2 className="text-xl font-semibold text-[#e2e8f0]">Asset Inventory</h2>
              {assetStatus && (
                <p className="text-[#94a3b8] text-xs mt-0.5">
                  {assetStatus.lastUpdatedAt ? (
                    <>
                      Last updated: {formatRelativeTime(new Date(assetStatus.lastUpdatedAt))}
                      {assetStatus.nextUpdateAt && (
                        <> | Next update: {formatRelativeTime(new Date(assetStatus.nextUpdateAt))}</>
                      )}
                    </>
                  ) : (
                    'Assets have not been updated yet'
                  )}
                </p>
              )}
            </div>

            {/* Summary Stats */}
            <div className="flex gap-2.5 items-center flex-wrap">
              <div className="flex items-center gap-2 px-3 py-2 rounded-lg bg-[rgba(0,212,255,0.08)] border border-[rgba(0,212,255,0.2)]">
                <Package className="h-5 w-5 text-[#00d4ff]" />
                <div>
                  <p className="text-sm font-bold text-[#e2e8f0]">{formatCompact(totalItems)}</p>
                  <p className="text-[0.7rem] text-[#94a3b8]">Items</p>
                </div>
              </div>
              <div className="flex items-center gap-2 px-3 py-2 rounded-lg bg-[rgba(139,92,246,0.08)] border border-[rgba(139,92,246,0.2)]">
                <Layers className="h-5 w-5 text-[#8b5cf6]" />
                <div>
                  <p className="text-sm font-bold text-[#e2e8f0]">{formatCompact(uniqueTypes)}</p>
                  <p className="text-[0.7rem] text-[#94a3b8]">Types</p>
                </div>
              </div>
              <div className="flex items-center gap-2 px-3 py-2 rounded-lg bg-[rgba(245,158,11,0.08)] border border-[rgba(245,158,11,0.2)]">
                <MapPin className="h-5 w-5 text-[#f59e0b]" />
                <div>
                  <p className="text-sm font-bold text-[#e2e8f0]">{formatCompact(totalVolume)}</p>
                  <p className="text-[0.7rem] text-[#94a3b8]">m³</p>
                </div>
              </div>
              <div className="flex items-center gap-2 px-3 py-2 rounded-lg bg-[rgba(16,185,129,0.08)] border border-[rgba(16,185,129,0.2)]">
                <DollarSign className="h-5 w-5 text-[#10b981]" />
                <div>
                  <p className="text-sm font-bold text-[#10b981]">{formatISK(totalValue)}</p>
                  <p className="text-[0.7rem] text-[#94a3b8]">Total Value</p>
                </div>
              </div>
              {totalDeficit > 0 && (
                <div className="flex items-center gap-2 px-3 py-2 rounded-lg bg-[rgba(239,68,68,0.08)] border border-[rgba(239,68,68,0.2)]">
                  <AlertTriangle className="h-5 w-5 text-[#ef4444]" />
                  <div>
                    <p className="text-sm font-bold text-[#ef4444]">{formatISK(totalDeficit)}</p>
                    <p className="text-[0.7rem] text-[#94a3b8]">Deficit Cost</p>
                  </div>
                </div>
              )}
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={handleRefreshPrices}
                      disabled={refreshingPrices}
                      className="h-9 w-9 text-[#94a3b8] hover:text-[#e2e8f0]"
                    >
                      {refreshingPrices ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      ) : (
                        <RefreshCw className="h-4 w-4" />
                      )}
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Refresh market prices</TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
          </div>

          {/* Search Bar and Filter */}
          <div className="flex gap-3 items-center">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-2.5 h-4 w-4 text-[#64748b]" />
              <Input
                placeholder="Search items..."
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                className="pl-9 bg-[#12151f] border-[rgba(148,163,184,0.2)] text-[#e2e8f0] placeholder:text-[#64748b]"
              />
            </div>
            <div className="flex items-center gap-2 whitespace-nowrap">
              <Switch
                id="below-target"
                checked={showBelowTargetOnly}
                onCheckedChange={setShowBelowTargetOnly}
              />
              <Label htmlFor="below-target" className="text-sm text-[#e2e8f0] cursor-pointer">
                Below target only
              </Label>
            </div>
          </div>
        </div>

        {/* Tree View - Visible Stations */}
        <Card className="bg-[#12151f] border-[rgba(148,163,184,0.15)]">
          <CardContent className="p-0">
            <div className="divide-y divide-[rgba(148,163,184,0.05)]">
              {visibleStructures.map((structure) => renderStructureTree(structure, false))}
            </div>
          </CardContent>
        </Card>

        {/* Hidden Stations */}
        {hiddenStructuresList.length > 0 && (
          <Card className="mt-4 bg-[#12151f] border-[rgba(148,163,184,0.15)]">
            <CardContent className="p-0">
              <div className="flex items-center gap-2 px-4 py-3 border-b border-[rgba(148,163,184,0.1)]">
                <EyeOff className="h-4 w-4 text-[#94a3b8]" />
                <h3 className="text-base font-semibold text-[#e2e8f0]">
                  Hidden Stations ({hiddenStructuresList.length})
                </h3>
              </div>
              <div className="divide-y divide-[rgba(148,163,184,0.05)]">
                {hiddenStructuresList.map((structure) => renderStructureTree(structure, true))}
              </div>
            </CardContent>
          </Card>
        )}

        {filteredStructures.length === 0 && searchQuery && (
          <Card className="bg-[#12151f] border-[rgba(148,163,184,0.15)]">
            <CardContent className="py-8 text-center">
              <p className="text-sm text-[#94a3b8]">
                No items found matching &quot;{searchQuery}&quot;
              </p>
            </CardContent>
          </Card>
        )}

        {/* Stockpile Modal */}
        <Dialog
          open={stockpileModalOpen}
          onOpenChange={(o) => {
            if (!o) setStockpileModalOpen(false);
            else desiredQuantityInputRef.current?.focus();
          }}
        >
          <DialogContent className="sm:max-w-md bg-[#12151f] border border-[rgba(148,163,184,0.15)] text-[#e2e8f0]">
            <DialogHeader>
              <DialogTitle className="text-[#e2e8f0]">
                {selectedAsset?.asset.desiredQuantity ? 'Edit' : 'Set'} Stockpile Marker
              </DialogTitle>
            </DialogHeader>
            <div className="pt-1 flex flex-col gap-3">
              <p className="text-sm text-[#e2e8f0]"><strong>Item:</strong> {selectedAsset?.asset.name}</p>
              <p className="text-sm text-[#e2e8f0] mb-1"><strong>Current Quantity:</strong> {selectedAsset?.asset.quantity.toLocaleString()}</p>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="stockpile-qty" className="text-xs text-[#94a3b8]">Desired Quantity *</Label>
                <Input
                  id="stockpile-qty"
                  type="text"
                  value={desiredQuantity}
                  onChange={(e) => handleQuantityChange(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && desiredQuantity) {
                      e.preventDefault();
                      handleSaveStockpile();
                    }
                  }}
                  placeholder="0"
                  ref={desiredQuantityInputRef}
                  className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0]"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="stockpile-notes" className="text-xs text-[#94a3b8]">Notes (optional)</Label>
                <textarea
                  id="stockpile-notes"
                  rows={3}
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  className="w-full rounded-md border border-[rgba(148,163,184,0.2)] bg-[#0f1219] px-3 py-2 text-sm text-[#e2e8f0] placeholder:text-[#64748b] resize-none focus:outline-none focus:ring-1 focus:ring-[#3b82f6]"
                />
              </div>
              <Separator className="bg-[rgba(148,163,184,0.1)]" />
              <p className="text-xs font-semibold text-[#94a3b8]">Auto-Production</p>
              <div className="flex items-center gap-2">
                <Switch
                  id="stockpile-auto-prod"
                  checked={autoProductionEnabled}
                  onCheckedChange={setAutoProductionEnabled}
                />
                <Label htmlFor="stockpile-auto-prod" className="text-sm text-[#e2e8f0] cursor-pointer">
                  Enable Auto-Production
                </Label>
              </div>
              {autoProductionEnabled && (
                <div className="flex flex-col gap-3">
                  <div className="flex flex-col gap-1.5">
                    <Label className="text-xs text-[#94a3b8]">Production Plan</Label>
                    <Select
                      value={selectedPlanId !== null ? String(selectedPlanId) : ''}
                      onValueChange={(v) => setSelectedPlanId(v ? Number(v) : null)}
                      disabled={plansLoading || availablePlans.length === 0}
                    >
                      <SelectTrigger className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0]">
                        <SelectValue
                          placeholder={
                            plansLoading
                              ? 'Loading plans...'
                              : availablePlans.length === 0
                              ? 'No plans for this item'
                              : 'Select a plan...'
                          }
                        />
                      </SelectTrigger>
                      <SelectContent className="bg-[#1a1f2e] border-[rgba(148,163,184,0.15)]">
                        {availablePlans.map((plan) => (
                          <SelectItem
                            key={plan.id}
                            value={String(plan.id)}
                            className="text-[#e2e8f0] focus:bg-[rgba(0,212,255,0.08)]"
                          >
                            {plan.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="stockpile-parallelism" className="text-xs text-[#94a3b8]">Max Parallelism</Label>
                    <Input
                      id="stockpile-parallelism"
                      type="number"
                      min={0}
                      value={parallelism}
                      onChange={(e) => setParallelism(Math.max(0, parseInt(e.target.value, 10) || 0))}
                      className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0]"
                    />
                    <p className="text-xs text-[#64748b]">0 = no character assignment</p>
                  </div>
                </div>
              )}
            </div>
            <DialogFooter className="gap-2">
              <Button variant="outline" onClick={() => setStockpileModalOpen(false)}
                className="border-[rgba(148,163,184,0.2)] text-[#e2e8f0] hover:bg-[rgba(148,163,184,0.1)]">
                Cancel
              </Button>
              <Button
                onClick={handleSaveStockpile}
                disabled={!desiredQuantity}
                className="bg-[#3b82f6] hover:bg-[#2563eb] text-white"
              >
                Save
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* List for Sale Dialog */}
        <Dialog open={listingDialogOpen} onOpenChange={(o) => { if (!o) setListingDialogOpen(false); }}>
          <DialogContent className="sm:max-w-md bg-[#12151f] border border-[rgba(148,163,184,0.15)] text-[#e2e8f0]">
            <DialogHeader>
              <DialogTitle className="text-[#e2e8f0]">
                {editingListingId ? 'Edit Listing' : 'List Item for Sale'}
              </DialogTitle>
            </DialogHeader>
            <div className="pt-1 flex flex-col gap-3">
              <p className="text-sm text-[#e2e8f0]"><strong>Item:</strong> {listingAsset?.asset.name}</p>
              <p className="text-sm text-[#e2e8f0]"><strong>Owner:</strong> {listingAsset?.asset.ownerName}</p>
              <p className="text-sm text-[#e2e8f0] mb-1"><strong>Available Quantity:</strong> {listingAsset?.asset.quantity.toLocaleString()}</p>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="listing-qty" className="text-xs text-[#94a3b8]">Quantity to List *</Label>
                <Input
                  id="listing-qty"
                  type="text"
                  ref={listingQuantityRef}
                  onBlur={() => handleListingInputChange(listingQuantityRef)}
                  placeholder="0"
                  className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0]"
                />
                <p className="text-xs text-[#64748b]">Max: {listingAsset?.asset.quantity.toLocaleString() || 0}</p>
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="listing-price" className="text-xs text-[#94a3b8]">Price Per Unit (ISK) *</Label>
                <Input
                  id="listing-price"
                  type="text"
                  ref={listingPriceRef}
                  onBlur={() => handleListingInputChange(listingPriceRef, true)}
                  placeholder="0"
                  className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0]"
                />
                {listingTotalValue && (
                  <p className="text-xs text-[#64748b]">Total Value: {listingTotalValue} ISK</p>
                )}
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="listing-notes" className="text-xs text-[#94a3b8]">Notes (optional)</Label>
                <Input
                  id="listing-notes"
                  type="text"
                  ref={listingNotesRef}
                  placeholder="Add any notes about this listing..."
                  className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0]"
                />
              </div>
            </div>
            <DialogFooter className="gap-2">
              {editingListingId && (
                <Button
                  variant="destructive"
                  onClick={handleDeleteListing}
                  disabled={submittingListing}
                  className="mr-auto bg-[rgba(239,68,68,0.15)] text-[#ef4444] hover:bg-[rgba(239,68,68,0.25)] border border-[rgba(239,68,68,0.3)]"
                >
                  Delete
                </Button>
              )}
              <Button variant="outline" onClick={() => setListingDialogOpen(false)}
                className="border-[rgba(148,163,184,0.2)] text-[#e2e8f0] hover:bg-[rgba(148,163,184,0.1)]">
                Cancel
              </Button>
              <Button
                onClick={handleCreateListing}
                disabled={submittingListing}
                className="bg-[#3b82f6] hover:bg-[#2563eb] text-white"
              >
                {submittingListing ? (editingListingId ? 'Updating...' : 'Creating...') : (editingListingId ? 'Update Listing' : 'Create Listing')}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Auto-Sell Configuration Dialog */}
        <Dialog open={autoSellDialogOpen} onOpenChange={(o) => { if (!o) setAutoSellDialogOpen(false); }}>
          <DialogContent className="sm:max-w-md bg-[#12151f] border border-[rgba(148,163,184,0.15)] text-[#e2e8f0]">
            <DialogHeader>
              <DialogTitle className="text-[#e2e8f0]">
                {autoSellContainer && getAutoSellForContainer(
                  autoSellContainer.containerId,
                  autoSellContainer.ownerType,
                  autoSellContainer.ownerId,
                  autoSellContainer.locationId,
                  autoSellContainer.divisionNumber
                ) ? 'Edit Auto-Sell' : 'Enable Auto-Sell'}
              </DialogTitle>
            </DialogHeader>
            <div className="pt-1 flex flex-col gap-3">
              <p className="text-sm text-[#e2e8f0]">
                <strong>{autoSellContainer?.containerId ? 'Container' : 'Division'}:</strong> {autoSellContainer?.containerName}
              </p>
              <p className="text-sm text-[#94a3b8] mb-1">
                All items in this {autoSellContainer?.containerId ? 'container' : 'hangar division'} will be automatically listed for sale at the specified percentage of the selected price source. Listings sync on asset refresh and market price updates.
              </p>
              <div className="flex flex-col gap-1.5">
                <Label className="text-xs text-[#94a3b8]">Price Source</Label>
                <Select value={autoSellPriceSource} onValueChange={setAutoSellPriceSource}>
                  <SelectTrigger className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-[#1a1f2e] border-[rgba(148,163,184,0.15)]">
                    {PRICE_SOURCE_OPTIONS.map((option) => (
                      <SelectItem key={option.value} value={option.value} className="text-[#e2e8f0] focus:bg-[rgba(0,212,255,0.08)]">
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="auto-sell-pct" className="text-xs text-[#94a3b8]">
                  Price Percentage of {getPriceSourceLabel(autoSellPriceSource)}
                </Label>
                <div className="relative">
                  <Input
                    id="auto-sell-pct"
                    type="number"
                    min={1}
                    max={200}
                    step={0.5}
                    value={autoSellPercentage}
                    onChange={(e) => setAutoSellPercentage(e.target.value)}
                    className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0] pr-8"
                  />
                  <span className="absolute right-3 top-2.5 text-sm text-[#64748b]">%</span>
                </div>
                <p className="text-xs text-[#64748b]">
                  Items will be listed at {autoSellPercentage}% of {getPriceSourceLabel(autoSellPriceSource).toLowerCase()}
                </p>
              </div>
            </div>
            <DialogFooter className="gap-2">
              {autoSellContainer && getAutoSellForContainer(
                autoSellContainer.containerId,
                autoSellContainer.ownerType,
                autoSellContainer.ownerId,
                autoSellContainer.locationId,
                autoSellContainer.divisionNumber
              ) && (
                <Button
                  variant="destructive"
                  onClick={handleDisableAutoSell}
                  disabled={submittingAutoSell}
                  className="mr-auto bg-[rgba(239,68,68,0.15)] text-[#ef4444] hover:bg-[rgba(239,68,68,0.25)] border border-[rgba(239,68,68,0.3)]"
                >
                  Disable
                </Button>
              )}
              <Button variant="outline" onClick={() => setAutoSellDialogOpen(false)}
                className="border-[rgba(148,163,184,0.2)] text-[#e2e8f0] hover:bg-[rgba(148,163,184,0.1)]">
                Cancel
              </Button>
              <Button
                onClick={handleSaveAutoSell}
                disabled={submittingAutoSell || !autoSellPercentage || parseFloat(autoSellPercentage) <= 0 || parseFloat(autoSellPercentage) > 200}
                className="bg-[#3b82f6] hover:bg-[#2563eb] text-white"
              >
                {submittingAutoSell ? 'Saving...' : 'Save'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Auto-Buy Configuration Dialog */}
        <Dialog open={autoBuyDialogOpen} onOpenChange={(o) => { if (!o) setAutoBuyDialogOpen(false); }}>
          <DialogContent className="sm:max-w-md bg-[#12151f] border border-[rgba(148,163,184,0.15)] text-[#e2e8f0]">
            <DialogHeader>
              <DialogTitle className="text-[#e2e8f0]">
                {autoBuyContainer && getAutoBuyForContainer(
                  autoBuyContainer.containerId,
                  autoBuyContainer.ownerType,
                  autoBuyContainer.ownerId,
                  autoBuyContainer.locationId,
                  autoBuyContainer.divisionNumber
                ) ? 'Edit Auto-Buy' : 'Enable Auto-Buy'}
              </DialogTitle>
            </DialogHeader>
            <div className="pt-1 flex flex-col gap-3">
              <p className="text-sm text-[#e2e8f0]">
                <strong>{autoBuyContainer?.containerId ? 'Container' : 'Division'}:</strong> {autoBuyContainer?.containerName}
              </p>
              <p className="text-sm text-[#94a3b8] mb-1">
                Buy orders will be automatically created for understocked stockpile items in this {autoBuyContainer?.containerId ? 'container' : 'hangar division'} at the specified percentage of the selected price source. Orders sync on asset refresh and market price updates.
              </p>
              <div className="flex flex-col gap-1.5">
                <Label className="text-xs text-[#94a3b8]">Price Source</Label>
                <Select value={autoBuyPriceSource} onValueChange={setAutoBuyPriceSource}>
                  <SelectTrigger className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-[#1a1f2e] border-[rgba(148,163,184,0.15)]">
                    {PRICE_SOURCE_OPTIONS.map((option) => (
                      <SelectItem key={option.value} value={option.value} className="text-[#e2e8f0] focus:bg-[rgba(0,212,255,0.08)]">
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="flex gap-3">
                <div className="flex flex-col gap-1.5 flex-1">
                  <Label htmlFor="auto-buy-min" className="text-xs text-[#94a3b8]">
                    Min % of {getPriceSourceLabel(autoBuyPriceSource)}
                  </Label>
                  <div className="relative">
                    <Input
                      id="auto-buy-min"
                      type="number"
                      min={0}
                      max={200}
                      step={0.5}
                      value={autoBuyMinPercentage}
                      onChange={(e) => setAutoBuyMinPercentage(e.target.value)}
                      className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0] pr-8"
                    />
                    <span className="absolute right-3 top-2.5 text-sm text-[#64748b]">%</span>
                  </div>
                  <p className="text-xs text-[#64748b]">Floor price for auto-fulfill matching</p>
                </div>
                <div className="flex flex-col gap-1.5 flex-1">
                  <Label htmlFor="auto-buy-max" className="text-xs text-[#94a3b8]">
                    Max % of {getPriceSourceLabel(autoBuyPriceSource)}
                  </Label>
                  <div className="relative">
                    <Input
                      id="auto-buy-max"
                      type="number"
                      min={1}
                      max={200}
                      step={0.5}
                      value={autoBuyMaxPercentage}
                      onChange={(e) => setAutoBuyMaxPercentage(e.target.value)}
                      className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0] pr-8"
                    />
                    <span className="absolute right-3 top-2.5 text-sm text-[#64748b]">%</span>
                  </div>
                  <p className="text-xs text-[#64748b]">Ceiling price for buy orders</p>
                </div>
              </div>
            </div>
            <DialogFooter className="gap-2">
              {autoBuyContainer && getAutoBuyForContainer(
                autoBuyContainer.containerId,
                autoBuyContainer.ownerType,
                autoBuyContainer.ownerId,
                autoBuyContainer.locationId,
                autoBuyContainer.divisionNumber
              ) && (
                <Button
                  variant="destructive"
                  onClick={handleDisableAutoBuy}
                  disabled={submittingAutoBuy}
                  className="mr-auto bg-[rgba(239,68,68,0.15)] text-[#ef4444] hover:bg-[rgba(239,68,68,0.25)] border border-[rgba(239,68,68,0.3)]"
                >
                  Disable
                </Button>
              )}
              <Button variant="outline" onClick={() => setAutoBuyDialogOpen(false)}
                className="border-[rgba(148,163,184,0.2)] text-[#e2e8f0] hover:bg-[rgba(148,163,184,0.1)]">
                Cancel
              </Button>
              <Button
                onClick={handleSaveAutoBuy}
                disabled={submittingAutoBuy || !autoBuyMaxPercentage || parseFloat(autoBuyMaxPercentage) <= 0 || parseFloat(autoBuyMaxPercentage) > 200 || parseFloat(autoBuyMinPercentage) < 0 || parseFloat(autoBuyMinPercentage) > parseFloat(autoBuyMaxPercentage)}
                className="bg-[#3b82f6] hover:bg-[#2563eb] text-white"
              >
                {submittingAutoBuy ? 'Saving...' : 'Save'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Add Stockpile Dialog */}
        {addStockpileContext && (
          <AddStockpileDialog
            open={addStockpileDialogOpen}
            onClose={() => setAddStockpileDialogOpen(false)}
            onSaved={handleAddStockpileSaved}
            locationId={addStockpileContext.locationId}
            containerId={addStockpileContext.containerId}
            divisionNumber={addStockpileContext.divisionNumber}
            owners={addStockpileContext.owners}
          />
        )}
      </div>
    </>
  );
}
