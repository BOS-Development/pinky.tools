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
    case "planned": return "bg-[rgba(0,212,255,0.1)] border-[rgba(0,212,255,0.3)] text-[#00d4ff]";
    case "active": return "bg-[rgba(16,185,129,0.1)] border-[rgba(16,185,129,0.3)] text-[#10b981]";
    case "completed": return "bg-[rgba(148,163,184,0.1)] border-[rgba(148,163,184,0.3)] text-[#94a3b8]";
    case "cancelled": return "bg-[rgba(239,68,68,0.1)] border-[rgba(239,68,68,0.3)] text-[#ef4444]";
    default: return "bg-[rgba(148,163,184,0.1)] border-[rgba(148,163,184,0.3)] text-[#94a3b8]";
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
        <Loader2 className="h-8 w-8 animate-spin text-[#00d4ff]" />
      </div>
    );
  }

  return (
    <TooltipProvider>
      <div className="overflow-x-auto rounded-sm border border-[rgba(148,163,184,0.1)]">
        <Table>
          <TableHeader>
            <TableRow className="bg-[#0f1219]">
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
                <TableCell colSpan={16} className="text-center py-4 text-[#64748b]">
                  No jobs in queue
                </TableCell>
              </TableRow>
            ) : (
              entries.map((entry, idx) => (
                <TableRow
                  key={entry.id}
                  className={idx % 2 === 0 ? "bg-[#0d1117]" : "bg-[#12151f]"}
                >
                  {entry.activity === "transport" ? (
                    <>
                      <TableCell>
                        <span className="text-sm text-[#e2e8f0]">
                          {entry.transportOriginName && entry.transportDestName
                            ? `${entry.transportOriginName} → ${entry.transportDestName}`
                            : "-"}
                        </span>
                      </TableCell>
                      <TableCell>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <span className="text-sm text-[#cbd5e1] max-w-[200px] overflow-hidden text-ellipsis whitespace-nowrap block">
                              {entry.transportItemsSummary || "-"}
                            </span>
                          </TooltipTrigger>
                          <TooltipContent>{entry.transportItemsSummary || ""}</TooltipContent>
                        </Tooltip>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-[#cbd5e1]">
                          {entry.transportMethod ? formatTransportMethod(entry.transportMethod) : "Transport"}
                        </span>
                      </TableCell>
                      <TableCell className="text-right">
                        <span className="text-sm text-[#94a3b8]">
                          {entry.transportJumps ? `${formatNumber(entry.transportJumps)} jumps` : "-"}
                        </span>
                      </TableCell>
                      <TableCell className="text-right">
                        <span className="text-sm text-[#94a3b8]">
                          {entry.transportVolumeM3 ? `${formatNumber(entry.transportVolumeM3)} m³` : "-"}
                        </span>
                      </TableCell>
                      <TableCell><span className="text-sm text-[#64748b]">-</span></TableCell>
                      <TableCell><span className="text-sm text-[#64748b]">-</span></TableCell>
                      <TableCell><span className="text-sm text-[#64748b]">-</span></TableCell>
                      <TableCell>
                        <span className="text-sm text-[#94a3b8]">
                          {entry.transportFulfillment ? formatFulfillment(entry.transportFulfillment) : "-"}
                        </span>
                      </TableCell>
                      <TableCell className="text-right">
                        <span className="text-sm text-[#cbd5e1]">
                          {entry.estimatedCost ? formatISK(entry.estimatedCost) : "-"}
                        </span>
                      </TableCell>
                      <TableCell><span className="text-sm text-[#64748b]">-</span></TableCell>
                      <TableCell><span className="text-sm text-[#64748b]">-</span></TableCell>
                      <TableCell className="text-center"><span className="text-sm text-[#64748b]">-</span></TableCell>
                    </>
                  ) : (
                    <>
                      <TableCell>
                        <span className="text-sm text-[#e2e8f0]">
                          {entry.blueprintName || `Type ${entry.blueprintTypeId}`}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-[#cbd5e1]">
                          {entry.productName || "-"}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-[#cbd5e1] capitalize">
                          {entry.activity}
                        </span>
                      </TableCell>
                      <TableCell className="text-right">
                        <span className="text-sm text-[#e2e8f0]">
                          {formatNumber(entry.runs)}
                        </span>
                      </TableCell>
                      <TableCell className="text-right">
                        <span className="text-sm text-[#94a3b8]">
                          {entry.meLevel}/{entry.teLevel}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-[#94a3b8]">
                          {entry.stationName || "-"}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-[#94a3b8]">
                          {entry.inputLocation || "-"}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-[#94a3b8]">
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
                                    ? "bg-[rgba(0,212,255,0.1)] border-[rgba(0,212,255,0.3)] text-[#60a5fa]"
                                    : "bg-transparent border-[#334155] text-[#94a3b8]"
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
                                  <span className="ml-1 text-xs text-[#64748b]">
                                    ({char.mfgSlotsUsed}/{char.mfgSlotsMax} mfg)
                                  </span>
                                </DropdownMenuItem>
                              ))}
                              {eligibleCharacters.length === 0 && (
                                <DropdownMenuItem disabled>
                                  <span className="text-[#64748b]">No characters available</span>
                                </DropdownMenuItem>
                              )}
                              <DropdownMenuSeparator />
                              <DropdownMenuItem onClick={() => handleReassign(entry.id, null)}>
                                <em>Unassign</em>
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        ) : (
                          <span className="text-sm text-[#94a3b8]">
                            {entry.characterName || "-"}
                          </span>
                        )}
                      </TableCell>
                      <TableCell className="text-right">
                        <span className="text-sm text-[#cbd5e1]">
                          {entry.estimatedCost ? formatISK(entry.estimatedCost) : "-"}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-[#94a3b8]">
                          {entry.estimatedDuration ? formatDuration(entry.estimatedDuration) : "-"}
                        </span>
                      </TableCell>
                      <TableCell>
                        {entry.esiJobEndDate ? (
                          <div>
                            <span className="text-sm text-[#00d4ff] font-mono font-semibold">
                              {formatTimeRemaining(entry.esiJobEndDate)}
                            </span>
                            <span className="block text-xs text-[#64748b]">
                              {formatEndDate(entry.esiJobEndDate)}
                            </span>
                          </div>
                        ) : (
                          <span className="text-sm text-[#64748b]">-</span>
                        )}
                      </TableCell>
                      <TableCell className="text-center">
                        {entry.esiJobSource ? (
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <span className="inline-flex">
                                {entry.esiJobSource === "corporation" ? (
                                  <Building2 className="h-[18px] w-[18px] text-[#f59e0b]" />
                                ) : (
                                  <User className="h-[18px] w-[18px] text-[#94a3b8]" />
                                )}
                              </span>
                            </TooltipTrigger>
                            <TooltipContent>
                              {entry.esiJobSource === "corporation" ? "Corporation Job" : "Character Job"}
                            </TooltipContent>
                          </Tooltip>
                        ) : (
                          <span className="text-sm text-[#64748b]">-</span>
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
                    <span className="text-sm text-[#94a3b8] max-w-[150px] overflow-hidden text-ellipsis whitespace-nowrap block">
                      {entry.notes || "-"}
                    </span>
                  </TableCell>
                  <TableCell className="text-center">
                    {(entry.status === "planned" || entry.status === "active") && (
                      <button
                        className="p-1 rounded hover:bg-[rgba(239,68,68,0.1)] text-[#ef4444]"
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
