import Head from "next/head";
import { useSession } from "next-auth/react";
import { useState, useCallback } from "react";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import Navbar from "@industry-tool/components/Navbar";
import { formatISK } from "@industry-tool/utils/formatting";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
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
import { Loader2, ChevronDown, ChevronUp, ChevronRight, Settings2, Search, X, ChevronsUpDown } from "lucide-react";
import { cn } from "@/lib/utils";

// --- Types ---

export interface ArbiterSettings {
  reaction_structure: string;
  reaction_rig: string;
  reaction_security: string;
  reaction_system_id: number;
  reaction_system_name: string;
  invention_structure: string;
  invention_rig: string;
  invention_security: string;
  invention_system_id: number;
  invention_system_name: string;
  component_structure: string;
  component_rig: string;
  component_security: string;
  component_system_id: number;
  component_system_name: string;
  final_structure: string;
  final_rig: string;
  final_security: string;
  final_system_id: number;
  final_system_name: string;
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
  jita_sell_price: number;
  jita_buy_price: number;
  best_decryptor: DecryptorResult;
  all_decryptors: DecryptorResult[];
}

export interface OpportunitiesResponse {
  opportunities: Opportunity[];
  generated_at: string;
  total_scanned: number;
  best_character_name: string;
}

// --- Constants ---

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

const SECURITY_OPTIONS = [
  { value: "high", label: "High Sec" },
  { value: "low", label: "Low Sec" },
  { value: "null", label: "Null Sec" },
];

const DEFAULT_SETTINGS: ArbiterSettings = {
  reaction_structure: "tatara",
  reaction_rig: "t2",
  reaction_security: "null",
  reaction_system_id: 30000142,
  reaction_system_name: "Jita",
  invention_structure: "raitaru",
  invention_rig: "t1",
  invention_security: "high",
  invention_system_id: 30000142,
  invention_system_name: "Jita",
  component_structure: "raitaru",
  component_rig: "t2",
  component_security: "null",
  component_system_id: 30000142,
  component_system_name: "Jita",
  final_structure: "azbel",
  final_rig: "t2",
  final_security: "null",
  final_system_id: 30000142,
  final_system_name: "Jita",
};

type SortField = "profit" | "roi" | "isk_per_day";
type FilterCategory = "all" | "ship" | "module";
type CategorySort = "none" | "ships_first" | "modules_first";

// --- StructureSection ---

interface StructureSectionProps {
  title: string;
  prefix: "reaction" | "invention" | "component" | "final";
  settings: ArbiterSettings;
  onChange: (key: keyof ArbiterSettings, value: string | number) => void;
}

function StructureSection({ title, prefix, settings, onChange }: StructureSectionProps) {
  const structureKey = `${prefix}_structure` as keyof ArbiterSettings;
  const rigKey = `${prefix}_rig` as keyof ArbiterSettings;
  const securityKey = `${prefix}_security` as keyof ArbiterSettings;
  const systemNameKey = `${prefix}_system_name` as keyof ArbiterSettings;

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
        <Label className="text-xs text-text-secondary">Security</Label>
        <Select
          value={settings[securityKey] as string}
          onValueChange={(v) => onChange(securityKey, v)}
        >
          <SelectTrigger className="h-8 text-sm">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {SECURITY_OPTIONS.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-1">
        <Label className="text-xs text-text-secondary">System</Label>
        <Input
          className="h-8 text-sm"
          value={settings[systemNameKey] as string}
          onChange={(e) => onChange(systemNameKey, e.target.value)}
          placeholder="System name"
        />
      </div>
    </div>
  );
}

// --- DecryptorRow ---

interface DecryptorRowProps {
  decryptor: DecryptorResult;
  isBest: boolean;
}

function DecryptorRow({ decryptor, isBest }: DecryptorRowProps) {
  const chancePercent = Math.round(decryptor.probability_multiplier * 100);

  return (
    <TableRow
      className={cn(
        isBest
          ? "bg-teal-success/10 border-teal-success/20"
          : "hover:bg-interactive-hover"
      )}
    >
      <TableCell className="py-1.5 pl-8 text-sm">
        <span className="flex items-center gap-2">
          {decryptor.name}
          {isBest && (
            <Badge className="text-[10px] px-1.5 py-0 bg-teal-success/20 text-teal-success border-teal-success/30">
              Best
            </Badge>
          )}
        </span>
      </TableCell>
      <TableCell className="py-1.5 text-sm text-text-secondary">
        {chancePercent}%
      </TableCell>
      <TableCell className="py-1.5 text-sm text-text-secondary">
        ME{decryptor.resulting_me}
      </TableCell>
      <TableCell className="py-1.5 text-sm text-text-secondary">
        {decryptor.resulting_runs} run{decryptor.resulting_runs !== 1 ? "s" : ""}
      </TableCell>
      <TableCell className="py-1.5 text-sm text-text-data-value font-mono">
        {formatISK(decryptor.invention_cost)}
      </TableCell>
      <TableCell className="py-1.5 text-sm text-text-data-value font-mono">
        {formatISK(decryptor.material_cost)}
      </TableCell>
      <TableCell className="py-1.5 text-sm text-text-data-value font-mono">
        {formatISK(decryptor.total_cost)}
      </TableCell>
      <TableCell
        className={cn(
          "py-1.5 text-sm font-mono",
          decryptor.profit >= 0 ? "text-teal-success" : "text-rose-danger"
        )}
      >
        {formatISK(decryptor.profit)}
      </TableCell>
      <TableCell
        className={cn(
          "py-1.5 text-sm font-mono",
          decryptor.roi >= 0 ? "text-teal-success" : "text-rose-danger"
        )}
      >
        {decryptor.roi.toFixed(1)}%
      </TableCell>
    </TableRow>
  );
}

// --- OpportunityRow ---

interface OpportunityRowProps {
  opp: Opportunity;
  rank: number;
}

function OpportunityRow({ opp, rank }: OpportunityRowProps) {
  const [expanded, setExpanded] = useState(false);

  return (
    <>
      <TableRow
        className="cursor-pointer hover:bg-interactive-hover"
        onClick={() => setExpanded((v) => !v)}
      >
        <TableCell className="py-2 text-sm text-text-muted w-8">{rank}</TableCell>
        <TableCell className="py-2">
          <span className="flex items-center gap-2 text-sm font-medium text-text-primary">
            {expanded ? (
              <ChevronDown className="h-3.5 w-3.5 text-text-muted flex-shrink-0" />
            ) : (
              <ChevronRight className="h-3.5 w-3.5 text-text-muted flex-shrink-0" />
            )}
            {opp.product_name}
          </span>
        </TableCell>
        <TableCell className="py-2">
          <Badge
            className={cn(
              "text-[10px] px-1.5 py-0",
              opp.category === "ship"
                ? "bg-blue-science/20 text-blue-science border-blue-science/30"
                : "bg-category-violet/20 text-category-violet border-category-violet/30"
            )}
          >
            {opp.category}
          </Badge>
        </TableCell>
        <TableCell className="py-2 text-sm text-text-data-value font-mono">
          {formatISK(opp.jita_sell_price)}
        </TableCell>
        <TableCell className="py-2 text-sm text-text-data-value font-mono">
          {formatISK(opp.best_decryptor.total_cost)}
        </TableCell>
        <TableCell
          className={cn(
            "py-2 text-sm font-mono",
            opp.best_decryptor.profit >= 0 ? "text-teal-success" : "text-rose-danger"
          )}
        >
          {formatISK(opp.best_decryptor.profit)}
        </TableCell>
        <TableCell
          className={cn(
            "py-2 text-sm font-mono",
            opp.best_decryptor.roi >= 0 ? "text-teal-success" : "text-rose-danger"
          )}
        >
          {opp.best_decryptor.roi.toFixed(1)}%
        </TableCell>
        <TableCell className="py-2 text-sm text-text-data-value font-mono">
          {formatISK(opp.best_decryptor.isk_per_day)}/day
        </TableCell>
        <TableCell className="py-2 text-sm text-text-secondary">
          {opp.best_decryptor.name}
        </TableCell>
      </TableRow>

      {expanded && (
        <TableRow className="bg-background-panel hover:bg-background-panel">
          <TableCell colSpan={9} className="p-0">
            <div className="border-t border-overlay-subtle">
              <Table>
                <TableHeader>
                  <TableRow className="hover:bg-transparent border-overlay-subtle">
                    <TableHead className="py-1.5 pl-8 text-xs text-text-muted">Decryptor</TableHead>
                    <TableHead className="py-1.5 text-xs text-text-muted">Chance</TableHead>
                    <TableHead className="py-1.5 text-xs text-text-muted">Result ME</TableHead>
                    <TableHead className="py-1.5 text-xs text-text-muted">Result Runs</TableHead>
                    <TableHead className="py-1.5 text-xs text-text-muted">Invention Cost</TableHead>
                    <TableHead className="py-1.5 text-xs text-text-muted">Material Cost</TableHead>
                    <TableHead className="py-1.5 text-xs text-text-muted">Total Cost</TableHead>
                    <TableHead className="py-1.5 text-xs text-text-muted">Profit</TableHead>
                    <TableHead className="py-1.5 text-xs text-text-muted">ROI</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {opp.all_decryptors.map((d) => (
                    <DecryptorRow
                      key={d.type_id}
                      decryptor={d}
                      isBest={d.type_id === opp.best_decryptor.type_id}
                    />
                  ))}
                </TableBody>
              </Table>
            </div>
          </TableCell>
        </TableRow>
      )}
    </>
  );
}

// --- ArbiterPage ---

export default function ArbiterPage() {
  const { status } = useSession();
  const [settings, setSettings] = useState<ArbiterSettings>(DEFAULT_SETTINGS);
  const [settingsOpen, setSettingsOpen] = useState(true);
  const [savingSettings, setSavingSettings] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  const [opportunities, setOpportunities] = useState<OpportunitiesResponse | null>(null);
  const [scanning, setScanning] = useState(false);
  const [scanError, setScanError] = useState<string | null>(null);

  const [filter, setFilter] = useState<FilterCategory>("all");
  const [sortBy, setSortBy] = useState<SortField>("profit");
  const [categorySort, setCategorySort] = useState<CategorySort>("none");
  const [search, setSearch] = useState("");

  const handleSettingChange = useCallback(
    (key: keyof ArbiterSettings, value: string | number) => {
      setSettings((prev) => ({ ...prev, [key]: value }));
    },
    []
  );

  const handleSaveSettings = async () => {
    setSavingSettings(true);
    setSaveError(null);
    try {
      const body = {
        reaction_structure: settings.reaction_structure,
        reaction_rig: settings.reaction_rig,
        reaction_security: settings.reaction_security,
        reaction_system_id: settings.reaction_system_id,
        invention_structure: settings.invention_structure,
        invention_rig: settings.invention_rig,
        invention_security: settings.invention_security,
        invention_system_id: settings.invention_system_id,
        component_structure: settings.component_structure,
        component_rig: settings.component_rig,
        component_security: settings.component_security,
        component_system_id: settings.component_system_id,
        final_structure: settings.final_structure,
        final_rig: settings.final_rig,
        final_security: settings.final_security,
        final_system_id: settings.final_system_id,
      };
      const res = await fetch("/api/arbiter/settings", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (!res.ok) {
        const err = await res.text();
        setSaveError(err || "Failed to save settings");
      }
    } catch (err) {
      setSaveError("Failed to save settings");
      console.error("Save settings error:", err);
    } finally {
      setSavingSettings(false);
    }
  };

  const handleScan = useCallback(async () => {
    setScanning(true);
    setScanError(null);
    try {
      const res = await fetch("/api/arbiter/opportunities");
      if (!res.ok) {
        const err = await res.text();
        setScanError(err || "Scan failed");
        return;
      }
      const data: OpportunitiesResponse = await res.json();
      setOpportunities(data);
    } catch (err) {
      setScanError("Failed to scan opportunities");
      console.error("Scan error:", err);
    } finally {
      setScanning(false);
    }
  }, []);

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unauthorized />;
  }

  // Filter + sort
  const searchLower = search.toLowerCase();
  const filtered = (opportunities?.opportunities ?? []).filter((o) => {
    if (filter !== "all" && o.category !== filter) return false;
    if (searchLower && !o.product_name.toLowerCase().includes(searchLower)) return false;
    return true;
  });

  const sorted = [...filtered].sort((a, b) => {
    // Category sort takes precedence when active
    if (categorySort !== "none") {
      const aIsShip = a.category === "ship" ? 0 : 1;
      const bIsShip = b.category === "ship" ? 0 : 1;
      if (categorySort === "ships_first" && aIsShip !== bIsShip) return aIsShip - bIsShip;
      if (categorySort === "modules_first" && aIsShip !== bIsShip) return bIsShip - aIsShip;
    }
    switch (sortBy) {
      case "roi":
        return b.best_decryptor.roi - a.best_decryptor.roi;
      case "isk_per_day":
        return b.best_decryptor.isk_per_day - a.best_decryptor.isk_per_day;
      default:
        return b.best_decryptor.profit - a.best_decryptor.profit;
    }
  });

  const handleCategorySort = () => {
    setCategorySort((prev) => {
      if (prev === "none") return "ships_first";
      if (prev === "ships_first") return "modules_first";
      return "none";
    });
  };

  const categorySortIcon = categorySort === "ships_first"
    ? <ChevronUp className="h-3 w-3 inline-block ml-1" />
    : categorySort === "modules_first"
    ? <ChevronDown className="h-3 w-3 inline-block ml-1" />
    : <ChevronsUpDown className="h-3 w-3 inline-block ml-1 opacity-40" />;

  const SortButton = ({
    field,
    label,
  }: {
    field: SortField;
    label: string;
  }) => (
    <button
      onClick={() => setSortBy(field)}
      className={cn(
        "px-3 py-1 text-xs rounded border transition-colors",
        sortBy === field
          ? "bg-primary/10 border-primary/30 text-primary"
          : "border-overlay-subtle text-text-muted hover:border-overlay-medium hover:text-text-secondary"
      )}
    >
      {label}
    </button>
  );

  const FilterButton = ({
    value,
    label,
  }: {
    value: FilterCategory;
    label: string;
  }) => (
    <button
      onClick={() => setFilter(value)}
      className={cn(
        "px-3 py-1 text-sm rounded transition-colors",
        filter === value
          ? "bg-primary/10 text-primary"
          : "text-text-muted hover:text-text-secondary"
      )}
    >
      {label}
    </button>
  );

  return (
    <>
      <Head>
        <title>Arbiter — pinky.tools</title>
      </Head>
      <Navbar />
      <div className="px-4 pb-8 bg-background-void min-h-screen">
        {/* Page Title */}
        <div className="flex items-center gap-3 mb-4">
          <h1 className="text-xl font-semibold text-text-heading">
            Arbiter
          </h1>
          <span className="text-sm text-text-muted">Industry Advisor</span>
        </div>

        {/* Settings Panel */}
        <Collapsible open={settingsOpen} onOpenChange={setSettingsOpen}>
          <div className="border border-overlay-subtle rounded-lg mb-6 bg-background-panel">
            <CollapsibleTrigger asChild>
              <button className="w-full flex items-center justify-between px-4 py-3 hover:bg-interactive-hover rounded-t-lg transition-colors">
                <span className="flex items-center gap-2 text-sm font-medium text-text-primary">
                  <Settings2 className="h-4 w-4 text-text-muted" />
                  Production Structure Settings
                </span>
                {settingsOpen ? (
                  <ChevronUp className="h-4 w-4 text-text-muted" />
                ) : (
                  <ChevronDown className="h-4 w-4 text-text-muted" />
                )}
              </button>
            </CollapsibleTrigger>

            <CollapsibleContent>
              <div className="px-4 pb-4 border-t border-overlay-subtle">
                <div className="flex flex-wrap gap-6 mt-4">
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

                {saveError && (
                  <p className="text-sm text-rose-danger mt-3">{saveError}</p>
                )}

                <div className="mt-4 flex justify-end">
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

        {/* Opportunities Section */}
        <div className="border border-overlay-subtle rounded-lg bg-background-panel">
          {/* Header controls */}
          <div className="flex items-center justify-between px-4 py-3 border-b border-overlay-subtle">
            <div className="flex items-center gap-1">
              <FilterButton value="all" label="All" />
              <FilterButton value="ship" label="Ships" />
              <FilterButton value="module" label="Modules" />
            </div>

            <div className="flex items-center gap-3">
              <div className="flex items-center gap-1.5">
                <span className="text-xs text-text-muted">Sort:</span>
                <SortButton field="profit" label="Net Profit" />
                <SortButton field="roi" label="ROI" />
                <SortButton field="isk_per_day" label="ISK/Day" />
              </div>

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
          </div>

          {/* Search input */}
          <div className="px-4 py-2 border-b border-overlay-subtle">
            <div className="relative max-w-xs">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-text-muted pointer-events-none" />
              <Input
                className="pl-8 pr-8 h-8 text-sm bg-background-elevated border-overlay-subtle"
                placeholder="Search items..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
              {search && (
                <button
                  onClick={() => setSearch("")}
                  className="absolute right-2 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-primary transition-colors"
                  aria-label="Clear search"
                >
                  <X className="h-3.5 w-3.5" />
                </button>
              )}
            </div>
          </div>

          {/* Results */}
          {scanning ? (
            <div className="flex flex-col items-center justify-center py-16 gap-3">
              <Loader2 className="h-8 w-8 animate-spin text-primary" />
              <p className="text-sm text-text-secondary">
                Scanning T2 blueprints... this may take 10–30 seconds
              </p>
            </div>
          ) : scanError ? (
            <div className="flex items-center justify-center py-12">
              <p className="text-sm text-rose-danger">{scanError}</p>
            </div>
          ) : opportunities === null ? (
            <div className="flex flex-col items-center justify-center py-16 gap-2">
              <p className="text-sm text-text-secondary">
                Configure your structures above, then click Scan Opportunities
              </p>
            </div>
          ) : sorted.length === 0 ? (
            <div className="flex items-center justify-center py-12">
              <p className="text-sm text-text-secondary">
                {search
                  ? `No results matching "${search}"`
                  : "No profitable opportunities found with current market data"}
              </p>
            </div>
          ) : (
            <>
              {opportunities && (
                <div className="flex items-center gap-4 px-4 py-2 border-b border-overlay-subtle text-xs text-text-muted">
                  <span>
                    Scanned {opportunities.total_scanned} blueprints
                  </span>
                  {opportunities.best_character_name && (
                    <span>
                      Inventor: <span className="text-text-secondary">{opportunities.best_character_name}</span>
                    </span>
                  )}
                  <span>
                    {sorted.length} result{sorted.length !== 1 ? "s" : ""}
                    {filter !== "all" ? ` (${filter}s)` : ""}
                    {search ? ` matching "${search}"` : ""}
                  </span>
                </div>
              )}
              <Table>
                <TableHeader>
                  <TableRow className="border-overlay-subtle hover:bg-transparent">
                    <TableHead className="py-2 w-8 text-xs text-text-muted">#</TableHead>
                    <TableHead className="py-2 text-xs text-text-muted">Item</TableHead>
                    <TableHead className="py-2 text-xs text-text-muted">
                      <button
                        onClick={handleCategorySort}
                        className="flex items-center gap-0.5 hover:text-text-primary transition-colors"
                        title={
                          categorySort === "none"
                            ? "Sort: ships first"
                            : categorySort === "ships_first"
                            ? "Sort: modules first"
                            : "Remove category sort"
                        }
                      >
                        Category
                        {categorySortIcon}
                      </button>
                    </TableHead>
                    <TableHead className="py-2 text-xs text-text-muted">Sell Price</TableHead>
                    <TableHead className="py-2 text-xs text-text-muted">Total Cost</TableHead>
                    <TableHead className="py-2 text-xs text-text-muted">Net Profit</TableHead>
                    <TableHead className="py-2 text-xs text-text-muted">ROI</TableHead>
                    <TableHead className="py-2 text-xs text-text-muted">ISK/Day</TableHead>
                    <TableHead className="py-2 text-xs text-text-muted">Best Decryptor</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {sorted.map((opp, i) => (
                    <OpportunityRow
                      key={opp.product_type_id}
                      opp={opp}
                      rank={i + 1}
                    />
                  ))}
                </TableBody>
              </Table>
            </>
          )}
        </div>
      </div>
    </>
  );
}
