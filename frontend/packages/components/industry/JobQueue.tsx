import { useState } from "react";
import { CharacterSlotInfo, IndustryJobQueueEntry } from "@industry-tool/client/data/models";
import { formatISK, formatNumber } from "@industry-tool/utils/formatting";
import { Loader2, XCircle, User, Building2 } from "lucide-react";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider } from "@/components/ui/tooltip";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger, DropdownMenuSeparator } from "@/components/ui/dropdown-menu";
import { Separator } from "@/components/ui/separator";

type Props = {
  entries: IndustryJobQueueEntry[];
  loading: boolean;
  onCancel: (id: number) => void;
  onRefresh?: () => void;
};

function getStatusClasses(status: string): string {
  switch (status) {
    case "planned": return "bg-interactive-selected border-border-active text-primary";
    case "active": return "bg-teal-success/10 border-teal-success/30 text-teal-success";
    case "completed": return "bg-overlay-subtle border-overlay-strong text-text-secondary";
    case "cancelled": return "bg-rose-danger/10 border-rose-danger/30 text-rose-danger";
    default: return "bg-overlay-subtle border-overlay-strong text-text-secondary";
  }
}

function formatTransportMethod(method: string): string {
  switch (method) {
    case "freighter": return "Freighter";
    case "jump_freighter": return "Jump Freighter";
    case "dst": return "DST";
    case "blockade_runner": return "Blockade Runner";
    default: return method;
  }
}

function formatFulfillment(type: string): string {
  switch (type) {
    case "self_haul": return "Self Haul";
    case "courier_contract": return "Courier Contract";
    case "contact_haul": return "Contact Haul";
    default: return type;
  }
}

function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);

  if (hours > 24) {
    const days = Math.floor(hours / 24);
    const remHours = hours % 24;
    return `${days}d ${remHours}h`;
  }

  return `${hours}h ${minutes}m`;
}

function formatTimeRemaining(endDateStr: string): string {
  const end = new Date(endDateStr);
  const now = new Date();
  const diffMs = end.getTime() - now.getTime();

  if (diffMs <= 0) return "Ready";

  const totalSecs = Math.floor(diffMs / 1000);
  const days = Math.floor(totalSecs / 86400);
  const hours = Math.floor((totalSecs % 86400) / 3600);
  const minutes = Math.floor((totalSecs % 3600) / 60);
  const seconds = totalSecs % 60;

  const pad = (n: number) => n.toString().padStart(2, "0");

  if (days > 0) {
    return `${days}D ${pad(hours)}:${pad(minutes)}:${pad(seconds)}`;
  }
  return `${pad(hours)}:${pad(minutes)}:${pad(seconds)}`;
}

function formatEndDate(endDateStr: string): string {
  const d = new Date(endDateStr);
  const year = d.getUTCFullYear();
  const month = (d.getUTCMonth() + 1).toString().padStart(2, "0");
  const day = d.getUTCDate().toString().padStart(2, "0");
  const hours = d.getUTCHours().toString().padStart(2, "0");
  const minutes = d.getUTCMinutes().toString().padStart(2, "0");
  return `${year}.${month}.${day} ${hours}:${minutes}`;
}

export default function JobQueue({ entries, loading, onCancel, onRefresh }: Props) {
  const [eligibleCharacters, setEligibleCharacters] = useState<CharacterSlotInfo[]>([]);
  const [reassignLoading, setReassignLoading] = useState(false);

  const handleOpenReassign = async () => {
    if (eligibleCharacters.length === 0) {
      try {
        const res = await fetch("/api/industry/character-slots");
        if (res.ok) {
          const data = await res.json();
          setEligibleCharacters(data || []);
        }
      } catch (err) {
        console.error("Failed to fetch character slots:", err);
      }
    }
  };

  const handleReassign = async (entryId: number, characterId: number | null) => {
    setReassignLoading(true);
    try {
      await fetch(`/api/industry/queue/${entryId}/character`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ characterId }),
      });
      if (onRefresh) onRefresh();
    } catch (err) {
      console.error("Failed to reassign character:", err);
    } finally {
      setReassignLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <TooltipProvider>
      <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
        <Table>
          <TableHeader>
            <TableRow className="bg-background-void">
              <TableHead>Blueprint</TableHead>
              <TableHead>Product</TableHead>
              <TableHead>Activity</TableHead>
              <TableHead className="text-right">Runs</TableHead>
              <TableHead className="text-right">ME/TE</TableHead>
              <TableHead>Station</TableHead>
              <TableHead>Input</TableHead>
              <TableHead>Output</TableHead>
              <TableHead>Character</TableHead>
              <TableHead className="text-right">Est. Cost</TableHead>
              <TableHead>Duration</TableHead>
              <TableHead>Finishes</TableHead>
              <TableHead className="text-center">Source</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Notes</TableHead>
              <TableHead className="text-center">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {entries.length === 0 ? (
              <TableRow>
                <TableCell colSpan={16} className="text-center py-4 text-text-muted">
                  No jobs in queue
                </TableCell>
              </TableRow>
            ) : (
              entries.map((entry, idx) => (
                <TableRow
                  key={entry.id}
                  className={idx % 2 === 0 ? "bg-background-void" : "bg-background-panel"}
                >
                  {entry.activity === "transport" ? (
                    <>
                      <TableCell>
                        <span className="text-sm text-text-emphasis">
                          {entry.transportOriginName && entry.transportDestName
                            ? `${entry.transportOriginName} → ${entry.transportDestName}`
                            : "-"}
                        </span>
                      </TableCell>
                      <TableCell>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <span className="text-sm text-text-primary max-w-[200px] overflow-hidden text-ellipsis whitespace-nowrap block">
                              {entry.transportItemsSummary || "-"}
                            </span>
                          </TooltipTrigger>
                          <TooltipContent>{entry.transportItemsSummary || ""}</TooltipContent>
                        </Tooltip>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-text-primary">
                          {entry.transportMethod ? formatTransportMethod(entry.transportMethod) : "Transport"}
                        </span>
                      </TableCell>
                      <TableCell className="text-right">
                        <span className="text-sm text-text-secondary">
                          {entry.transportJumps ? `${formatNumber(entry.transportJumps)} jumps` : "-"}
                        </span>
                      </TableCell>
                      <TableCell className="text-right">
                        <span className="text-sm text-text-secondary">
                          {entry.transportVolumeM3 ? `${formatNumber(entry.transportVolumeM3)} m³` : "-"}
                        </span>
                      </TableCell>
                      <TableCell><span className="text-sm text-text-muted">-</span></TableCell>
                      <TableCell><span className="text-sm text-text-muted">-</span></TableCell>
                      <TableCell><span className="text-sm text-text-muted">-</span></TableCell>
                      <TableCell>
                        <span className="text-sm text-text-secondary">
                          {entry.transportFulfillment ? formatFulfillment(entry.transportFulfillment) : "-"}
                        </span>
                      </TableCell>
                      <TableCell className="text-right">
                        <span className="text-sm text-text-primary">
                          {entry.estimatedCost ? formatISK(entry.estimatedCost) : "-"}
                        </span>
                      </TableCell>
                      <TableCell><span className="text-sm text-text-muted">-</span></TableCell>
                      <TableCell><span className="text-sm text-text-muted">-</span></TableCell>
                      <TableCell className="text-center"><span className="text-sm text-text-muted">-</span></TableCell>
                    </>
                  ) : (
                    <>
                      <TableCell>
                        <div className="flex items-center gap-1.5">
                          <img
                            src={`https://images.evetech.net/types/${entry.blueprintTypeId}/icon?size=32`}
                            alt=""
                            width={24}
                            height={24}
                            className="flex-shrink-0"
                            loading="lazy"
                            style={{ filter: "sepia(1) saturate(3) hue-rotate(180deg)" }}
                          />
                          <span className="text-sm text-text-emphasis">
                            {entry.blueprintName || `Type ${entry.blueprintTypeId}`}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1.5">
                          {entry.productTypeId && (
                            <img
                              src={`https://images.evetech.net/types/${entry.productTypeId}/icon?size=32`}
                              alt=""
                              width={24}
                              height={24}
                              className="flex-shrink-0"
                              loading="lazy"
                            />
                          )}
                          <span className="text-sm text-text-primary">
                            {entry.productName || "-"}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-text-primary capitalize">
                          {entry.activity}
                        </span>
                      </TableCell>
                      <TableCell className="text-right">
                        <span className="text-sm text-text-emphasis">
                          {formatNumber(entry.runs)}
                        </span>
                      </TableCell>
                      <TableCell className="text-right">
                        <span className="text-sm text-text-secondary">
                          {entry.meLevel}/{entry.teLevel}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-text-secondary">
                          {entry.stationName || "-"}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-text-secondary">
                          {entry.inputLocation || "-"}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-text-secondary">
                          {entry.outputLocation || "-"}
                        </span>
                      </TableCell>
                      <TableCell>
                        {entry.status === "planned" ? (
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <button
                                className={`inline-flex items-center rounded-sm px-2 py-0.5 text-xs border cursor-pointer ${
                                  entry.characterName
                                    ? "bg-interactive-selected border-border-active text-blue-science"
                                    : "bg-transparent border-overlay-medium text-text-secondary"
                                }`}
                                onClick={handleOpenReassign}
                                disabled={reassignLoading}
                              >
                                {entry.characterName || "Assign"}
                              </button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent>
                              {eligibleCharacters.map((char) => (
                                <DropdownMenuItem key={char.characterId} onClick={() => handleReassign(entry.id, char.characterId)}>
                                  {char.characterName}
                                  <span className="ml-1 text-xs text-text-muted">
                                    ({char.mfgSlotsUsed}/{char.mfgSlotsMax} mfg)
                                  </span>
                                </DropdownMenuItem>
                              ))}
                              {eligibleCharacters.length === 0 && (
                                <DropdownMenuItem disabled>
                                  <span className="text-text-muted">No characters available</span>
                                </DropdownMenuItem>
                              )}
                              <DropdownMenuSeparator />
                              <DropdownMenuItem onClick={() => handleReassign(entry.id, null)}>
                                <em>Unassign</em>
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        ) : (
                          <span className="text-sm text-text-secondary">
                            {entry.characterName || "-"}
                          </span>
                        )}
                      </TableCell>
                      <TableCell className="text-right">
                        <span className="text-sm text-text-primary">
                          {entry.estimatedCost ? formatISK(entry.estimatedCost) : "-"}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-text-secondary">
                          {entry.estimatedDuration ? formatDuration(entry.estimatedDuration) : "-"}
                        </span>
                      </TableCell>
                      <TableCell>
                        {entry.esiJobEndDate ? (
                          <div>
                            <span className="text-sm text-primary font-mono font-semibold">
                              {formatTimeRemaining(entry.esiJobEndDate)}
                            </span>
                            <span className="block text-xs text-text-muted">
                              {formatEndDate(entry.esiJobEndDate)}
                            </span>
                          </div>
                        ) : (
                          <span className="text-sm text-text-muted">-</span>
                        )}
                      </TableCell>
                      <TableCell className="text-center">
                        {entry.esiJobSource ? (
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <span className="inline-flex">
                                {entry.esiJobSource === "corporation" ? (
                                  <Building2 className="h-[18px] w-[18px] text-amber-manufacturing" />
                                ) : (
                                  <User className="h-[18px] w-[18px] text-text-secondary" />
                                )}
                              </span>
                            </TooltipTrigger>
                            <TooltipContent>
                              {entry.esiJobSource === "corporation" ? "Corporation Job" : "Character Job"}
                            </TooltipContent>
                          </Tooltip>
                        ) : (
                          <span className="text-sm text-text-muted">-</span>
                        )}
                      </TableCell>
                    </>
                  )}
                  <TableCell>
                    <Badge className={`border ${getStatusClasses(entry.status)} cursor-default`}>
                      {entry.status}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <span className="text-sm text-text-secondary max-w-[150px] overflow-hidden text-ellipsis whitespace-nowrap block">
                      {entry.notes || "-"}
                    </span>
                  </TableCell>
                  <TableCell className="text-center">
                    {(entry.status === "planned" || entry.status === "active") && (
                      <button
                        className="p-1 rounded hover:bg-rose-danger/10 text-rose-danger"
                        onClick={() => onCancel(entry.id)}
                        title="Cancel job"
                      >
                        <XCircle className="h-4 w-4" />
                      </button>
                    )}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </TooltipProvider>
  );
}
