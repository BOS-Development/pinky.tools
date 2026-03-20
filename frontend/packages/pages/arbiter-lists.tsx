import Head from "next/head";
import { useSession } from "next-auth/react";
import { useState, useEffect, useCallback } from "react";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import Navbar from "@industry-tool/components/Navbar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Loader2, Search, X, ArrowLeft } from "lucide-react";
import Link from "next/link";

interface ListItem {
  type_id: number;
  name: string;
}

interface ItemSearchResult {
  TypeID: number;
  Name: string;
}

export default function ArbiterListsPage() {
  const { status } = useSession();

  const [whitelist, setWhitelist] = useState<ListItem[]>([]);
  const [blacklist, setBlacklist] = useState<ListItem[]>([]);
  const [loading, setLoading] = useState(true);

  const [wSearch, setWSearch] = useState("");
  const [bSearch, setBSearch] = useState("");

  const [wQuery, setWQuery] = useState("");
  const [bQuery, setBQuery] = useState("");
  const [wResults, setWResults] = useState<ItemSearchResult[]>([]);
  const [bResults, setBResults] = useState<ItemSearchResult[]>([]);
  const [wSearching, setWSearching] = useState(false);
  const [bSearching, setBSearching] = useState(false);

  useEffect(() => {
    if (status !== "authenticated") return;
    Promise.all([
      fetch("/api/arbiter/whitelist").then((r) => r.ok ? r.json() : []),
      fetch("/api/arbiter/blacklist").then((r) => r.ok ? r.json() : []),
    ])
      .then(([w, b]) => {
        if (Array.isArray(w)) setWhitelist(w);
        if (Array.isArray(b)) setBlacklist(b);
      })
      .finally(() => setLoading(false));
  }, [status]);

  // Search whitelist items
  useEffect(() => {
    if (!wQuery.trim()) {
      setWResults([]);
      return;
    }
    const t = setTimeout(async () => {
      setWSearching(true);
      try {
        const res = await fetch(
          `/api/item-types/search?q=${encodeURIComponent(wQuery)}&limit=10`,
        );
        if (res.ok) setWResults(await res.json());
        else setWResults([]);
      } catch {
        setWResults([]);
      } finally {
        setWSearching(false);
      }
    }, 300);
    return () => clearTimeout(t);
  }, [wQuery]);

  // Search blacklist items
  useEffect(() => {
    if (!bQuery.trim()) {
      setBResults([]);
      return;
    }
    const t = setTimeout(async () => {
      setBSearching(true);
      try {
        const res = await fetch(
          `/api/item-types/search?q=${encodeURIComponent(bQuery)}&limit=10`,
        );
        if (res.ok) setBResults(await res.json());
        else setBResults([]);
      } catch {
        setBResults([]);
      } finally {
        setBSearching(false);
      }
    }, 300);
    return () => clearTimeout(t);
  }, [bQuery]);

  const addToList = useCallback(
    async (list: "whitelist" | "blacklist", item: ItemSearchResult) => {
      try {
        const res = await fetch(`/api/arbiter/${list}`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ type_id: item.TypeID, name: item.Name }),
        });
        if (res.ok) {
          const newItem = { type_id: item.TypeID, name: item.Name };
          if (list === "whitelist") {
            setWhitelist((prev) =>
              prev.some((i) => i.type_id === item.TypeID)
                ? prev
                : [...prev, newItem],
            );
            setWQuery("");
            setWResults([]);
          } else {
            setBlacklist((prev) =>
              prev.some((i) => i.type_id === item.TypeID)
                ? prev
                : [...prev, newItem],
            );
            setBQuery("");
            setBResults([]);
          }
        }
      } catch {
        // ignore
      }
    },
    [],
  );

  const removeFromList = useCallback(
    async (list: "whitelist" | "blacklist", typeId: number) => {
      try {
        const res = await fetch(`/api/arbiter/${list}/${typeId}`, {
          method: "DELETE",
        });
        if (res.ok) {
          if (list === "whitelist") {
            setWhitelist((prev) => prev.filter((i) => i.type_id !== typeId));
          } else {
            setBlacklist((prev) => prev.filter((i) => i.type_id !== typeId));
          }
        }
      } catch {
        // ignore
      }
    },
    [],
  );

  const filteredWhitelist = whitelist.filter((i) =>
    !wSearch || i.name.toLowerCase().includes(wSearch.toLowerCase()),
  );
  const filteredBlacklist = blacklist.filter((i) =>
    !bSearch || i.name.toLowerCase().includes(bSearch.toLowerCase()),
  );

  if (status === "loading") return <Loading />;
  if (status !== "authenticated") return <Unauthorized />;
  if (loading) return <Loading />;

  return (
    <>
      <Head>
        <title>Arbiter Lists — pinky.tools</title>
      </Head>
      <Navbar />
      <div className="px-4 pb-8 bg-background-void min-h-screen">
        <div className="flex items-center gap-3 py-3">
          <Link
            href="/arbiter"
            className="text-text-muted hover:text-text-primary transition-colors"
          >
            <ArrowLeft className="h-5 w-5" />
          </Link>
          <h1 className="text-xl font-semibold text-text-heading">
            Arbiter Lists
          </h1>
          <span className="text-sm text-text-muted">
            Whitelist &amp; Blacklist Management
          </span>
        </div>

        <div className="grid grid-cols-2 gap-6">
          {/* Whitelist */}
          <div className="border border-overlay-subtle rounded-lg bg-background-panel">
            <div className="px-4 py-3 border-b border-overlay-subtle">
              <h2 className="text-sm font-semibold text-teal-success">
                Whitelist ({whitelist.length})
              </h2>
              <p className="text-xs text-text-muted mt-0.5">
                When "Use Whitelist" is enabled, only these items will appear in results.
              </p>
            </div>

            {/* Add to whitelist */}
            <div className="px-4 py-3 border-b border-overlay-subtle">
              <div className="relative">
                <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-text-muted pointer-events-none" />
                {wSearching && (
                  <Loader2 className="absolute right-2 top-1/2 -translate-y-1/2 h-3.5 w-3.5 animate-spin text-text-muted" />
                )}
                <Input
                  className="pl-8 pr-8 h-8 text-sm bg-background-elevated"
                  placeholder="Search items to add..."
                  value={wQuery}
                  onChange={(e) => setWQuery(e.target.value)}
                />
              </div>
              {wResults.length > 0 && (
                <div className="mt-1 border border-overlay-medium rounded bg-background-elevated max-h-40 overflow-y-auto">
                  {wResults.map((item) => (
                    <div
                      key={item.TypeID}
                      className="flex items-center justify-between px-3 py-1.5 hover:bg-interactive-hover cursor-pointer"
                      onClick={() => addToList("whitelist", item)}
                    >
                      <span className="text-sm text-text-primary">
                        {item.Name}
                      </span>
                      <span className="text-xs text-text-muted">
                        #{item.TypeID}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Filter existing */}
            <div className="px-4 py-2 border-b border-overlay-subtle">
              <div className="relative">
                <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-text-muted pointer-events-none" />
                <Input
                  className="pl-8 pr-8 h-7 text-xs bg-background-elevated"
                  placeholder="Filter list..."
                  value={wSearch}
                  onChange={(e) => setWSearch(e.target.value)}
                />
                {wSearch && (
                  <button
                    onClick={() => setWSearch("")}
                    className="absolute right-2 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-primary"
                  >
                    <X className="h-3.5 w-3.5" />
                  </button>
                )}
              </div>
            </div>

            <div className="overflow-y-auto max-h-[500px]">
              {filteredWhitelist.length === 0 ? (
                <div className="flex items-center justify-center py-12">
                  <p className="text-sm text-text-muted">
                    {whitelist.length === 0
                      ? "No items in whitelist"
                      : "No matches"}
                  </p>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow className="border-overlay-subtle hover:bg-transparent">
                      <TableHead className="py-2 text-xs text-text-muted">
                        Item
                      </TableHead>
                      <TableHead className="py-2 text-xs text-text-muted text-right">
                        Type ID
                      </TableHead>
                      <TableHead className="py-2 w-10" />
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredWhitelist.map((item) => (
                      <TableRow
                        key={item.type_id}
                        className="border-overlay-subtle hover:bg-interactive-hover"
                      >
                        <TableCell className="py-2 text-sm text-text-primary">
                          <span className="flex items-center gap-2">
                            <img
                              src={`https://images.evetech.net/types/${item.type_id}/icon?size=32`}
                              alt={item.name}
                              className="w-5 h-5 rounded flex-shrink-0"
                              onError={(e) => {
                                (e.target as HTMLImageElement).style.display =
                                  "none";
                              }}
                            />
                            {item.name}
                          </span>
                        </TableCell>
                        <TableCell className="py-2 text-xs text-text-muted text-right font-mono">
                          {item.type_id}
                        </TableCell>
                        <TableCell className="py-2 text-right">
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 text-text-muted hover:text-rose-danger"
                            onClick={() =>
                              removeFromList("whitelist", item.type_id)
                            }
                          >
                            <X className="h-3.5 w-3.5" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </div>
          </div>

          {/* Blacklist */}
          <div className="border border-overlay-subtle rounded-lg bg-background-panel">
            <div className="px-4 py-3 border-b border-overlay-subtle">
              <h2 className="text-sm font-semibold text-rose-danger">
                Blacklist ({blacklist.length})
              </h2>
              <p className="text-xs text-text-muted mt-0.5">
                When "Use Blacklist" is enabled, these items will be hidden from results.
              </p>
            </div>

            {/* Add to blacklist */}
            <div className="px-4 py-3 border-b border-overlay-subtle">
              <div className="relative">
                <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-text-muted pointer-events-none" />
                {bSearching && (
                  <Loader2 className="absolute right-2 top-1/2 -translate-y-1/2 h-3.5 w-3.5 animate-spin text-text-muted" />
                )}
                <Input
                  className="pl-8 pr-8 h-8 text-sm bg-background-elevated"
                  placeholder="Search items to add..."
                  value={bQuery}
                  onChange={(e) => setBQuery(e.target.value)}
                />
              </div>
              {bResults.length > 0 && (
                <div className="mt-1 border border-overlay-medium rounded bg-background-elevated max-h-40 overflow-y-auto">
                  {bResults.map((item) => (
                    <div
                      key={item.TypeID}
                      className="flex items-center justify-between px-3 py-1.5 hover:bg-interactive-hover cursor-pointer"
                      onClick={() => addToList("blacklist", item)}
                    >
                      <span className="text-sm text-text-primary">
                        {item.Name}
                      </span>
                      <span className="text-xs text-text-muted">
                        #{item.TypeID}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Filter existing */}
            <div className="px-4 py-2 border-b border-overlay-subtle">
              <div className="relative">
                <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-text-muted pointer-events-none" />
                <Input
                  className="pl-8 pr-8 h-7 text-xs bg-background-elevated"
                  placeholder="Filter list..."
                  value={bSearch}
                  onChange={(e) => setBSearch(e.target.value)}
                />
                {bSearch && (
                  <button
                    onClick={() => setBSearch("")}
                    className="absolute right-2 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-primary"
                  >
                    <X className="h-3.5 w-3.5" />
                  </button>
                )}
              </div>
            </div>

            <div className="overflow-y-auto max-h-[500px]">
              {filteredBlacklist.length === 0 ? (
                <div className="flex items-center justify-center py-12">
                  <p className="text-sm text-text-muted">
                    {blacklist.length === 0
                      ? "No items in blacklist"
                      : "No matches"}
                  </p>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow className="border-overlay-subtle hover:bg-transparent">
                      <TableHead className="py-2 text-xs text-text-muted">
                        Item
                      </TableHead>
                      <TableHead className="py-2 text-xs text-text-muted text-right">
                        Type ID
                      </TableHead>
                      <TableHead className="py-2 w-10" />
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredBlacklist.map((item) => (
                      <TableRow
                        key={item.type_id}
                        className="border-overlay-subtle hover:bg-interactive-hover"
                      >
                        <TableCell className="py-2 text-sm text-text-primary">
                          <span className="flex items-center gap-2">
                            <img
                              src={`https://images.evetech.net/types/${item.type_id}/icon?size=32`}
                              alt={item.name}
                              className="w-5 h-5 rounded flex-shrink-0"
                              onError={(e) => {
                                (e.target as HTMLImageElement).style.display =
                                  "none";
                              }}
                            />
                            {item.name}
                          </span>
                        </TableCell>
                        <TableCell className="py-2 text-xs text-text-muted text-right font-mono">
                          {item.type_id}
                        </TableCell>
                        <TableCell className="py-2 text-right">
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 text-text-muted hover:text-rose-danger"
                            onClick={() =>
                              removeFromList("blacklist", item.type_id)
                            }
                          >
                            <X className="h-3.5 w-3.5" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
