import { useState, useEffect, useCallback } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Loader2, Trash2, ScanSearch, AlertTriangle, Plus } from 'lucide-react';
import { UserTradingStructure } from '@industry-tool/client/data/models';
import { toast } from '@/components/ui/sonner';

interface UserStructuresDialogProps {
  open: boolean;
  onClose: () => void;
  onStructuresChanged: () => void;
}

export default function UserStructuresDialog({
  open,
  onClose,
  onStructuresChanged,
}: UserStructuresDialogProps) {
  const [structures, setStructures] = useState<UserTradingStructure[]>([]);
  const [loading, setLoading] = useState(false);
  const [characters, setCharacters] = useState<Array<{id: number; name: string}>>([]);
  const [loadingCharacters, setLoadingCharacters] = useState(false);
  const [assetStructures, setAssetStructures] = useState<Array<{structureId: number; name: string}>>([]);
  const [loadingAssetStructures, setLoadingAssetStructures] = useState(false);
  const [selectedCharacterId, setSelectedCharacterId] = useState('');
  const [selectedStructureId, setSelectedStructureId] = useState('');
  const [adding, setAdding] = useState(false);
  const [addError, setAddError] = useState<string | null>(null);
  const [scanningId, setScanningId] = useState<number | null>(null);
  const [deletingId, setDeletingId] = useState<number | null>(null);

  const fetchStructures = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch('/api/hauling/structures');
      if (res.ok) {
        const data = await res.json();
        setStructures(Array.isArray(data) ? data : []);
      }
    } catch (err) {
      console.error('Failed to fetch structures:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchCharacters = useCallback(async () => {
    setLoadingCharacters(true);
    try {
      const res = await fetch('/api/characters');
      if (res.ok) {
        const data = await res.json();
        setCharacters(Array.isArray(data) ? data : []);
      }
    } catch (err) {
      console.error('Failed to fetch characters:', err);
    } finally {
      setLoadingCharacters(false);
    }
  }, []);

  const fetchAssetStructures = useCallback(async (characterId: string) => {
    if (!characterId) {
      setAssetStructures([]);
      return;
    }
    setLoadingAssetStructures(true);
    setSelectedStructureId('');
    try {
      const res = await fetch(`/api/hauling/character-asset-structures?characterId=${characterId}`);
      if (res.ok) {
        const data = await res.json();
        setAssetStructures(Array.isArray(data) ? data : []);
      }
    } catch (err) {
      console.error('Failed to fetch asset structures:', err);
    } finally {
      setLoadingAssetStructures(false);
    }
  }, []);

  useEffect(() => {
    if (open) {
      fetchStructures();
      fetchCharacters();
    } else {
      setSelectedCharacterId('');
      setSelectedStructureId('');
      setAssetStructures([]);
      setAddError(null);
    }
  }, [open, fetchStructures, fetchCharacters]);

  const handleAdd = async () => {
    if (!selectedCharacterId || !selectedStructureId) return;
    setAdding(true);
    setAddError(null);
    try {
      const res = await fetch('/api/hauling/structures', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          structureId: Number(selectedStructureId),
          characterId: Number(selectedCharacterId),
        }),
      });

      const data = await res.json();

      if (res.status === 403 || data?.accessOk === false) {
        setAddError(
          data?.error ||
            'Unable to access structure market (403). Character may not have docking rights.',
        );
      } else if (!res.ok) {
        setAddError(data?.error || 'Failed to add structure.');
      } else {
        setSelectedStructureId('');
        setSelectedCharacterId('');
        setAssetStructures([]);
        await fetchStructures();
        onStructuresChanged();
        toast.success('Structure added successfully.');
      }
    } catch (err) {
      console.error('Failed to add structure:', err);
      setAddError('An unexpected error occurred.');
    } finally {
      setAdding(false);
    }
  };

  const handleDelete = async (id: number) => {
    setDeletingId(id);
    try {
      const res = await fetch(`/api/hauling/structures?id=${id}`, {
        method: 'DELETE',
      });

      if (res.ok || res.status === 204) {
        setStructures((prev) => prev.filter((s) => s.id !== id));
        onStructuresChanged();
        toast.success('Structure removed.');
      } else {
        toast.error('Failed to remove structure.');
      }
    } catch (err) {
      console.error('Failed to delete structure:', err);
      toast.error('An unexpected error occurred.');
    } finally {
      setDeletingId(null);
    }
  };

  const handleScan = async (structure: UserTradingStructure) => {
    setScanningId(structure.id);
    try {
      const res = await fetch('/api/hauling/structure-scan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: structure.id }),
      });

      const data = await res.json();

      if (res.ok && data?.accessOk !== false) {
        // Refresh structures to get updated lastScannedAt and accessOk
        await fetchStructures();
        onStructuresChanged();
        toast.success(`Market scan complete for ${structure.name}.`);
      } else {
        await fetchStructures();
        toast.error(
          data?.error ||
            'Scan failed. Character may not have docking rights.',
        );
      }
    } catch (err) {
      console.error('Failed to scan structure:', err);
      toast.error('An unexpected error occurred during scan.');
    } finally {
      setScanningId(null);
    }
  };

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) onClose(); }}>
      <DialogContent className="max-w-3xl bg-background-panel border-overlay-medium">
        <DialogHeader>
          <DialogTitle className="text-text-heading">Manage Trading Structures</DialogTitle>
        </DialogHeader>

        {/* Add Structure Form */}
        <div className="border border-overlay-subtle rounded-md p-4 mb-4 bg-background-void">
          <p className="text-sm font-semibold text-text-emphasis mb-3">Add Structure</p>
          <div className="flex gap-3 flex-wrap items-end">
            <div className="flex-1 min-w-[160px]">
              <Label className="text-xs text-text-secondary mb-1 block">Character</Label>
              {loadingCharacters ? (
                <div className="flex items-center gap-2 text-text-muted text-sm">
                  <Loader2 className="h-4 w-4 animate-spin" /> Loading characters...
                </div>
              ) : (
                <Select
                  value={selectedCharacterId}
                  onValueChange={(v) => {
                    setSelectedCharacterId(v);
                    fetchAssetStructures(v);
                  }}
                >
                  <SelectTrigger className="bg-background-elevated border-overlay-strong text-text-emphasis">
                    <SelectValue placeholder="Select character" />
                  </SelectTrigger>
                  <SelectContent>
                    {characters.map((char) => (
                      <SelectItem key={char.id} value={String(char.id)}>
                        {char.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </div>
            <div className="flex-1 min-w-[200px]">
              <Label className="text-xs text-text-secondary mb-1 block">Structure</Label>
              {loadingAssetStructures ? (
                <div className="flex items-center gap-2 text-text-muted text-sm">
                  <Loader2 className="h-4 w-4 animate-spin" /> Loading structures...
                </div>
              ) : !selectedCharacterId ? (
                <p className="text-xs text-text-muted italic">Select a character first</p>
              ) : assetStructures.length === 0 ? (
                <p className="text-xs text-text-muted italic">No structures found in assets</p>
              ) : (
                <Select
                  value={selectedStructureId}
                  onValueChange={setSelectedStructureId}
                  disabled={!selectedCharacterId}
                >
                  <SelectTrigger className="bg-background-elevated border-overlay-strong text-text-emphasis">
                    <SelectValue placeholder="Select structure" />
                  </SelectTrigger>
                  <SelectContent>
                    {assetStructures.map((s) => (
                      <SelectItem key={s.structureId} value={String(s.structureId)}>
                        {s.name || `Structure ${s.structureId}`}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </div>
            <Button
              onClick={handleAdd}
              disabled={adding || !selectedCharacterId || !selectedStructureId}
              className="shrink-0"
            >
              {adding ? (
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              ) : (
                <Plus className="h-4 w-4 mr-2" />
              )}
              Add
            </Button>
          </div>
          {addError && (
            <div className="mt-3 flex items-start gap-2 text-sm text-rose-500">
              <AlertTriangle className="h-4 w-4 shrink-0 mt-0.5" />
              <span>{addError}</span>
            </div>
          )}
        </div>

        {/* Structures List */}
        {loading ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin text-text-muted" />
          </div>
        ) : structures.length === 0 ? (
          <p className="text-center text-text-secondary text-sm py-6">
            No trading structures configured. Add one above to get started.
          </p>
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="bg-background-void border-overlay-subtle">
                  <TableHead className="font-bold text-text-emphasis">Name</TableHead>
                  <TableHead className="font-bold text-text-emphasis">Status</TableHead>
                  <TableHead className="font-bold text-text-emphasis">Last Scanned</TableHead>
                  <TableHead className="font-bold text-text-emphasis">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {structures.map((structure) => (
                  <TableRow key={structure.id} className="border-overlay-subtle">
                    <TableCell>
                      <div>
                        <p className="text-sm font-semibold text-text-emphasis">
                          {structure.name}
                        </p>
                        <p className="text-xs text-text-muted mt-0.5">
                          ID: {structure.structureId}
                        </p>
                      </div>
                    </TableCell>
                    <TableCell>
                      {structure.accessOk ? (
                        <Badge
                          variant="outline"
                          className="border-teal-success/30 text-teal-500 bg-status-success-tint"
                        >
                          Access OK
                        </Badge>
                      ) : (
                        <Badge
                          variant="outline"
                          className="border-rose-danger/30 text-rose-500 bg-status-error-tint"
                        >
                          <AlertTriangle className="h-3 w-3 mr-1" />
                          No Access
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-sm text-text-secondary">
                      {structure.lastScannedAt
                        ? new Date(structure.lastScannedAt).toLocaleString()
                        : 'Never'}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleScan(structure)}
                          disabled={scanningId === structure.id}
                          title="Scan market"
                        >
                          {scanningId === structure.id ? (
                            <Loader2 className="h-3.5 w-3.5 mr-1 animate-spin" />
                          ) : (
                            <ScanSearch className="h-3.5 w-3.5 mr-1" />
                          )}
                          Scan
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => handleDelete(structure.id)}
                          disabled={deletingId === structure.id}
                          className="text-rose-500 hover:text-rose-400"
                          title="Remove structure"
                        >
                          {deletingId === structure.id ? (
                            <Loader2 className="h-3.5 w-3.5 animate-spin" data-testid="DeleteIcon" />
                          ) : (
                            <Trash2 className="h-3.5 w-3.5" data-testid="DeleteIcon" />
                          )}
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
