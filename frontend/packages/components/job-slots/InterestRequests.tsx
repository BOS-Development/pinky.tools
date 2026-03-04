import { useState, useEffect } from 'react';
import { useSession } from 'next-auth/react';
import { Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { toast } from '@/components/ui/sonner';
import { formatISK } from '@industry-tool/utils/formatting';

type JobSlotInterestRequest = {
  id: number;
  listingId: number;
  requesterUserId: number;
  requesterName: string;
  slotsRequested: number;
  durationDays: number | null;
  message: string | null;
  status: string;
  createdAt: string;
  updatedAt: string;
  listingActivityType?: string;
  listingCharacterName?: string;
  listingOwnerName?: string;
  listingPriceAmount?: number;
  listingPricingUnit?: string;
};

const ACTIVITY_LABELS: Record<string, string> = {
  manufacturing: 'Manufacturing',
  reaction: 'Reactions',
  copying: 'Copying',
  invention: 'Invention',
  me_research: 'ME Research',
  te_research: 'TE Research',
};

const PRICING_UNIT_LABELS: Record<string, string> = {
  per_slot_day: 'Per Slot/Day',
  per_job: 'Per Job',
  flat_fee: 'Flat Fee',
};

const STATUS_CLASSES: Record<string, string> = {
  pending: 'bg-[rgba(245,158,11,0.1)] border-[rgba(245,158,11,0.3)] text-[#f59e0b]',
  accepted: 'bg-[rgba(16,185,129,0.1)] border-[rgba(16,185,129,0.3)] text-[#10b981]',
  declined: 'bg-[rgba(239,68,68,0.1)] border-[rgba(239,68,68,0.3)] text-[#ef4444]',
  withdrawn: 'bg-[rgba(107,114,128,0.1)] border-[rgba(107,114,128,0.3)] text-[#6b7280]',
};

export default function InterestRequests() {
  const { data: session } = useSession();
  const [tab, setTab] = useState('sent');
  const [sentRequests, setSentRequests] = useState<JobSlotInterestRequest[]>([]);
  const [receivedRequests, setReceivedRequests] = useState<JobSlotInterestRequest[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (session) {
      fetchRequests();
    }
  }, [session]);

  const fetchRequests = async () => {
    setLoading(true);
    try {
      const [sentResponse, receivedResponse] = await Promise.all([
        fetch('/api/job-slots/interest/sent'),
        fetch('/api/job-slots/interest/received'),
      ]);

      if (sentResponse.ok) {
        const sentData = await sentResponse.json();
        setSentRequests(sentData || []);
      }

      if (receivedResponse.ok) {
        const receivedData = await receivedResponse.json();
        setReceivedRequests(receivedData || []);
      }
    } catch (error) {
      console.error('Failed to fetch interest requests:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleStatusUpdate = async (id: number, status: string, actionLabel: string) => {
    if (status !== 'accepted' && !confirm(`Are you sure you want to ${actionLabel.toLowerCase()} this interest request?`)) return;

    try {
      const response = await fetch(`/api/job-slots/interest/${id}/status`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status }),
      });

      if (response.ok) {
        toast.success(`Interest ${actionLabel.toLowerCase()}`);
        fetchRequests();
      } else {
        const error = await response.json();
        toast.error(error.error || `Failed to ${actionLabel.toLowerCase()}`);
      }
    } catch (error) {
      console.error(`${actionLabel} failed:`, error);
      toast.error(`Failed to ${actionLabel.toLowerCase()}`);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-[#00d4ff]" />
      </div>
    );
  }

  return (
    <div>
      <Tabs value={tab} onValueChange={setTab}>
        <TabsList className="mb-3">
          <TabsTrigger value="sent">{`Sent (${sentRequests.length})`}</TabsTrigger>
          <TabsTrigger value="received">{`Received (${receivedRequests.length})`}</TabsTrigger>
        </TabsList>

        <TabsContent value="sent">
          {sentRequests.length === 0 ? (
            <div className="bg-[#12151f] rounded-sm border border-[rgba(148,163,184,0.1)] p-8 text-center">
              <h3 className="text-lg font-semibold text-[#94a3b8]">No sent interest requests</h3>
              <p className="text-sm text-[#64748b] mt-1">
                Browse listings and express interest to get started.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto rounded-sm border border-[rgba(148,163,184,0.1)]">
              <Table>
                <TableHeader>
                  <TableRow className="bg-[#0f1219]">
                    <TableHead>Owner</TableHead>
                    <TableHead>Character</TableHead>
                    <TableHead>Activity</TableHead>
                    <TableHead className="text-right">Slots</TableHead>
                    <TableHead className="text-right">Duration</TableHead>
                    <TableHead className="text-right">Price</TableHead>
                    <TableHead>Message</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-center">Action</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {sentRequests.map((request) => (
                    <TableRow key={request.id} className="hover:bg-[rgba(0,212,255,0.04)]">
                      <TableCell className="text-[#e2e8f0]">{request.listingOwnerName || '-'}</TableCell>
                      <TableCell className="text-[#e2e8f0]">{request.listingCharacterName || '-'}</TableCell>
                      <TableCell>
                        {request.listingActivityType && (
                          <Badge className="bg-[rgba(0,212,255,0.1)] border border-[rgba(0,212,255,0.3)] text-[#60a5fa] hover:bg-[rgba(0,212,255,0.15)] cursor-default">
                            {ACTIVITY_LABELS[request.listingActivityType] || request.listingActivityType}
                          </Badge>
                        )}
                      </TableCell>
                      <TableCell className="text-right text-[#e2e8f0]">{request.slotsRequested}</TableCell>
                      <TableCell className="text-right text-[#94a3b8]">{request.durationDays ? `${request.durationDays} days` : '-'}</TableCell>
                      <TableCell className="text-right text-[#94a3b8]">
                        {request.listingPriceAmount !== undefined && request.listingPricingUnit
                          ? `${formatISK(request.listingPriceAmount)} ${PRICING_UNIT_LABELS[request.listingPricingUnit] || request.listingPricingUnit}`
                          : '-'}
                      </TableCell>
                      <TableCell>
                        {request.message ? (
                          <span className="text-xs text-[#94a3b8]">{request.message}</span>
                        ) : '-'}
                      </TableCell>
                      <TableCell>
                        <Badge className={`border capitalize cursor-default ${STATUS_CLASSES[request.status] || STATUS_CLASSES.withdrawn}`}>
                          {request.status}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-center">
                        {request.status === 'pending' && (
                          <Button
                            variant="outline"
                            size="sm"
                            className="text-[#ef4444] border-[#ef4444] hover:bg-[rgba(239,68,68,0.1)]"
                            onClick={() => handleStatusUpdate(request.id, 'withdrawn', 'Withdraw')}
                          >
                            Withdraw
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </TabsContent>

        <TabsContent value="received">
          {receivedRequests.length === 0 ? (
            <div className="bg-[#12151f] rounded-sm border border-[rgba(148,163,184,0.1)] p-8 text-center">
              <h3 className="text-lg font-semibold text-[#94a3b8]">No received interest requests</h3>
              <p className="text-sm text-[#64748b] mt-1">
                Create listings to receive interest requests.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto rounded-sm border border-[rgba(148,163,184,0.1)]">
              <Table>
                <TableHeader>
                  <TableRow className="bg-[#0f1219]">
                    <TableHead>Requester</TableHead>
                    <TableHead className="text-right">Slots</TableHead>
                    <TableHead className="text-right">Duration</TableHead>
                    <TableHead>Message</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Requested At</TableHead>
                    <TableHead className="text-center">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {receivedRequests.map((request) => (
                    <TableRow key={request.id} className="hover:bg-[rgba(0,212,255,0.04)]">
                      <TableCell>
                        <Badge className="bg-[rgba(0,212,255,0.1)] border border-[rgba(0,212,255,0.3)] text-[#60a5fa] hover:bg-[rgba(0,212,255,0.15)] cursor-default">
                          {request.requesterName}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right text-[#e2e8f0]">{request.slotsRequested}</TableCell>
                      <TableCell className="text-right text-[#94a3b8]">{request.durationDays ? `${request.durationDays} days` : '-'}</TableCell>
                      <TableCell>
                        {request.message ? (
                          <span className="text-xs text-[#94a3b8]">{request.message}</span>
                        ) : '-'}
                      </TableCell>
                      <TableCell>
                        <Badge className={`border capitalize cursor-default ${STATUS_CLASSES[request.status] || STATUS_CLASSES.withdrawn}`}>
                          {request.status}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-[#94a3b8] text-sm">{new Date(request.createdAt).toLocaleDateString()}</TableCell>
                      <TableCell className="text-center">
                        {request.status === 'pending' && (
                          <div className="flex gap-1 justify-center">
                            <Button
                              size="sm"
                              className="bg-[#10b981] hover:bg-[#059669]"
                              onClick={() => handleStatusUpdate(request.id, 'accepted', 'Accept')}
                            >
                              Accept
                            </Button>
                            <Button
                              variant="outline"
                              size="sm"
                              className="text-[#ef4444] border-[#ef4444] hover:bg-[rgba(239,68,68,0.1)]"
                              onClick={() => handleStatusUpdate(request.id, 'declined', 'Decline')}
                            >
                              Decline
                            </Button>
                          </div>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}
