import { useState, useMemo, useEffect, useRef } from 'react';
import { AssetsResponse, Asset, AssetContainer, CorporationHanger, StockpileMarker } from "@industry-tool/client/data/models";
import { useSession } from "next-auth/react";
import Navbar from "@industry-tool/components/Navbar";
import Container from '@mui/material/Container';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Grid from '@mui/material/Grid';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import SearchIcon from '@mui/icons-material/Search';
import LocationOnIcon from '@mui/icons-material/LocationOn';
import InventoryIcon from '@mui/icons-material/Inventory';
import CategoryIcon from '@mui/icons-material/Category';
import RefreshIcon from '@mui/icons-material/Refresh';
import AttachMoneyIcon from '@mui/icons-material/AttachMoney';
import WarningIcon from '@mui/icons-material/Warning';
import List from '@mui/material/List';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemText from '@mui/material/ListItemText';
import Collapse from '@mui/material/Collapse';
import ExpandLess from '@mui/icons-material/ExpandLess';
import ExpandMore from '@mui/icons-material/ExpandMore';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import VisibilityOffIcon from '@mui/icons-material/VisibilityOff';
import VisibilityIcon from '@mui/icons-material/Visibility';
import PushPinIcon from '@mui/icons-material/PushPin';
import PushPinOutlinedIcon from '@mui/icons-material/PushPinOutlined';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';

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
  const desiredQuantityInputRef = useRef<HTMLInputElement>(null);
  const [refreshingPrices, setRefreshingPrices] = useState(false);

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

  // Show loading state first, before checking if assets are empty
  if (loading) {
    return (
      <>
        <Navbar />
        <Container maxWidth={false} sx={{ mt: 2, mb: 2, px: 3 }}>
          <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
            <Typography variant="h6" color="text.secondary">Loading assets...</Typography>
          </Box>
        </Container>
      </>
    );
  }

  if (!assets?.structures || assets.structures.length === 0) {
    return (
      <>
        <Navbar />
        <Container maxWidth="xl" sx={{ mt: 4 }}>
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              minHeight: '60vh',
              textAlign: 'center',
            }}
          >
            <Typography variant="h4" gutterBottom>
              No Assets Found
            </Typography>
            <Typography variant="body1" color="text.secondary">
              You don't have any assets yet, or they haven't been synced.
            </Typography>
          </Box>
        </Container>
      </>
    );
  }

  const renderAssetsTable = (assets: Asset[], showOwner: boolean, locationId: number, containerId?: number, divisionNumber?: number) => (
    <Box sx={{ px: 2, pb: 1 }}>
      <TableContainer component={Paper} variant="outlined">
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Item</TableCell>
              <TableCell align="right">Quantity</TableCell>
              <TableCell align="right">Stockpile</TableCell>
              <TableCell align="right">Volume (m³)</TableCell>
              <TableCell align="right">Unit Price</TableCell>
              <TableCell align="right">Total Value</TableCell>
              <TableCell align="right">Deficit Cost</TableCell>
              {showOwner && <TableCell>Owner</TableCell>}
              <TableCell align="center">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {assets.map((asset, idx) => (
              <TableRow
                key={idx}
                hover
                sx={{
                  '&:nth-of-type(odd)': {
                    backgroundColor: 'action.hover',
                  },
                  ...(asset.stockpileDelta !== undefined && asset.stockpileDelta < 0 && {
                    borderLeft: '4px solid #d32f2f',
                    '& .MuiTableCell-root': {
                      fontWeight: 600,
                    }
                  }),
                }}
              >
                <TableCell>{asset.name}</TableCell>
                <TableCell align="right">{asset.quantity.toLocaleString()}</TableCell>
                <TableCell align="right">
                  {asset.desiredQuantity ? (
                    <Typography variant="body2" sx={{ fontWeight: 500 }}>
                      <Box
                        component="span"
                        sx={{
                          color: asset.stockpileDelta! >= 0 ? 'success.main' : 'error.main',
                          fontWeight: 600,
                          fontSize: '1rem'
                        }}
                      >
                        {asset.stockpileDelta! >= 0 ? '+' : ''}{asset.stockpileDelta!.toLocaleString()}
                      </Box>
                      {' / '}
                      {asset.desiredQuantity.toLocaleString()}
                    </Typography>
                  ) : (
                    <Typography variant="caption" color="text.secondary">-</Typography>
                  )}
                </TableCell>
                <TableCell align="right">
                  {asset.volume.toLocaleString(undefined, { maximumFractionDigits: 2 })}
                </TableCell>
                {/* Unit Price */}
                <TableCell align="right">
                  {asset.unitPrice ? (
                    <Typography variant="body2">
                      {asset.unitPrice.toLocaleString(undefined, {
                        minimumFractionDigits: 2,
                        maximumFractionDigits: 2
                      })} ISK
                    </Typography>
                  ) : (
                    <Typography variant="caption" color="text.secondary">-</Typography>
                  )}
                </TableCell>
                {/* Total Value */}
                <TableCell align="right">
                  {asset.totalValue ? (
                    <Typography variant="body2" fontWeight={600}>
                      {asset.totalValue.toLocaleString(undefined, {
                        maximumFractionDigits: 0
                      })} ISK
                    </Typography>
                  ) : (
                    <Typography variant="caption" color="text.secondary">-</Typography>
                  )}
                </TableCell>
                {/* Deficit Cost */}
                <TableCell align="right">
                  {asset.deficitValue && asset.deficitValue > 0 ? (
                    <Typography variant="body2" fontWeight={600} sx={{ color: 'error.main' }}>
                      {asset.deficitValue.toLocaleString(undefined, {
                        maximumFractionDigits: 0
                      })} ISK
                    </Typography>
                  ) : (
                    <Typography variant="caption" color="text.secondary">-</Typography>
                  )}
                </TableCell>
                {showOwner && (
                  <TableCell>
                    <Chip
                      label={asset.ownerName}
                      size="small"
                      variant="outlined"
                      sx={{ height: 20, fontSize: '0.7rem' }}
                    />
                  </TableCell>
                )}
                <TableCell align="center">
                  <IconButton
                    size="small"
                    onClick={() => handleOpenStockpileModal(asset, locationId, containerId, divisionNumber)}
                  >
                    {asset.desiredQuantity ? <EditIcon fontSize="small" /> : <AddIcon fontSize="small" />}
                  </IconButton>
                  {asset.desiredQuantity && (
                    <IconButton
                      size="small"
                      onClick={() => handleDeleteStockpile(asset, locationId, containerId, divisionNumber)}
                    >
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );

  const renderContainer = (container: AssetContainer, parentId: string, showOwner: boolean, locationId: number, divisionNumber?: number) => {
    const nodeId = `${parentId}-container-${container.id}`;
    const isExpanded = expandedNodes.has(nodeId);

    return (
      <Box key={container.id}>
        <ListItemButton onClick={() => toggleNode(nodeId)} sx={{ pl: 3 }}>
          {isExpanded ? <ExpandLess /> : <ExpandMore />}
          <ListItemText
            primary={`📦 ${container.name}`}
            secondary={`${container.assets.length} items`}
            primaryTypographyProps={{ variant: 'body2', fontWeight: 500 }}
          />
        </ListItemButton>
        <Collapse in={isExpanded} timeout={150} unmountOnExit>
          {renderAssetsTable(container.assets, showOwner, locationId, container.id, divisionNumber)}
        </Collapse>
      </Box>
    );
  };

  const renderCorporationHanger = (hanger: CorporationHanger, structureId: number) => {
    const nodeId = `structure-${structureId}-corp-${hanger.id}`;
    const isExpanded = expandedNodes.has(nodeId);

    return (
      <Box key={hanger.id}>
        <ListItemButton onClick={() => toggleNode(nodeId)} sx={{ pl: 2 }}>
          {isExpanded ? <ExpandLess /> : <ExpandMore />}
          <ListItemText
            primary={`${hanger.corporationName} - ${hanger.name}`}
            secondary={`${hanger.assets.length} items, ${hanger.hangarContainers?.length || 0} containers`}
            primaryTypographyProps={{ variant: 'body2', fontWeight: 500 }}
          />
        </ListItemButton>
        <Collapse in={isExpanded} timeout={150} unmountOnExit>
          <Box>
            {hanger.assets.length > 0 && renderAssetsTable(hanger.assets, false, structureId, undefined, hanger.id)}
            {hanger.hangarContainers?.map((container) =>
              renderContainer(container, nodeId, false, structureId, hanger.id)
            )}
          </Box>
        </Collapse>
      </Box>
    );
  };

  return (
    <>
      <Navbar />
      <Container maxWidth={false} sx={{ mt: 2, mb: 2, px: 3 }}>
        {/* Sticky Header Section */}
        <Box
          sx={{
            position: 'sticky',
            top: 64,
            zIndex: 100,
            backgroundColor: 'background.default',
            pb: 1.5,
            mb: 1.5,
          }}
        >
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 1.5, pt: 0.5 }}>
            <Typography variant="h5">Asset Inventory</Typography>

            {/* Summary Stats */}
            <Box sx={{ display: 'flex', gap: 3, alignItems: 'center' }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                <InventoryIcon sx={{ fontSize: 20, color: 'primary.main' }} />
                <Box>
                  <Typography variant="body2" fontWeight={600}>{totalItems.toLocaleString()}</Typography>
                  <Typography variant="caption" color="text.secondary">Items</Typography>
                </Box>
              </Box>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                <CategoryIcon sx={{ fontSize: 20, color: 'primary.main' }} />
                <Box>
                  <Typography variant="body2" fontWeight={600}>{uniqueTypes.toLocaleString()}</Typography>
                  <Typography variant="caption" color="text.secondary">Types</Typography>
                </Box>
              </Box>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                <LocationOnIcon sx={{ fontSize: 20, color: 'primary.main' }} />
                <Box>
                  <Typography variant="body2" fontWeight={600}>
                    {totalVolume.toLocaleString(undefined, { maximumFractionDigits: 0 })}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">m³</Typography>
                </Box>
              </Box>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                <AttachMoneyIcon sx={{ fontSize: 20, color: 'success.main' }} />
                <Box>
                  <Typography variant="body2" fontWeight={600}>
                    {totalValue.toLocaleString(undefined, { maximumFractionDigits: 0 })} ISK
                  </Typography>
                  <Typography variant="caption" color="text.secondary">Total Value</Typography>
                </Box>
              </Box>
              {totalDeficit > 0 && (
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                  <WarningIcon sx={{ fontSize: 20, color: 'error.main' }} />
                  <Box>
                    <Typography variant="body2" fontWeight={600} color="error.main">
                      {totalDeficit.toLocaleString(undefined, { maximumFractionDigits: 0 })} ISK
                    </Typography>
                    <Typography variant="caption" color="text.secondary">Deficit Cost</Typography>
                  </Box>
                </Box>
              )}
              <IconButton onClick={handleRefreshPrices} disabled={refreshingPrices} title="Refresh market prices">
                <RefreshIcon />
              </IconButton>
            </Box>
          </Box>

          {/* Search Bar and Filter */}
          <Box sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
          <TextField
            fullWidth
            variant="outlined"
            placeholder="Search items..."
            size="small"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon fontSize="small" />
                </InputAdornment>
              ),
            }}
          />
          <FormControlLabel
            control={
              <Switch
                checked={showBelowTargetOnly}
                onChange={(e) => setShowBelowTargetOnly(e.target.checked)}
              />
            }
            label="Below target only"
            sx={{ whiteSpace: 'nowrap' }}
          />
        </Box>
        </Box>

        {/* Tree View - Visible Stations */}
        <Card>
          <CardContent>
            <List dense>
              {visibleStructures.map((structure) => {
                const structureNodeId = `structure-${structure.id}`;
                const isStructureExpanded = expandedNodes.has(structureNodeId);

                return (
                  <Box key={structure.id}>
                    {/* Station/Structure Node */}
                    <ListItemButton onClick={() => toggleNode(structureNodeId)} sx={{ pl: 0 }}>
                      {isStructureExpanded ? <ExpandLess /> : <ExpandMore />}
                      <LocationOnIcon sx={{ mr: 1, color: 'primary.main' }} />
                      <ListItemText
                        primary={structure.name}
                        secondary={`${structure.solarSystem} · ${structure.region}`}
                        primaryTypographyProps={{ fontWeight: 700, variant: 'body1' }}
                      />
                      <IconButton
                        size="small"
                        onClick={(e) => {
                          e.stopPropagation();
                          togglePinStructure(structure.id);
                        }}
                        sx={{ ml: 1 }}
                      >
                        {pinnedStructures.has(structure.id) ? (
                          <PushPinIcon fontSize="small" color="primary" />
                        ) : (
                          <PushPinOutlinedIcon fontSize="small" />
                        )}
                      </IconButton>
                      <IconButton
                        size="small"
                        onClick={(e) => {
                          e.stopPropagation();
                          toggleHideStructure(structure.id);
                        }}
                        sx={{ ml: 1 }}
                      >
                        <VisibilityOffIcon fontSize="small" />
                      </IconButton>
                    </ListItemButton>

                    <Collapse in={isStructureExpanded} timeout={150} unmountOnExit>
                      <List component="div" disablePadding dense>
                        {/* Personal Hangar */}
                        {structure.hangarAssets && structure.hangarAssets.length > 0 && (
                          <Box>
                            <ListItemButton onClick={() => toggleNode(`structure-${structure.id}-hangar`)} sx={{ pl: 2 }}>
                              {expandedNodes.has(`structure-${structure.id}-hangar`) ? <ExpandLess /> : <ExpandMore />}
                              <ListItemText
                                primary="Personal Hangar"
                                secondary={`${structure.hangarAssets.length} items`}
                                primaryTypographyProps={{ fontWeight: 600 }}
                              />
                            </ListItemButton>
                            <Collapse in={expandedNodes.has(`structure-${structure.id}-hangar`)} timeout={150} unmountOnExit>
                              {renderAssetsTable(structure.hangarAssets, true, structure.id)}
                            </Collapse>
                          </Box>
                        )}

                        {/* Personal Hangar Containers */}
                        {structure.hangarContainers && structure.hangarContainers.length > 0 &&
                          structure.hangarContainers.map((container) =>
                            renderContainer(container, `structure-${structure.id}`, true, structure.id)
                          )}

                        {/* Deliveries */}
                        {structure.deliveries && structure.deliveries.length > 0 && (
                          <Box>
                            <ListItemButton onClick={() => toggleNode(`structure-${structure.id}-deliveries`)} sx={{ pl: 2 }}>
                              {expandedNodes.has(`structure-${structure.id}-deliveries`) ? <ExpandLess /> : <ExpandMore />}
                              <ListItemText
                                primary="📬 Deliveries"
                                secondary={`${structure.deliveries.length} items`}
                                primaryTypographyProps={{ fontWeight: 600 }}
                              />
                            </ListItemButton>
                            <Collapse in={expandedNodes.has(`structure-${structure.id}-deliveries`)} timeout={150} unmountOnExit>
                              {renderAssetsTable(structure.deliveries, true, structure.id)}
                            </Collapse>
                          </Box>
                        )}

                        {/* Asset Safety */}
                        {structure.assetSafety && structure.assetSafety.length > 0 && (
                          <Box>
                            <ListItemButton onClick={() => toggleNode(`structure-${structure.id}-safety`)} sx={{ pl: 2 }}>
                              {expandedNodes.has(`structure-${structure.id}-safety`) ? <ExpandLess /> : <ExpandMore />}
                              <ListItemText
                                primary="🛡️ Asset Safety"
                                secondary={`${structure.assetSafety.length} items`}
                                primaryTypographyProps={{ fontWeight: 600 }}
                              />
                            </ListItemButton>
                            <Collapse in={expandedNodes.has(`structure-${structure.id}-safety`)} timeout={150} unmountOnExit>
                              {renderAssetsTable(structure.assetSafety, true, structure.id)}
                            </Collapse>
                          </Box>
                        )}

                        {/* Corporation Hangars */}
                        {structure.corporationHangers && structure.corporationHangers.length > 0 &&
                          structure.corporationHangers
                            .sort((a, b) => a.id - b.id)
                            .map((hanger) =>
                              renderCorporationHanger(hanger, structure.id)
                            )}
                      </List>
                    </Collapse>
                  </Box>
                );
              })}
            </List>
          </CardContent>
        </Card>

        {/* Hidden Stations */}
        {hiddenStructuresList.length > 0 && (
          <Card sx={{ mt: 2 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <VisibilityOffIcon /> Hidden Stations ({hiddenStructuresList.length})
              </Typography>
              <List dense>
                {hiddenStructuresList.map((structure) => {
                  const structureNodeId = `structure-${structure.id}`;
                  const isStructureExpanded = expandedNodes.has(structureNodeId);

                  return (
                    <Box key={structure.id}>
                      {/* Hidden Station/Structure Node */}
                      <ListItemButton onClick={() => toggleNode(structureNodeId)} sx={{ pl: 0, opacity: 0.6 }}>
                        {isStructureExpanded ? <ExpandLess /> : <ExpandMore />}
                        <LocationOnIcon sx={{ mr: 1, color: 'text.secondary' }} />
                        <ListItemText
                          primary={structure.name}
                          secondary={`${structure.solarSystem} · ${structure.region}`}
                          primaryTypographyProps={{ fontWeight: 700, variant: 'body1' }}
                        />
                        <IconButton
                          size="small"
                          onClick={(e) => {
                            e.stopPropagation();
                            togglePinStructure(structure.id);
                          }}
                          sx={{ ml: 1 }}
                        >
                          {pinnedStructures.has(structure.id) ? (
                            <PushPinIcon fontSize="small" color="primary" />
                          ) : (
                            <PushPinOutlinedIcon fontSize="small" />
                          )}
                        </IconButton>
                        <IconButton
                          size="small"
                          onClick={(e) => {
                            e.stopPropagation();
                            toggleHideStructure(structure.id);
                          }}
                          sx={{ ml: 1 }}
                        >
                          <VisibilityIcon fontSize="small" />
                        </IconButton>
                      </ListItemButton>

                      <Collapse in={isStructureExpanded} timeout={150} unmountOnExit>
                        <List component="div" disablePadding dense>
                          {/* Personal Hangar */}
                          {structure.hangarAssets && structure.hangarAssets.length > 0 && (
                            <Box>
                              <ListItemButton onClick={() => toggleNode(`structure-${structure.id}-hangar`)} sx={{ pl: 2 }}>
                                {expandedNodes.has(`structure-${structure.id}-hangar`) ? <ExpandLess /> : <ExpandMore />}
                                <ListItemText
                                  primary="Personal Hangar"
                                  secondary={`${structure.hangarAssets.length} items`}
                                  primaryTypographyProps={{ fontWeight: 600 }}
                                />
                              </ListItemButton>
                              <Collapse in={expandedNodes.has(`structure-${structure.id}-hangar`)} timeout={150} unmountOnExit>
                                {renderAssetsTable(structure.hangarAssets, true, structure.id)}
                              </Collapse>
                            </Box>
                          )}

                          {/* Personal Hangar Containers */}
                          {structure.hangarContainers && structure.hangarContainers.length > 0 &&
                            structure.hangarContainers.map((container) =>
                              renderContainer(container, `structure-${structure.id}`, true, structure.id)
                            )}

                          {/* Deliveries */}
                          {structure.deliveries && structure.deliveries.length > 0 && (
                            <Box>
                              <ListItemButton onClick={() => toggleNode(`structure-${structure.id}-deliveries`)} sx={{ pl: 2 }}>
                                {expandedNodes.has(`structure-${structure.id}-deliveries`) ? <ExpandLess /> : <ExpandMore />}
                                <ListItemText
                                  primary="📬 Deliveries"
                                  secondary={`${structure.deliveries.length} items`}
                                  primaryTypographyProps={{ fontWeight: 600 }}
                                />
                              </ListItemButton>
                              <Collapse in={expandedNodes.has(`structure-${structure.id}-deliveries`)} timeout={150} unmountOnExit>
                                {renderAssetsTable(structure.deliveries, true, structure.id)}
                              </Collapse>
                            </Box>
                          )}

                          {/* Asset Safety */}
                          {structure.assetSafety && structure.assetSafety.length > 0 && (
                            <Box>
                              <ListItemButton onClick={() => toggleNode(`structure-${structure.id}-safety`)} sx={{ pl: 2 }}>
                                {expandedNodes.has(`structure-${structure.id}-safety`) ? <ExpandLess /> : <ExpandMore />}
                                <ListItemText
                                  primary="🛡️ Asset Safety"
                                  secondary={`${structure.assetSafety.length} items`}
                                  primaryTypographyProps={{ fontWeight: 600 }}
                                />
                              </ListItemButton>
                              <Collapse in={expandedNodes.has(`structure-${structure.id}-safety`)} timeout={150} unmountOnExit>
                                {renderAssetsTable(structure.assetSafety, true, structure.id)}
                              </Collapse>
                            </Box>
                          )}

                          {/* Corporation Hangars */}
                          {structure.corporationHangers && structure.corporationHangers.length > 0 &&
                            structure.corporationHangers
                              .sort((a, b) => a.id - b.id)
                              .map((hanger) =>
                                renderCorporationHanger(hanger, structure.id)
                              )}
                        </List>
                      </Collapse>
                    </Box>
                  );
                })}
              </List>
            </CardContent>
          </Card>
        )}

        {filteredStructures.length === 0 && searchQuery && (
          <Card>
            <CardContent>
              <Typography variant="body1" color="text.secondary" textAlign="center">
                No items found matching "{searchQuery}"
              </Typography>
            </CardContent>
          </Card>
        )}

        {/* Stockpile Modal */}
        <Dialog
          open={stockpileModalOpen}
          onClose={() => setStockpileModalOpen(false)}
          maxWidth="sm"
          fullWidth
          TransitionProps={{
            onEntered: () => {
              desiredQuantityInputRef.current?.focus();
            }
          }}
        >
          <DialogTitle>
            {selectedAsset?.asset.desiredQuantity ? 'Edit' : 'Set'} Stockpile Marker
          </DialogTitle>
          <DialogContent>
            <Box sx={{ pt: 1 }}>
              <Typography variant="body2" gutterBottom>
                <strong>Item:</strong> {selectedAsset?.asset.name}
              </Typography>
              <Typography variant="body2" gutterBottom sx={{ mb: 2 }}>
                <strong>Current Quantity:</strong> {selectedAsset?.asset.quantity.toLocaleString()}
              </Typography>
              <TextField
                fullWidth
                label="Desired Quantity"
                type="text"
                value={desiredQuantity}
                onChange={(e) => handleQuantityChange(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && desiredQuantity) {
                    e.preventDefault();
                    handleSaveStockpile();
                  }
                }}
                sx={{ mb: 2 }}
                required
                placeholder="0"
                inputRef={desiredQuantityInputRef}
              />
              <TextField
                fullWidth
                label="Notes (optional)"
                multiline
                rows={3}
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
              />
            </Box>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setStockpileModalOpen(false)}>Cancel</Button>
            <Button onClick={handleSaveStockpile} variant="contained" disabled={!desiredQuantity}>
              Save
            </Button>
          </DialogActions>
        </Dialog>
      </Container>
    </>
  );
}
