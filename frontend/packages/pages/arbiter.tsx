import Head from "next/head";
import { useSession } from "next-auth/react";
import { useState, useCallback, useEffect, useRef, useMemo } from "react";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import Navbar from "@industry-tool/components/Navbar";
import { formatISK, formatDuration } from "@industry-tool/utils/formatting";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import {
  Collapsible,
  CollapsibleTrigger,
  CollapsibleContent,
} from "@/components/ui/collapsible";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import {
  Tooltip,
  TooltipTrigger,
  TooltipContent,
  TooltipProvider,
} from "@/components/ui/tooltip";
import {
  Loader2,
  ChevronDown,
  ChevronUp,
  ChevronRight,
  Settings2,
  Search,
  X,
  Plus,
  Trash2,
  Copy,
  Check,
  ExternalLink,
} from "lucide-react";
import { cn } from "@/lib/utils";
import Link from "next/link";

// ─── Types ───────────────────────────────────────────────────────────────────

export interface ArbiterSettings {
  reaction_structure: string;
  reaction_rig: string;
  reaction_system_id: number;
  reaction_system_name: string;
  invention_structure: string;
  invention_rig: string;
  invention_system_id: number;
  invention_system_name: string;
  component_structure: string;
  component_rig: string;
  component_system_id: number;
  component_system_name: string;
  final_structure: string;
  final_rig: string;
  final_system_id: number;
  final_system_name: string;
}

export interface TaxProfile {
  sales_tax_rate: number;
  broker_fee_rate: number;
  structure_broker_fee: number;
  input_price_type: "sell" | "buy";
  output_price_type: "sell" | "buy";
  trader_character_id: number | null;
}

export interface ArbiterScope {
  id: number;
  name: string;
}

export interface ListItem {
  type_id: number;
  name: string;
}

export interface DecryptorOption {
  type_id: number;
  name: string;
}

export interface DecryptorResult {
  type_id: number;
  name: string;
  probability_multiplier: number;
  me_modifier: number;
  te_modifier: number;
  run_modifier: number;
  resulting_me: number;
  resulting_runs: number;
  invention_cost: number;
  material_cost: number;
  job_cost: number;
  total_cost: number;
  profit: number;
  roi: number;
  isk_per_day: number;
  build_time_sec: number;
}

export interface Opportunity {
  product_type_id: number;
  product_name: string;
  category: string;
  group: string;
  tech_level: string;
  jita_sell_price: number;
  jita_buy_price: number;
  demand_per_day: number;
  days_of_supply: number;
  duration_sec: number;
  base_runs: number;
  runs: number;
  me: number;
  te: number;
  material_cost: number;
  job_cost: number;
  invention_cost: number;
  total_cost: number;
  revenue: number;
  sales_tax: number;
  broker_fee: number;
  profit: number;
  roi: number;
  best_decryptor: DecryptorResult | null;
  all_decryptors: DecryptorResult[];
  is_blacklisted: boolean;
  is_whitelisted: boolean;
}

export interface OpportunitiesResponse {
  opportunities: Opportunity[];
  generated_at: string;
  total_scanned: number;
  best_character_name: string;
}

export interface BomNode {
  type_id: number;
  name: string;
  quantity: number;
  available: number;
  needed: number;
  delta: number;
  unit_buy_price: number;
  unit_build_cost: number;
  decision: "build" | "buy" | "buy_override" | "build_override";
  children: BomNode[];
  is_blacklisted: boolean;
  is_whitelisted: boolean;
}

export interface SolarSystem {
  solar_system_id: number;
  name: string;
  security_class: string;
  security: number;
}

// ─── Constants ───────────────────────────────────────────────────────────────

const STRUCTURE_OPTIONS = [
  { value: "station", label: "Station" },
  { value: "raitaru", label: "Raitaru" },
  { value: "azbel", label: "Azbel" },
  { value: "sotiyo", label: "Sotiyo" },
  { value: "athanor", label: "Athanor" },
  { value: "tatara", label: "Tatara" },
];

const RIG_OPTIONS = [
  { value: "none", label: "None" },
  { value: "t1", label: "T1" },
  { value: "t2", label: "T2" },
];

const DEFAULT_SETTINGS: ArbiterSettings = {
  reaction_structure: "tatara",
  reaction_rig: "t2",
  reaction_system_id: 30000142,
  reaction_system_name: "Jita",
  invention_structure: "raitaru",
  invention_rig: "t1",
  invention_system_id: 30000142,
  invention_system_name: "Jita",
  component_structure: "raitaru",
  component_rig: "t2",
  component_system_id: 30000142,
  component_system_name: "Jita",
  final_structure: "azbel",
  final_rig: "t2",
  final_system_id: 30000142,
  final_system_name: "Jita",
};

const DEFAULT_TAX_PROFILE: TaxProfile = {
  sales_tax_rate: 8,
  broker_fee_rate: 3,
  structure_broker_fee: 1,
  input_price_type: "sell",
  output_price_type: "buy",
  trader_character_id: null,
};

type SortField =
  | "profit"
  | "roi"
  | "demand_per_day"
  | "days_of_supply"
  | "material_cost"
  | "revenue"
  | "sales_tax";

// ─── Security color helper ────────────────────────────────────────────────────

function getSecurityColor(security: number): string {
  if (security >= 0.5) return "#4caf50"; // high sec - green
  if (security > 0) return "#ff9800"; // low sec - yellow
  if (security === 0) return "#9c27b0"; // wormhole / null 0.0 - purple
  return "#f44336"; // null sec - red
}

function getSecurityLabel(security: number): string {
  if (security >= 0.5) return "High";
  if (security > 0) return "Low";
  if (security === 0) return "W-Space";
  return "Null";
}

// ─── SystemSearch ─────────────────────────────────────────────────────────────

interface SystemSearchProps {
  value: string;
  onChange: (system: SolarSystem) => void;
  placeholder?: string;
}

function SystemSearch({ value, onChange, placeholder }: SystemSearchProps) {
  const [query, setQuery] = useState(value);
  const [results, setResults] = useState<SolarSystem[]>([]);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // sync external value changes
  useEffect(() => {
    setQuery(value);
  }, [value]);

  useEffect(() => {
    if (!query || query === value) {
      setResults([]);
      return;
    }
    const t = setTimeout(async () => {
      setLoading(true);
      try {
        const res = await fetch(
          `/api/solar-systems/search?q=${encodeURIComponent(query)}&limit=10`,
        );
        if (res.ok) {
          setResults(await res.json());
        }
      } catch {
        setResults([]);
      } finally {
        setLoading(false);
      }
    }, 300);
    return () => clearTimeout(t);
  }, [query, value]);

  // close on outside click
  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  return (
    <div className="relative" ref={containerRef}>
      <div className="relative">
        <Input
          className="h-8 text-sm pr-7"
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setOpen(true);
          }}
          onFocus={() => {
            if (results.length > 0) setOpen(true);
          }}
          placeholder={placeholder ?? "Search system..."}
        />
        {loading && (
          <Loader2 className="absolute right-2 top-1/2 -translate-y-1/2 h-3.5 w-3.5 animate-spin text-text-muted" />
        )}
      </div>
      {open && results.length > 0 && (
        <div className="absolute z-50 w-full mt-1 bg-background-elevated border border-overlay-medium rounded shadow-lg max-h-48 overflow-y-auto">
          {results.map((sys) => (
            <div
              key={sys.solar_system_id}
              className="px-3 py-2 cursor-pointer hover:bg-interactive-hover flex items-center gap-2 text-sm"
              onMouseDown={(e) => e.preventDefault()}
              onClick={() => {
                onChange(sys);
                setQuery(sys.name);
                setOpen(false);
                setResults([]);
              }}
            >
              <span
                className="inline-block w-2 h-2 rounded-full flex-shrink-0"
                style={{ backgroundColor: getSecurityColor(sys.security) }}
              />
              <span className="text-text-primary">{sys.name}</span>
              <span
                className="text-xs ml-auto"
                style={{ color: getSecurityColor(sys.security) }}
              >
                {sys.security.toFixed(1)} {getSecurityLabel(sys.security)}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// ─── StructureSection ────────────────────────────────────────────────────────

interface StructureSectionProps {
  title: string;
  prefix: "reaction" | "invention" | "component" | "final";
  settings: ArbiterSettings;
  onChange: (key: keyof ArbiterSettings, value: string | number) => void;
}

function StructureSection({
  title,
  prefix,
  settings,
  onChange,
}: StructureSectionProps) {
  const structureKey = `${prefix}_structure` as keyof ArbiterSettings;
  const rigKey = `${prefix}_rig` as keyof ArbiterSettings;
  const systemNameKey = `${prefix}_system_name` as keyof ArbiterSettings;
  const systemIdKey = `${prefix}_system_id` as keyof ArbiterSettings;

  return (
    <div className="flex-1 min-w-[200px] space-y-3">
      <h3 className="text-sm font-semibold text-text-heading border-b border-overlay-subtle pb-1">
        {title}
      </h3>

      <div className="space-y-1">
        <Label className="text-xs text-text-secondary">Structure</Label>
        <Select
          value={settings[structureKey] as string}
          onValueChange={(v) => onChange(structureKey, v)}
        >
          <SelectTrigger className="h-8 text-sm">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {STRUCTURE_OPTIONS.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-1">
        <Label className="text-xs text-text-secondary">Rig</Label>
        <Select
          value={settings[rigKey] as string}
          onValueChange={(v) => onChange(rigKey, v)}
        >
          <SelectTrigger className="h-8 text-sm">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {RIG_OPTIONS.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-1">
        <Label className="text-xs text-text-secondary">System</Label>
        <SystemSearch
          value={settings[systemNameKey] as string}
          onChange={(sys) => {
            onChange(systemNameKey, sys.name);
            onChange(systemIdKey, sys.solar_system_id);
          }}
        />
      </div>
    </div>
  );
}

// ─── BomTreeRow ───────────────────────────────────────────────────────────────

interface BomTreeRowProps {
  node: BomNode;
  depth: number;
  onContextMenu: (e: React.MouseEvent, typeId: number, name: string) => void;
}

function getDecisionBadgeClasses(decision: BomNode["decision"]): string {
  switch (decision) {
    case "build":
      return "bg-teal-success/20 text-teal-success border-teal-success/30";
    case "buy":
      return "bg-overlay-subtle text-text-secondary border-overlay-medium";
    case "build_override":
      return "bg-teal-success/30 text-teal-success border-teal-success/40";
    case "buy_override":
      return "bg-rose-danger/20 text-rose-danger border-rose-danger/30";
    default:
      return "bg-overlay-subtle text-text-secondary border-overlay-medium";
  }
}

function getDecisionLabel(decision: BomNode["decision"]): string {
  switch (decision) {
    case "build":
      return "BUILD";
    case "buy":
      return "BUY";
    case "build_override":
      return "BUILD*";
    case "buy_override":
      return "BUY*";
    default:
      return decision;
  }
}

function BomTreeRow({ node, depth, onContextMenu }: BomTreeRowProps) {
  const [expanded, setExpanded] = useState(false);
  const hasChildren = node.children && node.children.length > 0;
  const isBuy = node.decision === "buy" || node.decision === "buy_override";

  return (
    <>
      <TableRow
        className={cn(
          "cursor-pointer hover:bg-interactive-hover group",
          node.is_blacklisted && "bg-rose-danger/5",
          node.is_whitelisted && "bg-teal-success/5",
          isBuy && depth > 0 && "opacity-60",
        )}
        onClick={() => hasChildren && setExpanded((v) => !v)}
        onContextMenu={(e) => onContextMenu(e, node.type_id, node.name)}
      >
        <TableCell className="py-1.5" style={{ paddingLeft: `${16 + depth * 20}px` }}>
          <span className="flex items-center gap-2 text-sm">
            {hasChildren ? (
              expanded ? (
                <ChevronDown className="h-3.5 w-3.5 text-text-muted flex-shrink-0" />
              ) : (
                <ChevronRight className="h-3.5 w-3.5 text-text-muted flex-shrink-0" />
              )
            ) : (
              <span className="w-3.5 flex-shrink-0" />
            )}
            <img
              src={`https://images.evetech.net/types/${node.type_id}/icon?size=32`}
              alt={node.name}
              className="w-5 h-5 rounded flex-shrink-0"
              onError={(e) => {
                (e.target as HTMLImageElement).style.display = "none";
              }}
            />
            <span className="text-text-primary">{node.name}</span>
          </span>
        </TableCell>
        <TableCell className="py-1.5 text-sm text-text-data-value font-mono text-right">
          {node.quantity.toLocaleString()}
        </TableCell>
        <TableCell className="py-1.5 text-sm text-right">
          <span
            className={cn(
              "font-mono",
              node.available >= node.needed
                ? "text-teal-success"
                : node.available > 0
                ? "text-amber-manufacturing"
                : "text-rose-danger",
            )}
          >
            {node.available.toLocaleString()}
          </span>
        </TableCell>
        <TableCell className="py-1.5 text-sm text-text-muted font-mono text-right">
          {node.needed.toLocaleString()}
        </TableCell>
        <TableCell className="py-1.5 text-sm text-right">
          <span
            className={cn(
              "font-mono",
              node.delta >= 0 ? "text-text-muted" : "text-rose-danger",
            )}
          >
            {node.delta >= 0 ? "0" : node.delta.toLocaleString()}
          </span>
        </TableCell>
        <TableCell className="py-1.5 text-sm text-text-data-value font-mono text-right">
          {formatISK(node.unit_buy_price)}
        </TableCell>
        <TableCell className="py-1.5">
          <Badge
            className={cn(
              "text-[10px] px-1.5 py-0",
              getDecisionBadgeClasses(node.decision),
            )}
          >
            {getDecisionLabel(node.decision)}
          </Badge>
        </TableCell>
      </TableRow>
      {expanded &&
        node.children.map((child, i) => (
          <BomTreeRow
            key={`${child.type_id}-${i}`}
            node={child}
            depth={depth + 1}
            onContextMenu={onContextMenu}
          />
        ))}
    </>
  );
}

// ─── ContextMenu ─────────────────────────────────────────────────────────────

interface ContextMenuState {
  x: number;
  y: number;
  typeId: number;
  name: string;
}

interface ContextMenuProps {
  menu: ContextMenuState;
  blacklist: ListItem[];
  whitelist: ListItem[];
  onClose: () => void;
  onAddBlacklist: (typeId: number, name: string) => void;
  onRemoveBlacklist: (typeId: number) => void;
  onAddWhitelist: (typeId: number, name: string) => void;
  onRemoveWhitelist: (typeId: number) => void;
}

function ContextMenu({
  menu,
  blacklist,
  whitelist,
  onClose,
  onAddBlacklist,
  onRemoveBlacklist,
  onAddWhitelist,
  onRemoveWhitelist,
}: ContextMenuProps) {
  const isBlacklisted = blacklist.some((i) => i.type_id === menu.typeId);
  const isWhitelisted = whitelist.some((i) => i.type_id === menu.typeId);

  useEffect(() => {
    const handler = () => onClose();
    document.addEventListener("click", handler);
    document.addEventListener("keydown", handler);
    return () => {
      document.removeEventListener("click", handler);
      document.removeEventListener("keydown", handler);
    };
  }, [onClose]);

  return (
    <div
      className="fixed z-[9999] bg-background-elevated border border-overlay-medium rounded shadow-lg py-1 min-w-[200px]"
      style={{ top: menu.y, left: menu.x }}
      onClick={(e) => e.stopPropagation()}
    >
      <div className="px-3 py-1.5 text-xs text-text-muted font-medium border-b border-overlay-subtle mb-1">
        {menu.name}
      </div>
      {isWhitelisted ? (
        <button
          className="w-full text-left px-3 py-1.5 text-sm text-text-primary hover:bg-interactive-hover"
          onClick={() => {
            onRemoveWhitelist(menu.typeId);
            onClose();
          }}
        >
          Remove from Whitelist
        </button>
      ) : (
        <button
          className="w-full text-left px-3 py-1.5 text-sm text-teal-success hover:bg-interactive-hover"
          onClick={() => {
            onAddWhitelist(menu.typeId, menu.name);
            onClose();
          }}
        >
          Add to Whitelist
        </button>
      )}
      {isBlacklisted ? (
        <button
          className="w-full text-left px-3 py-1.5 text-sm text-text-primary hover:bg-interactive-hover"
          onClick={() => {
            onRemoveBlacklist(menu.typeId);
            onClose();
          }}
        >
          Remove from Blacklist
        </button>
      ) : (
        <button
          className="w-full text-left px-3 py-1.5 text-sm text-rose-danger hover:bg-interactive-hover"
          onClick={() => {
            onAddBlacklist(menu.typeId, menu.name);
            onClose();
          }}
        >
          Add to Blacklist
        </button>
      )}
    </div>
  );
}

// ─── OpportunityRow ───────────────────────────────────────────────────────────

interface OpportunityRowProps {
  opp: Opportunity;
  rank: number;
  qty: number;
  onQtyChange: (typeId: number, qty: number) => void;
  scopeId: number | null;
  onContextMenu: (e: React.MouseEvent, typeId: number, name: string) => void;
  onSelect: (opp: Opportunity) => void;
  isSelected: boolean;
  inputPrice: "sell" | "buy";
  outputPrice: "sell" | "buy";
}

function OpportunityRow({
  opp,
  rank,
  qty,
  onQtyChange,
  scopeId,
  onContextMenu,
  onSelect,
  isSelected,
  inputPrice,
  outputPrice,
}: OpportunityRowProps) {
  const [expanded, setExpanded] = useState(false);
  const [bom, setBom] = useState<BomNode | null>(null);
  const [bomLoading, setBomLoading] = useState(false);
  const [buildAll, setBuildAll] = useState(false);
  const bomFetched = useRef(false);

  const effectiveQty = qty || 1;

  // Recalculate values based on qty
  const scaledCost = opp.total_cost * effectiveQty;
  const scaledRevenue = opp.revenue * effectiveQty;
  const scaledProfit = opp.profit * effectiveQty;
  const scaledSalesTax = opp.sales_tax * effectiveQty;

  async function fetchBom() {
    if (bomFetched.current) return;
    bomFetched.current = true;
    setBomLoading(true);
    try {
      const params = new URLSearchParams();
      if (scopeId) params.set("scope_id", String(scopeId));
      params.set("quantity", String(effectiveQty));
      params.set("build_all", String(buildAll));
      const res = await fetch(
        `/api/arbiter/${opp.product_type_id}/bom?${params.toString()}`,
      );
      if (res.ok) {
        setBom(await res.json());
      }
    } catch {
      // ignore
    } finally {
      setBomLoading(false);
    }
  }

  function handleExpand() {
    setExpanded((v) => {
      if (!v) fetchBom();
      return !v;
    });
    onSelect(opp);
  }

  async function handleBuildAllToggle() {
    const newBuildAll = !buildAll;
    setBuildAll(newBuildAll);
    setBomLoading(true);
    try {
      const params = new URLSearchParams();
      if (scopeId) params.set("scope_id", String(scopeId));
      params.set("quantity", String(effectiveQty));
      params.set("build_all", String(newBuildAll));
      const res = await fetch(
        `/api/arbiter/${opp.product_type_id}/bom?${params.toString()}`,
      );
      if (res.ok) {
        setBom(await res.json());
      }
    } catch {
      // ignore
    } finally {
      setBomLoading(false);
    }
  }

  const revenue = outputPrice === "sell" ? opp.jita_sell_price : opp.jita_buy_price;

  return (
    <>
      <TableRow
        className={cn(
          "cursor-pointer hover:bg-interactive-hover",
          opp.is_blacklisted && "bg-rose-danger/5 hover:bg-rose-danger/10",
          opp.is_whitelisted && "bg-teal-success/5 hover:bg-teal-success/10",
          isSelected && "bg-interactive-selected",
        )}
        onClick={handleExpand}
        onContextMenu={(e) =>
          onContextMenu(e, opp.product_type_id, opp.product_name)
        }
      >
        <TableCell className="py-2 text-xs text-text-muted w-8">{rank}</TableCell>
        <TableCell className="py-2">
          <span className="flex items-center gap-2 text-sm font-medium text-text-primary">
            {expanded ? (
              <ChevronDown className="h-3.5 w-3.5 text-text-muted flex-shrink-0" />
            ) : (
              <ChevronRight className="h-3.5 w-3.5 text-text-muted flex-shrink-0" />
            )}
            <img
              src={`https://images.evetech.net/types/${opp.product_type_id}/icon?size=32`}
              alt={opp.product_name}
              className="w-5 h-5 rounded flex-shrink-0"
              onError={(e) => {
                (e.target as HTMLImageElement).style.display = "none";
              }}
            />
            {opp.product_name}
          </span>
        </TableCell>
        <TableCell className="py-2 text-sm text-text-secondary">
          {formatDuration(opp.duration_sec)}
        </TableCell>
        <TableCell className="py-2 text-sm text-text-secondary font-mono">
          {opp.demand_per_day.toFixed(1)}/d
        </TableCell>
        <TableCell className="py-2">
          <span
            className={cn(
              "text-sm font-mono",
              opp.days_of_supply < 7
                ? "text-rose-danger"
                : opp.days_of_supply < 30
                ? "text-amber-manufacturing"
                : "text-teal-success",
            )}
          >
            {opp.days_of_supply.toFixed(1)}d
          </span>
        </TableCell>
        <TableCell
          className="py-2"
          onClick={(e) => e.stopPropagation()}
        >
          <Input
            type="number"
            min={1}
            className="h-7 w-16 text-xs text-center"
            value={effectiveQty}
            onChange={(e) =>
              onQtyChange(opp.product_type_id, Math.max(1, parseInt(e.target.value) || 1))
            }
          />
        </TableCell>
        <TableCell className="py-2">
          <span
            className={cn(
              "text-sm font-mono",
              opp.roi >= 0 ? "text-teal-success" : "text-rose-danger",
            )}
          >
            {opp.roi.toFixed(1)}%
          </span>
        </TableCell>
        <TableCell className="py-2 text-sm text-text-data-value font-mono">
          {formatISK(scaledCost)}
        </TableCell>
        <TableCell className="py-2 text-sm text-text-data-value font-mono">
          {formatISK(scaledRevenue)}
        </TableCell>
        <TableCell className="py-2 text-sm text-text-muted font-mono">
          {formatISK(scaledSalesTax)}
        </TableCell>
        <TableCell className="py-2">
          <span
            className={cn(
              "text-sm font-mono",
              scaledProfit >= 0 ? "text-teal-success" : "text-rose-danger",
            )}
          >
            {formatISK(scaledProfit)}
          </span>
        </TableCell>
        <TableCell className="py-2 text-sm text-text-secondary font-mono">
          {opp.me}
        </TableCell>
        <TableCell className="py-2 text-sm text-text-secondary font-mono">
          {opp.te}
        </TableCell>
        <TableCell className="py-2 text-sm text-text-secondary font-mono">
          {opp.runs}
        </TableCell>
        <TableCell className="py-2 text-xs text-text-secondary truncate max-w-[100px]">
          {opp.best_decryptor?.name ?? "None"}
        </TableCell>
        <TableCell className="py-2 text-xs text-text-secondary">
          {opp.group}
        </TableCell>
        <TableCell className="py-2">
          <Badge
            className={cn(
              "text-[10px] px-1.5 py-0",
              opp.category === "ship"
                ? "bg-blue-science/20 text-blue-science border-blue-science/30"
                : opp.category === "module"
                ? "bg-category-violet/20 text-category-violet border-category-violet/30"
                : "bg-overlay-subtle text-text-secondary border-overlay-medium",
            )}
          >
            {opp.category}
          </Badge>
        </TableCell>
      </TableRow>

      {expanded && (
        <TableRow className="bg-background-panel hover:bg-background-panel">
          <TableCell colSpan={17} className="p-0">
            <div className="border-t border-overlay-subtle">
              {/* BOM header */}
              <div className="flex items-center justify-between px-4 py-2 border-b border-overlay-subtle">
                <span className="text-xs text-text-muted">
                  Bill of Materials — {opp.product_name}
                </span>
                <div className="flex items-center gap-3">
                  <Label className="flex items-center gap-2 text-xs text-text-secondary cursor-pointer">
                    <Switch
                      checked={buildAll}
                      onCheckedChange={handleBuildAllToggle}
                    />
                    Build All
                  </Label>
                </div>
              </div>
              {bomLoading ? (
                <div className="flex justify-center py-6">
                  <Loader2 className="h-5 w-5 animate-spin text-primary" />
                </div>
              ) : bom ? (
                <Table>
                  <TableHeader>
                    <TableRow className="hover:bg-transparent border-overlay-subtle">
                      <TableHead className="py-1.5 text-xs text-text-muted">
                        Item
                      </TableHead>
                      <TableHead className="py-1.5 text-xs text-text-muted text-right">
                        Qty
                      </TableHead>
                      <TableHead className="py-1.5 text-xs text-text-muted text-right">
                        Available
                      </TableHead>
                      <TableHead className="py-1.5 text-xs text-text-muted text-right">
                        Needed
                      </TableHead>
                      <TableHead className="py-1.5 text-xs text-text-muted text-right">
                        Delta
                      </TableHead>
                      <TableHead className="py-1.5 text-xs text-text-muted text-right">
                        Unit Price
                      </TableHead>
                      <TableHead className="py-1.5 text-xs text-text-muted">
                        Decision
                      </TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    <BomTreeRow
                      node={bom}
                      depth={0}
                      onContextMenu={onContextMenu}
                    />
                  </TableBody>
                </Table>
              ) : (
                <div className="flex justify-center py-4">
                  <p className="text-sm text-text-muted">No BOM data</p>
                </div>
              )}
            </div>
          </TableCell>
        </TableRow>
      )}
    </>
  );
}

// ─── ShoppingList ─────────────────────────────────────────────────────────────

interface ShoppingListItem {
  type_id: number;
  name: string;
  req_qty: number;
  unit_price: number;
  total_value: number;
  warehouse: number;
  delta_qty: number;
  delta_cost: number;
}

// Collect all leaf-level nodes that require purchasing (items to buy from market)
function collectShoppingItems(node: BomNode): ShoppingListItem[] {
  if (!node) return [];
  const result: ShoppingListItem[] = [];

  if (node.decision === "buy" || node.decision === "buy_override") {
    // This item is being bought — add to shopping list if there's a delta
    if (node.delta > 0) {
      result.push({
        type_id: node.type_id,
        name: node.name,
        req_qty: Number(node.quantity),
        unit_price: node.unit_buy_price,
        total_value: node.unit_buy_price * Number(node.quantity),
        warehouse: Number(node.available),
        delta_qty: Number(node.delta),
        delta_cost: node.unit_buy_price * Number(node.delta),
      });
    }
  } else if (node.children && node.children.length > 0) {
    // Building this item — recurse into its children
    for (const child of node.children) {
      result.push(...collectShoppingItems(child));
    }
  }

  return result;
}

interface WarehousePanelProps {
  selectedOpp: Opportunity | null;
  scopeId: number | null;
  qty: number;
  inputPrice: "sell" | "buy";
}

function WarehousePanel({
  selectedOpp,
  scopeId,
  qty,
  inputPrice,
}: WarehousePanelProps) {
  const [copied, setCopied] = useState(false);
  const [bom, setBom] = useState<BomNode | null>(null);
  const [bomLoading, setBomLoading] = useState(false);

  useEffect(() => {
    if (!selectedOpp) {
      setBom(null);
      return;
    }
    setBomLoading(true);
    const params = new URLSearchParams();
    if (scopeId) params.set("scope_id", String(scopeId));
    params.set("quantity", String(qty || 1));
    fetch(`/api/arbiter/${selectedOpp.product_type_id}/bom?${params.toString()}`)
      .then((r) => (r.ok ? r.json() : null))
      .then((data) => setBom(data))
      .catch(() => setBom(null))
      .finally(() => setBomLoading(false));
  }, [selectedOpp?.product_type_id, scopeId, qty]);

  const ingredients = bom?.children ?? [];

  const shoppingItems = useMemo<ShoppingListItem[]>(() => {
    if (!bom) return [];
    return collectShoppingItems(bom);
  }, [bom]);

  const totalCost = shoppingItems.reduce((s, i) => s + i.delta_cost, 0);

  function handleExportMultibuy() {
    const lines = shoppingItems
      .filter((i) => i.delta_qty > 0)
      .map((i) => `${i.name} x ${i.delta_qty}`);
    navigator.clipboard.writeText(lines.join("\n")).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  }

  return (
    <div className="border-t border-overlay-subtle bg-background-panel h-[300px] flex">
      {/* Left: Ingredients */}
      <div className="w-64 border-r border-overlay-subtle flex flex-col">
        <div className="px-3 py-2 border-b border-overlay-subtle flex-shrink-0">
          <span className="text-xs font-medium text-text-heading">Ingredients</span>
        </div>
        <div className="flex-1 overflow-y-auto">
          {bomLoading ? (
            <div className="flex items-center justify-center h-full">
              <Loader2 className="h-4 w-4 animate-spin text-text-muted" />
            </div>
          ) : selectedOpp ? (
            ingredients.length > 0 ? (
              <table className="w-full text-xs">
                <thead>
                  <tr className="border-b border-overlay-subtle">
                    <th className="px-3 py-1 text-left text-text-muted font-normal">
                      Item
                    </th>
                    <th className="px-2 py-1 text-right text-text-muted font-normal">
                      Avail
                    </th>
                    <th className="px-2 py-1 text-right text-text-muted font-normal">
                      Need
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {ingredients.map((item) => (
                    <tr
                      key={item.type_id}
                      className="border-b border-overlay-subtle/50 hover:bg-interactive-hover"
                    >
                      <td className="px-3 py-1 text-text-primary truncate max-w-[120px]">
                        {item.name}
                      </td>
                      <td className="px-2 py-1 text-right font-mono">
                        <span
                          className={
                            item.available >= item.needed
                              ? "text-teal-success"
                              : item.available > 0
                              ? "text-amber-manufacturing"
                              : "text-rose-danger"
                          }
                        >
                          {Number(item.available).toLocaleString()}
                        </span>
                      </td>
                      <td className="px-2 py-1 text-right font-mono text-text-muted">
                        {Number(item.needed).toLocaleString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <div className="flex items-center justify-center h-full">
                <p className="text-xs text-text-muted">
                  No BOM data available
                </p>
              </div>
            )
          ) : (
            <div className="flex items-center justify-center h-full">
              <p className="text-xs text-text-muted">Click a row to select</p>
            </div>
          )}
        </div>
      </div>

      {/* Right: Tabs */}
      <div className="flex-1 flex flex-col min-w-0">
        <Tabs defaultValue="shopping" className="flex flex-col h-full">
          <TabsList className="border-b border-overlay-medium bg-transparent w-full justify-start rounded-none p-0 h-auto flex-shrink-0">
            <TabsTrigger
              value="shopping"
              className="text-text-muted data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none bg-transparent px-4 py-2 text-xs"
            >
              Shopping List
            </TabsTrigger>
            <TabsTrigger
              value="sell"
              className="text-text-muted data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none bg-transparent px-4 py-2 text-xs"
            >
              Sell Orders
            </TabsTrigger>
            <TabsTrigger
              value="buy"
              className="text-text-muted data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none bg-transparent px-4 py-2 text-xs"
            >
              Buy Orders
            </TabsTrigger>
          </TabsList>

          <TabsContent value="shopping" className="flex-1 overflow-y-auto m-0 p-0">
            {shoppingItems.length > 0 ? (
              <>
                {/* Summary row */}
                <div className="flex items-center justify-between px-3 py-1.5 border-b border-overlay-subtle bg-background-elevated">
                  <span className="text-xs text-text-muted">
                    {shoppingItems.filter((i) => i.delta_qty > 0).length} items to buy
                  </span>
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-mono text-text-data-value">
                      {formatISK(totalCost)}
                    </span>
                    <Button
                      size="sm"
                      variant="outline"
                      className="h-6 text-xs gap-1"
                      onClick={handleExportMultibuy}
                    >
                      {copied ? (
                        <Check className="h-3 w-3" />
                      ) : (
                        <Copy className="h-3 w-3" />
                      )}
                      Multibuy
                    </Button>
                  </div>
                </div>
                <table className="w-full text-xs">
                  <thead>
                    <tr className="border-b border-overlay-subtle">
                      <th className="px-3 py-1 text-left text-text-muted font-normal">
                        Item
                      </th>
                      <th className="px-2 py-1 text-right text-text-muted font-normal">
                        Req
                      </th>
                      <th className="px-2 py-1 text-right text-text-muted font-normal">
                        Unit Price
                      </th>
                      <th className="px-2 py-1 text-right text-text-muted font-normal">
                        Total
                      </th>
                      <th className="px-2 py-1 text-right text-text-muted font-normal">
                        WH
                      </th>
                      <th className="px-2 py-1 text-right text-text-muted font-normal">
                        Delta
                      </th>
                      <th className="px-2 py-1 text-right text-text-muted font-normal">
                        Cost
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {shoppingItems.map((item) => (
                      <tr
                        key={item.type_id}
                        className={cn(
                          "border-b border-overlay-subtle/50 hover:bg-interactive-hover",
                          item.delta_qty > 0 && "bg-rose-danger/5",
                        )}
                      >
                        <td className="px-3 py-1 text-text-primary truncate max-w-[160px]">
                          {item.name}
                        </td>
                        <td className="px-2 py-1 text-right font-mono text-text-data-value">
                          {item.req_qty.toLocaleString()}
                        </td>
                        <td className="px-2 py-1 text-right font-mono text-text-muted">
                          {formatISK(item.unit_price)}
                        </td>
                        <td className="px-2 py-1 text-right font-mono text-text-data-value">
                          {formatISK(item.total_value)}
                        </td>
                        <td className="px-2 py-1 text-right font-mono text-text-muted">
                          {item.warehouse.toLocaleString()}
                        </td>
                        <td className="px-2 py-1 text-right font-mono">
                          <span
                            className={
                              item.delta_qty > 0
                                ? "text-rose-danger"
                                : "text-text-muted"
                            }
                          >
                            {item.delta_qty > 0
                              ? `-${item.delta_qty.toLocaleString()}`
                              : "0"}
                          </span>
                        </td>
                        <td className="px-2 py-1 text-right font-mono text-rose-danger">
                          {item.delta_qty > 0 ? formatISK(item.delta_cost) : "—"}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </>
            ) : (
              <div className="flex items-center justify-center h-full">
                <p className="text-xs text-text-muted">
                  {bomLoading
                    ? "Loading..."
                    : selectedOpp
                    ? "Nothing to buy — warehouse is stocked"
                    : "Select an opportunity to see shopping list"}
                </p>
              </div>
            )}
          </TabsContent>

          <TabsContent value="sell" className="flex-1 m-0 p-0">
            <div className="flex items-center justify-center h-full">
              <p className="text-xs text-text-muted">Sell order tracking coming soon</p>
            </div>
          </TabsContent>

          <TabsContent value="buy" className="flex-1 m-0 p-0">
            <div className="flex items-center justify-center h-full">
              <p className="text-xs text-text-muted">Buy order tracking coming soon</p>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}

// ─── ArbiterPage ──────────────────────────────────────────────────────────────

export default function ArbiterPage() {
  const { status } = useSession();

  // Settings state
  const [settings, setSettings] = useState<ArbiterSettings>(DEFAULT_SETTINGS);
  const [taxProfile, setTaxProfile] = useState<TaxProfile>(DEFAULT_TAX_PROFILE);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [settingsTab, setSettingsTab] = useState("structures");
  const [savingSettings, setSavingSettings] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  // Lists state
  const [blacklist, setBlacklist] = useState<ListItem[]>([]);
  const [whitelist, setWhitelist] = useState<ListItem[]>([]);
  const [useBlacklist, setUseBlacklist] = useState(false);
  const [useWhitelist, setUseWhitelist] = useState(false);

  // Scopes state
  const [scopes, setScopes] = useState<ArbiterScope[]>([]);
  const [selectedScopeId, setSelectedScopeId] = useState<number | null>(null);
  const [newScopeName, setNewScopeName] = useState("");
  const [addingScope, setAddingScope] = useState(false);

  // Decryptors state
  const [decryptors, setDecryptors] = useState<DecryptorOption[]>([]);
  const [selectedDecryptorId, setSelectedDecryptorId] = useState<string>("maximize_roi");

  // Price toggles (no-rescan)
  const [inputPrice, setInputPrice] = useState<"sell" | "buy">("sell");
  const [outputPrice, setOutputPrice] = useState<"sell" | "buy">("buy");

  // Opportunities state
  const [opportunities, setOpportunities] = useState<OpportunitiesResponse | null>(null);
  const [scanning, setScanning] = useState(false);
  const [scanError, setScanError] = useState<string | null>(null);

  // Filters
  const [categoryFilters, setCategoryFilters] = useState<Set<string>>(
    new Set(["ship", "module"]),
  );
  const [techFilters, setTechFilters] = useState<Set<string>>(
    new Set(["Tech II"]),
  );
  const [search, setSearch] = useState("");
  const [sortField, setSortField] = useState<SortField>("profit");
  const [sortDesc, setSortDesc] = useState(true);

  // Per-row qty overrides
  const [rowQtys, setRowQtys] = useState<Record<number, number>>({});

  // Selected opportunity for warehouse panel
  const [selectedOpp, setSelectedOpp] = useState<Opportunity | null>(null);

  // Context menu
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);

  // Load initial data
  useEffect(() => {
    if (status !== "authenticated") return;

    // Load settings
    fetch("/api/arbiter/settings")
      .then((r) => r.ok ? r.json() : null)
      .then((data) => {
        if (data) setSettings(data);
      })
      .catch(() => {});

    // Load tax profile
    fetch("/api/arbiter/tax-profile")
      .then((r) => r.ok ? r.json() : null)
      .then((data) => {
        if (data) setTaxProfile(data);
      })
      .catch(() => {});

    // Load scopes
    fetch("/api/arbiter/scopes")
      .then((r) => r.ok ? r.json() : null)
      .then((data) => {
        if (data && Array.isArray(data)) {
          setScopes(data);
          if (data.length > 0) setSelectedScopeId(data[0].id);
        }
      })
      .catch(() => {});

    // Load decryptors
    fetch("/api/arbiter/decryptors")
      .then((r) => r.ok ? r.json() : null)
      .then((data) => {
        if (data && Array.isArray(data)) setDecryptors(data);
      })
      .catch(() => {});

    // Load lists
    fetch("/api/arbiter/blacklist")
      .then((r) => r.ok ? r.json() : null)
      .then((data) => {
        if (data && Array.isArray(data)) setBlacklist(data);
      })
      .catch(() => {});

    fetch("/api/arbiter/whitelist")
      .then((r) => r.ok ? r.json() : null)
      .then((data) => {
        if (data && Array.isArray(data)) setWhitelist(data);
      })
      .catch(() => {});
  }, [status]);

  const handleSettingChange = useCallback(
    (key: keyof ArbiterSettings, value: string | number) => {
      setSettings((prev) => ({ ...prev, [key]: value }));
    },
    [],
  );

  const handleSaveSettings = async () => {
    setSavingSettings(true);
    setSaveError(null);
    try {
      const [settingsRes, taxRes] = await Promise.all([
        fetch("/api/arbiter/settings", {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(settings),
        }),
        fetch("/api/arbiter/tax-profile", {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(taxProfile),
        }),
      ]);
      if (!settingsRes.ok || !taxRes.ok) {
        setSaveError("Failed to save settings");
      } else {
        setSettingsOpen(false);
      }
    } catch {
      setSaveError("Failed to save settings");
    } finally {
      setSavingSettings(false);
    }
  };

  const handleScan = useCallback(async () => {
    setScanning(true);
    setScanError(null);
    try {
      const params = new URLSearchParams();
      if (selectedScopeId) params.set("scope_id", String(selectedScopeId));
      params.set("input_price", inputPrice);
      params.set("output_price", outputPrice);
      if (selectedDecryptorId !== "maximize_roi") {
        params.set("decryptor_type_id", selectedDecryptorId);
      }

      const res = await fetch(`/api/arbiter/opportunities?${params.toString()}`);
      if (!res.ok) {
        const err = await res.text();
        setScanError(err || "Scan failed");
        return;
      }
      const data: OpportunitiesResponse = await res.json();
      setOpportunities(data);
    } catch {
      setScanError("Failed to scan opportunities");
    } finally {
      setScanning(false);
    }
  }, [selectedScopeId, inputPrice, outputPrice, selectedDecryptorId]);

  const handleCreateScope = async () => {
    if (!newScopeName.trim()) return;
    setAddingScope(true);
    try {
      const res = await fetch("/api/arbiter/scopes", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: newScopeName.trim() }),
      });
      if (res.ok) {
        const created = await res.json();
        setScopes((prev) => [...prev, created]);
        setSelectedScopeId(created.id);
        setNewScopeName("");
      }
    } catch {
      // ignore
    } finally {
      setAddingScope(false);
    }
  };

  const handleAddToList = useCallback(
    async (list: "blacklist" | "whitelist", typeId: number, name: string) => {
      try {
        const res = await fetch(`/api/arbiter/${list}`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ type_id: typeId, name }),
        });
        if (res.ok) {
          const item = { type_id: typeId, name };
          if (list === "blacklist") {
            setBlacklist((prev) => [...prev, item]);
            setOpportunities((prev) =>
              prev
                ? {
                    ...prev,
                    opportunities: prev.opportunities.map((o) =>
                      o.product_type_id === typeId
                        ? { ...o, is_blacklisted: true }
                        : o,
                    ),
                  }
                : null,
            );
          } else {
            setWhitelist((prev) => [...prev, item]);
            setOpportunities((prev) =>
              prev
                ? {
                    ...prev,
                    opportunities: prev.opportunities.map((o) =>
                      o.product_type_id === typeId
                        ? { ...o, is_whitelisted: true }
                        : o,
                    ),
                  }
                : null,
            );
          }
        }
      } catch {
        // ignore
      }
    },
    [],
  );

  const handleRemoveFromList = useCallback(
    async (list: "blacklist" | "whitelist", typeId: number) => {
      try {
        const res = await fetch(`/api/arbiter/${list}/${typeId}`, {
          method: "DELETE",
        });
        if (res.ok) {
          if (list === "blacklist") {
            setBlacklist((prev) => prev.filter((i) => i.type_id !== typeId));
            setOpportunities((prev) =>
              prev
                ? {
                    ...prev,
                    opportunities: prev.opportunities.map((o) =>
                      o.product_type_id === typeId
                        ? { ...o, is_blacklisted: false }
                        : o,
                    ),
                  }
                : null,
            );
          } else {
            setWhitelist((prev) => prev.filter((i) => i.type_id !== typeId));
            setOpportunities((prev) =>
              prev
                ? {
                    ...prev,
                    opportunities: prev.opportunities.map((o) =>
                      o.product_type_id === typeId
                        ? { ...o, is_whitelisted: false }
                        : o,
                    ),
                  }
                : null,
            );
          }
        }
      } catch {
        // ignore
      }
    },
    [],
  );

  function handleContextMenu(
    e: React.MouseEvent,
    typeId: number,
    name: string,
  ) {
    e.preventDefault();
    setContextMenu({ x: e.clientX, y: e.clientY, typeId, name });
  }

  function toggleCategory(cat: string) {
    setCategoryFilters((prev) => {
      const next = new Set(prev);
      if (next.has(cat)) next.delete(cat);
      else next.add(cat);
      return next;
    });
  }

  function toggleTech(tech: string) {
    setTechFilters((prev) => {
      const next = new Set(prev);
      if (next.has(tech)) next.delete(tech);
      else next.add(tech);
      return next;
    });
  }

  function handleSort(field: SortField) {
    if (sortField === field) {
      setSortDesc((v) => !v);
    } else {
      setSortField(field);
      setSortDesc(true);
    }
  }

  // Filter + sort
  const searchLower = search.toLowerCase();
  const filtered = useMemo(() => {
    return (opportunities?.opportunities ?? []).filter((o) => {
      if (categoryFilters.size > 0 && !categoryFilters.has(o.category)) return false;
      if (techFilters.size > 0 && !techFilters.has(o.tech_level)) return false;
      if (searchLower && !o.product_name.toLowerCase().includes(searchLower)) return false;
      if (useBlacklist && o.is_blacklisted) return false;
      if (useWhitelist && !o.is_whitelisted) return false;
      return true;
    });
  }, [opportunities, categoryFilters, techFilters, searchLower, useBlacklist, useWhitelist]);

  const sorted = useMemo(() => {
    return [...filtered].sort((a, b) => {
      let aVal: number;
      let bVal: number;
      switch (sortField) {
        case "roi":
          aVal = a.roi;
          bVal = b.roi;
          break;
        case "demand_per_day":
          aVal = a.demand_per_day;
          bVal = b.demand_per_day;
          break;
        case "days_of_supply":
          aVal = a.days_of_supply;
          bVal = b.days_of_supply;
          break;
        case "material_cost":
          aVal = a.material_cost;
          bVal = b.material_cost;
          break;
        case "revenue":
          aVal = a.revenue;
          bVal = b.revenue;
          break;
        case "sales_tax":
          aVal = a.sales_tax;
          bVal = b.sales_tax;
          break;
        default:
          aVal = a.profit;
          bVal = b.profit;
      }
      return sortDesc ? bVal - aVal : aVal - bVal;
    });
  }, [filtered, sortField, sortDesc]);

  function SortHeader({
    field,
    label,
    className,
  }: {
    field: SortField;
    label: string;
    className?: string;
  }) {
    const isActive = sortField === field;
    return (
      <TableHead
        className={cn("py-2 text-xs cursor-pointer select-none hover:text-text-primary transition-colors", className)}
        onClick={() => handleSort(field)}
      >
        <span className="flex items-center gap-0.5">
          {label}
          {isActive ? (
            sortDesc ? (
              <ChevronDown className="h-3 w-3" />
            ) : (
              <ChevronUp className="h-3 w-3" />
            )
          ) : null}
        </span>
      </TableHead>
    );
  }

  const CATEGORIES = ["ship", "module", "charge", "drone", "implant", "booster", "structure", "deployable", "fighter"];
  const TECH_LEVELS = ["Tech I", "Tech II", "Tech III", "Faction", "Storyline", "Officer"];

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unauthorized />;
  }

  return (
    <>
      <Head>
        <title>Arbiter — pinky.tools</title>
      </Head>
      <Navbar />
      <div className="px-4 pb-4 bg-background-void min-h-screen flex flex-col">
        {/* Page title */}
        <div className="flex items-center gap-3 py-3">
          <h1 className="text-xl font-semibold text-text-heading">Arbiter</h1>
          <span className="text-sm text-text-muted">Industry Advisor</span>
        </div>

        {/* Settings accordion */}
        <Collapsible open={settingsOpen} onOpenChange={setSettingsOpen}>
          <div className="border border-overlay-subtle rounded-lg mb-4 bg-background-panel">
            <CollapsibleTrigger asChild>
              <button className="w-full flex items-center justify-between px-4 py-3 hover:bg-interactive-hover rounded-t-lg transition-colors">
                <span className="flex items-center gap-2 text-sm font-medium text-text-primary">
                  <Settings2 className="h-4 w-4 text-text-muted" />
                  Settings
                </span>
                {settingsOpen ? (
                  <ChevronUp className="h-4 w-4 text-text-muted" />
                ) : (
                  <ChevronDown className="h-4 w-4 text-text-muted" />
                )}
              </button>
            </CollapsibleTrigger>

            <CollapsibleContent>
              <div className="border-t border-overlay-subtle">
                <Tabs value={settingsTab} onValueChange={setSettingsTab}>
                  <TabsList className="border-b border-overlay-medium bg-transparent w-full justify-start rounded-none p-0 h-auto">
                    <TabsTrigger
                      value="structures"
                      className="text-text-muted data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none bg-transparent px-4 py-2 text-sm"
                    >
                      Structures
                    </TabsTrigger>
                    <TabsTrigger
                      value="tax"
                      className="text-text-muted data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none bg-transparent px-4 py-2 text-sm"
                    >
                      Tax &amp; Pricing
                    </TabsTrigger>
                    <TabsTrigger
                      value="lists"
                      className="text-text-muted data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none bg-transparent px-4 py-2 text-sm"
                    >
                      Lists
                    </TabsTrigger>
                  </TabsList>

                  {/* Structures tab */}
                  <TabsContent value="structures" className="p-4 m-0">
                    <div className="flex flex-wrap gap-6">
                      <StructureSection
                        title="Reaction"
                        prefix="reaction"
                        settings={settings}
                        onChange={handleSettingChange}
                      />
                      <StructureSection
                        title="Invention"
                        prefix="invention"
                        settings={settings}
                        onChange={handleSettingChange}
                      />
                      <StructureSection
                        title="Component Build"
                        prefix="component"
                        settings={settings}
                        onChange={handleSettingChange}
                      />
                      <StructureSection
                        title="Final Build"
                        prefix="final"
                        settings={settings}
                        onChange={handleSettingChange}
                      />
                    </div>
                  </TabsContent>

                  {/* Tax & Pricing tab */}
                  <TabsContent value="tax" className="p-4 m-0">
                    <div className="grid grid-cols-2 md:grid-cols-3 gap-6 max-w-2xl">
                      <div className="space-y-1">
                        <Label className="text-xs text-text-secondary">
                          Sales Tax %
                        </Label>
                        <Input
                          type="number"
                          className="h-8 text-sm"
                          value={taxProfile.sales_tax_rate}
                          onChange={(e) =>
                            setTaxProfile((p) => ({
                              ...p,
                              sales_tax_rate: parseFloat(e.target.value) || 0,
                            }))
                          }
                          min={0}
                          max={100}
                          step={0.1}
                        />
                      </div>
                      <div className="space-y-1">
                        <Label className="text-xs text-text-secondary">
                          Broker Fee %
                        </Label>
                        <Input
                          type="number"
                          className="h-8 text-sm"
                          value={taxProfile.broker_fee_rate}
                          onChange={(e) =>
                            setTaxProfile((p) => ({
                              ...p,
                              broker_fee_rate: parseFloat(e.target.value) || 0,
                            }))
                          }
                          min={0}
                          max={100}
                          step={0.1}
                        />
                      </div>
                      <div className="space-y-1">
                        <Label className="text-xs text-text-secondary">
                          Structure Broker Fee %
                        </Label>
                        <Input
                          type="number"
                          className="h-8 text-sm"
                          value={taxProfile.structure_broker_fee}
                          onChange={(e) =>
                            setTaxProfile((p) => ({
                              ...p,
                              structure_broker_fee:
                                parseFloat(e.target.value) || 0,
                            }))
                          }
                          min={0}
                          max={100}
                          step={0.1}
                        />
                      </div>
                      <div className="space-y-2 col-span-2">
                        <Label className="text-xs text-text-secondary">
                          Input Price Type
                        </Label>
                        <div className="flex gap-2">
                          <button
                            onClick={() =>
                              setTaxProfile((p) => ({
                                ...p,
                                input_price_type: "sell",
                              }))
                            }
                            className={cn(
                              "px-3 py-1.5 text-sm rounded border transition-colors",
                              taxProfile.input_price_type === "sell"
                                ? "bg-primary/10 border-primary/30 text-primary"
                                : "border-overlay-subtle text-text-muted hover:border-overlay-medium hover:text-text-secondary",
                            )}
                          >
                            Jita Sell
                          </button>
                          <button
                            onClick={() =>
                              setTaxProfile((p) => ({
                                ...p,
                                input_price_type: "buy",
                              }))
                            }
                            className={cn(
                              "px-3 py-1.5 text-sm rounded border transition-colors",
                              taxProfile.input_price_type === "buy"
                                ? "bg-primary/10 border-primary/30 text-primary"
                                : "border-overlay-subtle text-text-muted hover:border-overlay-medium hover:text-text-secondary",
                            )}
                          >
                            Jita Buy
                          </button>
                        </div>
                      </div>
                      <div className="space-y-2 col-span-2">
                        <Label className="text-xs text-text-secondary">
                          Output Price Type
                        </Label>
                        <div className="flex gap-2">
                          <button
                            onClick={() =>
                              setTaxProfile((p) => ({
                                ...p,
                                output_price_type: "sell",
                              }))
                            }
                            className={cn(
                              "px-3 py-1.5 text-sm rounded border transition-colors",
                              taxProfile.output_price_type === "sell"
                                ? "bg-primary/10 border-primary/30 text-primary"
                                : "border-overlay-subtle text-text-muted hover:border-overlay-medium hover:text-text-secondary",
                            )}
                          >
                            Jita Sell
                          </button>
                          <button
                            onClick={() =>
                              setTaxProfile((p) => ({
                                ...p,
                                output_price_type: "buy",
                              }))
                            }
                            className={cn(
                              "px-3 py-1.5 text-sm rounded border transition-colors",
                              taxProfile.output_price_type === "buy"
                                ? "bg-primary/10 border-primary/30 text-primary"
                                : "border-overlay-subtle text-text-muted hover:border-overlay-medium hover:text-text-secondary",
                            )}
                          >
                            Jita Buy
                          </button>
                        </div>
                      </div>
                    </div>
                  </TabsContent>

                  {/* Lists tab */}
                  <TabsContent value="lists" className="p-4 m-0">
                    <div className="flex items-center gap-8 mb-4">
                      <Label className="flex items-center gap-2 text-sm cursor-pointer">
                        <Switch
                          checked={useWhitelist}
                          onCheckedChange={setUseWhitelist}
                        />
                        Use Whitelist (show only whitelisted)
                      </Label>
                      <Label className="flex items-center gap-2 text-sm cursor-pointer">
                        <Switch
                          checked={useBlacklist}
                          onCheckedChange={setUseBlacklist}
                        />
                        Use Blacklist (hide blacklisted)
                      </Label>
                    </div>
                    <div className="flex gap-6">
                      {/* Whitelist */}
                      <div className="flex-1">
                        <div className="flex items-center justify-between mb-2">
                          <span className="text-sm font-medium text-teal-success">
                            Whitelist ({whitelist.length})
                          </span>
                        </div>
                        <div className="border border-overlay-subtle rounded max-h-48 overflow-y-auto">
                          {whitelist.length === 0 ? (
                            <div className="flex items-center justify-center py-6">
                              <p className="text-xs text-text-muted">
                                No items — right-click in results to add
                              </p>
                            </div>
                          ) : (
                            <table className="w-full text-xs">
                              <tbody>
                                {whitelist.map((item) => (
                                  <tr
                                    key={item.type_id}
                                    className="border-b border-overlay-subtle/50 hover:bg-interactive-hover"
                                  >
                                    <td className="px-3 py-1.5 text-text-primary">
                                      {item.name}
                                    </td>
                                    <td className="px-2 py-1.5">
                                      <button
                                        onClick={() =>
                                          handleRemoveFromList(
                                            "whitelist",
                                            item.type_id,
                                          )
                                        }
                                        className="text-text-muted hover:text-rose-danger"
                                      >
                                        <X className="h-3 w-3" />
                                      </button>
                                    </td>
                                  </tr>
                                ))}
                              </tbody>
                            </table>
                          )}
                        </div>
                      </div>
                      {/* Blacklist */}
                      <div className="flex-1">
                        <div className="flex items-center justify-between mb-2">
                          <span className="text-sm font-medium text-rose-danger">
                            Blacklist ({blacklist.length})
                          </span>
                        </div>
                        <div className="border border-overlay-subtle rounded max-h-48 overflow-y-auto">
                          {blacklist.length === 0 ? (
                            <div className="flex items-center justify-center py-6">
                              <p className="text-xs text-text-muted">
                                No items — right-click in results to add
                              </p>
                            </div>
                          ) : (
                            <table className="w-full text-xs">
                              <tbody>
                                {blacklist.map((item) => (
                                  <tr
                                    key={item.type_id}
                                    className="border-b border-overlay-subtle/50 hover:bg-interactive-hover"
                                  >
                                    <td className="px-3 py-1.5 text-text-primary">
                                      {item.name}
                                    </td>
                                    <td className="px-2 py-1.5">
                                      <button
                                        onClick={() =>
                                          handleRemoveFromList(
                                            "blacklist",
                                            item.type_id,
                                          )
                                        }
                                        className="text-text-muted hover:text-rose-danger"
                                      >
                                        <X className="h-3 w-3" />
                                      </button>
                                    </td>
                                  </tr>
                                ))}
                              </tbody>
                            </table>
                          )}
                        </div>
                      </div>
                    </div>
                    <div className="mt-3">
                      <Link
                        href="/arbiter/lists"
                        className="text-xs text-primary hover:underline flex items-center gap-1"
                      >
                        Manage lists in full view
                        <ExternalLink className="h-3 w-3" />
                      </Link>
                    </div>
                  </TabsContent>
                </Tabs>

                {saveError && (
                  <p className="text-sm text-rose-danger px-4 pb-2">{saveError}</p>
                )}

                <div className="px-4 pb-4 flex justify-end">
                  <Button
                    size="sm"
                    onClick={handleSaveSettings}
                    disabled={savingSettings}
                  >
                    {savingSettings ? (
                      <Loader2 className="h-4 w-4 animate-spin mr-2" />
                    ) : null}
                    Save Settings
                  </Button>
                </div>
              </div>
            </CollapsibleContent>
          </div>
        </Collapsible>

        {/* Main controls bar */}
        <div className="flex items-center gap-3 mb-3 flex-wrap">
          {/* Left side */}
          <div className="flex items-center gap-2 flex-1 flex-wrap">
            {/* Scope dropdown + add */}
            <div className="flex items-center gap-1">
              <Select
                value={selectedScopeId ? String(selectedScopeId) : ""}
                onValueChange={(v) => setSelectedScopeId(parseInt(v))}
              >
                <SelectTrigger className="h-8 text-sm w-44">
                  <SelectValue placeholder="Select scope..." />
                </SelectTrigger>
                <SelectContent>
                  {scopes.map((s) => (
                    <SelectItem key={s.id} value={String(s.id)}>
                      {s.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <div className="flex items-center gap-1">
                      <Input
                        className="h-8 text-sm w-28"
                        placeholder="New scope..."
                        value={newScopeName}
                        onChange={(e) => setNewScopeName(e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === "Enter") handleCreateScope();
                        }}
                      />
                      <Button
                        size="sm"
                        variant="outline"
                        className="h-8 w-8 p-0"
                        onClick={handleCreateScope}
                        disabled={addingScope || !newScopeName.trim()}
                      >
                        {addingScope ? (
                          <Loader2 className="h-3.5 w-3.5 animate-spin" />
                        ) : (
                          <Plus className="h-3.5 w-3.5" />
                        )}
                      </Button>
                    </div>
                  </TooltipTrigger>
                  <TooltipContent>Create new scope</TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>

            {/* Decryptor dropdown */}
            <Select
              value={selectedDecryptorId}
              onValueChange={setSelectedDecryptorId}
            >
              <SelectTrigger className="h-8 text-sm w-48">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="maximize_roi">Maximize ROI</SelectItem>
                {decryptors.map((d) => (
                  <SelectItem key={d.type_id} value={String(d.type_id)}>
                    {d.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            {/* Scan button */}
            <Button
              size="sm"
              onClick={handleScan}
              disabled={scanning}
            >
              {scanning ? (
                <Loader2 className="h-4 w-4 animate-spin mr-2" />
              ) : null}
              {scanning ? "Scanning..." : "Scan Opportunities"}
            </Button>
          </div>

          {/* Right side: price toggles */}
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-1.5">
              <span className="text-xs text-text-muted">Input:</span>
              <button
                onClick={() => setInputPrice("sell")}
                className={cn(
                  "px-2.5 py-1 text-xs rounded border transition-colors",
                  inputPrice === "sell"
                    ? "bg-primary/10 border-primary/30 text-primary"
                    : "border-overlay-subtle text-text-muted hover:border-overlay-medium",
                )}
              >
                Sell
              </button>
              <button
                onClick={() => setInputPrice("buy")}
                className={cn(
                  "px-2.5 py-1 text-xs rounded border transition-colors",
                  inputPrice === "buy"
                    ? "bg-primary/10 border-primary/30 text-primary"
                    : "border-overlay-subtle text-text-muted hover:border-overlay-medium",
                )}
              >
                Buy
              </button>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="text-xs text-text-muted">Output:</span>
              <button
                onClick={() => setOutputPrice("sell")}
                className={cn(
                  "px-2.5 py-1 text-xs rounded border transition-colors",
                  outputPrice === "sell"
                    ? "bg-primary/10 border-primary/30 text-primary"
                    : "border-overlay-subtle text-text-muted hover:border-overlay-medium",
                )}
              >
                Sell
              </button>
              <button
                onClick={() => setOutputPrice("buy")}
                className={cn(
                  "px-2.5 py-1 text-xs rounded border transition-colors",
                  outputPrice === "buy"
                    ? "bg-primary/10 border-primary/30 text-primary"
                    : "border-overlay-subtle text-text-muted hover:border-overlay-medium",
                )}
              >
                Buy
              </button>
            </div>
          </div>
        </div>

        {/* Filter bar */}
        <div className="flex items-center gap-4 mb-3 flex-wrap">
          {/* Category filters */}
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-xs text-text-muted">Category:</span>
            {CATEGORIES.map((cat) => (
              <button
                key={cat}
                onClick={() => toggleCategory(cat)}
                className={cn(
                  "px-2.5 py-0.5 text-xs rounded-full border capitalize transition-colors",
                  categoryFilters.has(cat)
                    ? "bg-primary/10 border-primary/30 text-primary"
                    : "border-overlay-subtle text-text-muted hover:border-overlay-medium hover:text-text-secondary",
                )}
              >
                {cat}
              </button>
            ))}
          </div>

          {/* Tech level filters */}
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-xs text-text-muted">Tech:</span>
            {TECH_LEVELS.map((tech) => (
              <button
                key={tech}
                onClick={() => toggleTech(tech)}
                className={cn(
                  "px-2.5 py-0.5 text-xs rounded-full border transition-colors",
                  techFilters.has(tech)
                    ? "bg-primary/10 border-primary/30 text-primary"
                    : "border-overlay-subtle text-text-muted hover:border-overlay-medium hover:text-text-secondary",
                )}
              >
                {tech}
              </button>
            ))}
          </div>

          {/* Search */}
          <div className="relative ml-auto">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-text-muted pointer-events-none" />
            <Input
              className="pl-8 pr-7 h-7 text-xs w-48 bg-background-elevated border-overlay-subtle"
              placeholder="Search..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
            {search && (
              <button
                onClick={() => setSearch("")}
                className="absolute right-2 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-primary"
              >
                <X className="h-3.5 w-3.5" />
              </button>
            )}
          </div>
        </div>

        {/* Results table */}
        <div className="flex-1 border border-overlay-subtle rounded-lg bg-background-panel overflow-hidden">
          {scanning ? (
            <div className="flex flex-col items-center justify-center py-16 gap-3">
              <Loader2 className="h-8 w-8 animate-spin text-primary" />
              <p className="text-sm text-text-secondary">
                Scanning blueprints... this may take several minutes
              </p>
            </div>
          ) : scanError ? (
            <div className="flex items-center justify-center py-12">
              <p className="text-sm text-rose-danger">{scanError}</p>
            </div>
          ) : opportunities === null ? (
            <div className="flex flex-col items-center justify-center py-16 gap-2">
              <p className="text-sm text-text-secondary">
                Configure your settings above, then click Scan Opportunities
              </p>
            </div>
          ) : sorted.length === 0 ? (
            <div className="flex items-center justify-center py-12">
              <p className="text-sm text-text-secondary">
                {search
                  ? `No results matching "${search}"`
                  : "No opportunities with current filters"}
              </p>
            </div>
          ) : (
            <>
              {/* Scan summary */}
              <div className="flex items-center gap-4 px-4 py-2 border-b border-overlay-subtle text-xs text-text-muted">
                <span>
                  Scanned {opportunities.total_scanned} blueprints
                </span>
                {opportunities.best_character_name && (
                  <span>
                    Inventor:{" "}
                    <span className="text-text-secondary">
                      {opportunities.best_character_name}
                    </span>
                  </span>
                )}
                <span>
                  {sorted.length} result{sorted.length !== 1 ? "s" : ""}
                  {search ? ` matching "${search}"` : ""}
                </span>
              </div>
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow className="border-overlay-subtle hover:bg-transparent">
                      <TableHead className="py-2 w-8 text-xs text-text-muted">#</TableHead>
                      <TableHead className="py-2 text-xs text-text-muted">Name</TableHead>
                      <TableHead className="py-2 text-xs text-text-muted">Duration</TableHead>
                      <SortHeader field="demand_per_day" label="Demand/d" />
                      <SortHeader field="days_of_supply" label="D.O.S." />
                      <TableHead className="py-2 text-xs text-text-muted">Qty</TableHead>
                      <SortHeader field="roi" label="ROI" />
                      <SortHeader field="material_cost" label="Cost" />
                      <SortHeader field="revenue" label="Revenue" />
                      <SortHeader field="sales_tax" label="Sales Tax" />
                      <SortHeader field="profit" label="Profit" />
                      <TableHead className="py-2 text-xs text-text-muted">ME</TableHead>
                      <TableHead className="py-2 text-xs text-text-muted">TE</TableHead>
                      <TableHead className="text-xs font-medium text-text-secondary py-2">Max Runs</TableHead>
                      <TableHead className="text-xs font-medium text-text-secondary py-2">Decryptor</TableHead>
                      <TableHead className="py-2 text-xs text-text-muted">Group</TableHead>
                      <TableHead className="py-2 text-xs text-text-muted">Category</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {sorted.map((opp, i) => (
                      <OpportunityRow
                        key={opp.product_type_id}
                        opp={opp}
                        rank={i + 1}
                        qty={rowQtys[opp.product_type_id] ?? 1}
                        onQtyChange={(typeId, qty) =>
                          setRowQtys((prev) => ({ ...prev, [typeId]: qty }))
                        }
                        scopeId={selectedScopeId}
                        onContextMenu={handleContextMenu}
                        onSelect={setSelectedOpp}
                        isSelected={
                          selectedOpp?.product_type_id === opp.product_type_id
                        }
                        inputPrice={inputPrice}
                        outputPrice={outputPrice}
                      />
                    ))}
                  </TableBody>
                </Table>
              </div>
            </>
          )}
        </div>

        {/* Warehouse panel */}
        <WarehousePanel
          selectedOpp={selectedOpp}
          scopeId={selectedScopeId}
          qty={
            selectedOpp
              ? (rowQtys[selectedOpp.product_type_id] ?? 1)
              : 1
          }
          inputPrice={inputPrice}
        />
      </div>

      {/* Context menu */}
      {contextMenu && (
        <ContextMenu
          menu={contextMenu}
          blacklist={blacklist}
          whitelist={whitelist}
          onClose={() => setContextMenu(null)}
          onAddBlacklist={(id, name) => handleAddToList("blacklist", id, name)}
          onRemoveBlacklist={(id) => handleRemoveFromList("blacklist", id)}
          onAddWhitelist={(id, name) => handleAddToList("whitelist", id, name)}
          onRemoveWhitelist={(id) => handleRemoveFromList("whitelist", id)}
        />
      )}
    </>
  );
}
