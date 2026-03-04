import { useState, useEffect } from 'react';
import { useSession } from 'next-auth/react';
import { Loader2 } from 'lucide-react';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';

type ActivitySlotInfo = {
  activityType: string;
  slotsMax: number;
  slotsInUse: number;
  slotsReserved: number;
  slotsAvailable: number;
  slotsListed: number;
};

type CharacterSlotInventory = {
  characterId: number;
  characterName: string;
  slotsByActivity: Record<string, ActivitySlotInfo>;
};

const ACTIVITY_LABELS: Record<string, string> = {
  manufacturing: 'Manufacturing',
  reaction: 'Reactions',
  copying: 'Copying',
  invention: 'Invention',
  me_research: 'ME Research',
  te_research: 'TE Research',
};

export default function SlotInventoryPanel() {
  const { data: session } = useSession();
  const [inventory, setInventory] = useState<CharacterSlotInventory[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (session) {
      fetchInventory();
    }
  }, [session]);

  const fetchInventory = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/job-slots/inventory');
      if (response.ok) {
        const data = await response.json();
        setInventory(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch slot inventory:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-[#00d4ff]" />
      </div>
    );
  }

  if (inventory.length === 0) {
    return (
      <div className="bg-[#12151f] rounded-sm border border-[rgba(148,163,184,0.1)] p-8 text-center">
        <h3 className="text-lg font-semibold text-[#94a3b8]">No slot data available</h3>
        <p className="text-sm text-[#64748b] mt-1">
          Make sure you have characters with industry skills synced.
        </p>
      </div>
    );
  }

  return (
    <div className="overflow-x-auto rounded-sm border border-[rgba(148,163,184,0.1)]">
      <Table>
        <TableHeader>
          <TableRow className="bg-[#0f1219]">
            <TableHead>Character</TableHead>
            <TableHead>Activity Type</TableHead>
            <TableHead className="text-right">Max Slots</TableHead>
            <TableHead className="text-right">In Use</TableHead>
            <TableHead className="text-right">Reserved</TableHead>
            <TableHead className="text-right">Listed</TableHead>
            <TableHead className="text-right">Available</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {inventory.map((char) => (
            Object.entries(char.slotsByActivity).map(([activityType, slotInfo]) => (
              <TableRow key={`${char.characterId}-${activityType}`} className="hover:bg-[rgba(0,212,255,0.04)]">
                <TableCell>
                  <span className="text-sm font-medium text-[#e2e8f0]">
                    {char.characterName}
                  </span>
                </TableCell>
                <TableCell>
                  <Badge className="bg-[rgba(0,212,255,0.1)] border border-[rgba(0,212,255,0.3)] text-[#60a5fa] hover:bg-[rgba(0,212,255,0.15)] cursor-default">
                    {ACTIVITY_LABELS[activityType] || activityType}
                  </Badge>
                </TableCell>
                <TableCell className="text-right">{slotInfo.slotsMax}</TableCell>
                <TableCell className="text-right">{slotInfo.slotsInUse}</TableCell>
                <TableCell className="text-right">{slotInfo.slotsReserved}</TableCell>
                <TableCell className="text-right">{slotInfo.slotsListed}</TableCell>
                <TableCell className="text-right">
                  <span className={`font-semibold ${slotInfo.slotsAvailable > 0 ? 'text-[#10b981]' : 'text-[#ef4444]'}`}>
                    {slotInfo.slotsAvailable}
                  </span>
                </TableCell>
              </TableRow>
            ))
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
