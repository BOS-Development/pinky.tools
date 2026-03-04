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
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (inventory.length === 0) {
    return (
      <div className="bg-background-panel rounded-sm border border-overlay-subtle p-8 text-center">
        <h3 className="text-lg font-semibold text-text-secondary">No slot data available</h3>
        <p className="text-sm text-text-muted mt-1">
          Make sure you have characters with industry skills synced.
        </p>
      </div>
    );
  }

  return (
    <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
      <Table>
        <TableHeader>
          <TableRow className="bg-background-void">
            <TableHead>Character</TableHead>
            <TableHead className="w-[1%] whitespace-nowrap">Activity Type</TableHead>
            <TableHead className="text-right">Max Slots</TableHead>
            <TableHead className="text-right">In Use</TableHead>
            <TableHead className="text-right">Reserved</TableHead>
            <TableHead className="text-right">Listed</TableHead>
            <TableHead className="text-right">Available</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {inventory.map((char) => {
            const activities = Object.entries(char.slotsByActivity)
              .filter(([, info]) => info.slotsMax > 0);
            if (activities.length === 0) return null;
            return activities.map(([activityType, slotInfo], idx) => (
              <TableRow key={`${char.characterId}-${activityType}`} className="hover:bg-interactive-hover">
                <TableCell>
                  {idx === 0 && (
                    <span className="text-sm font-medium text-text-emphasis">
                      {char.characterName}
                    </span>
                  )}
                </TableCell>
                <TableCell className="w-[1%] whitespace-nowrap">
                  <Badge className="bg-interactive-selected border border-border-active text-blue-science hover:bg-interactive-active cursor-default">
                    {ACTIVITY_LABELS[activityType] || activityType}
                  </Badge>
                </TableCell>
                <TableCell className="text-right">{slotInfo.slotsMax}</TableCell>
                <TableCell className="text-right">{slotInfo.slotsInUse}</TableCell>
                <TableCell className="text-right">{slotInfo.slotsReserved}</TableCell>
                <TableCell className="text-right">{slotInfo.slotsListed}</TableCell>
                <TableCell className="text-right">
                  <span className={`font-semibold ${slotInfo.slotsAvailable > 0 ? 'text-teal-success' : 'text-rose-danger'}`}>
                    {slotInfo.slotsAvailable}
                  </span>
                </TableCell>
              </TableRow>
            ));
          })}
        </TableBody>
      </Table>
    </div>
  );
}
