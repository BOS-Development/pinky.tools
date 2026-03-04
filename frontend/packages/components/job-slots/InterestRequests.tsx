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
  pending: 'bg-amber-manufacturing/10 border-amber-manufacturing/30 text-amber-manufacturing',
  accepted: 'bg-teal-success/10 border-teal-success/30 text-teal-success',
  declined: 'bg-rose-danger/10 border-rose-danger/30 text-rose-danger',
  withdrawn: 'bg-overlay-subtle border-overlay-strong text-text-muted',
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
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
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
            <div className="bg-background-panel rounded-sm border border-overlay-subtle p-8 text-center">
              <h3 className="text-lg font-semibold text-text-secondary">No sent interest requests</h3>
              <p className="text-sm text-text-muted mt-1">
                Browse listings and express interest to get started.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
              <Table>
                <TableHeader>
                  <TableRow className="bg-background-void">
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
                    <TableRow key={request.id} className="hover:bg-interactive-hover">
                      <TableCell className="text-text-emphasis">{request.listingOwnerName || '-'}</TableCell>
                      <TableCell className="text-text-emphasis">{request.listingCharacterName || '-'}</TableCell>
                      <TableCell>
                        {request.listingActivityType && (
                          <Badge className="bg-interactive-selected border border-border-active text-blue-science hover:bg-interactive-active cursor-default">
                            {ACTIVITY_LABELS[request.listingActivityType] || request.listingActivityType}
                          </Badge>
                        )}
                      </TableCell>
                      <TableCell className="text-right text-text-emphasis">{request.slotsRequested}</TableCell>
                      <TableCell className="text-right text-text-secondary">{request.durationDays ? `${request.durationDays} days` : '-'}</TableCell>
                      <TableCell className="text-right text-text-secondary">
                        {request.listingPriceAmount !== undefined && request.listingPricingUnit
                          ? `${formatISK(request.listingPriceAmount)} ${PRICING_UNIT_LABELS[request.listingPricingUnit] || request.listingPricingUnit}`
                          : '-'}
                      </TableCell>
                      <TableCell>
                        {request.message ? (
                          <span className="text-xs text-text-secondary">{request.message}</span>
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
                            className="text-rose-danger border-rose-danger hover:bg-rose-danger/10"
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
            <div className="bg-background-panel rounded-sm border border-overlay-subtle p-8 text-center">
              <h3 className="text-lg font-semibold text-text-secondary">No received interest requests</h3>
              <p className="text-sm text-text-muted mt-1">
                Create listings to receive interest requests.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
              <Table>
                <TableHeader>
                  <TableRow className="bg-background-void">
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
                    <TableRow key={request.id} className="hover:bg-interactive-hover">
                      <TableCell>
                        <Badge className="bg-interactive-selected border border-border-active text-blue-science hover:bg-interactive-active cursor-default">
                          {request.requesterName}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right text-text-emphasis">{request.slotsRequested}</TableCell>
                      <TableCell className="text-right text-text-secondary">{request.durationDays ? `${request.durationDays} days` : '-'}</TableCell>
                      <TableCell>
                        {request.message ? (
                          <span className="text-xs text-text-secondary">{request.message}</span>
                        ) : '-'}
                      </TableCell>
                      <TableCell>
                        <Badge className={`border capitalize cursor-default ${STATUS_CLASSES[request.status] || STATUS_CLASSES.withdrawn}`}>
                          {request.status}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-text-secondary text-sm">{new Date(request.createdAt).toLocaleDateString()}</TableCell>
                      <TableCell className="text-center">
                        {request.status === 'pending' && (
                          <div className="flex gap-1 justify-center">
                            <Button
                              size="sm"
                              className="bg-teal-success hover:bg-teal-success/80"
                              onClick={() => handleStatusUpdate(request.id, 'accepted', 'Accept')}
                            >
                              Accept
                            </Button>
                            <Button
                              variant="outline"
                              size="sm"
                              className="text-rose-danger border-rose-danger hover:bg-rose-danger/10"
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
