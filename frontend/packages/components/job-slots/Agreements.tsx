import { useState, useEffect } from 'react';
import { useSession } from 'next-auth/react';
import { Loader2, ChevronDown, ChevronUp } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { toast } from '@/components/ui/sonner';
import { formatISK } from '@industry-tool/utils/formatting';

type Agreement = {
  id: number;
  listingId: number;
  sellerUserId: number;
  sellerName: string;
  renterUserId: number;
  renterName: string;
  activityType: string;
  characterName: string;
  slotsAgreed: number;
  priceAmount: number;
  pricingUnit: string;
  status: string;
  agreedAt: string;
  expectedEndAt: string | null;
  cancellationReason: string | null;
};

type AgreementJob = {
  jobId: number;
  activityType: string;
  blueprintTypeName: string;
  productTypeName: string | null;
  runs: number;
  startDate: string;
  endDate: string;
  status: string;
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
  active: 'bg-interactive-selected border border-border-active text-blue-science',
  completed: 'bg-teal-success/10 border-teal-success/30 text-teal-success',
  cancelled: 'bg-overlay-subtle border-overlay-strong text-text-muted',
};

const JOB_STATUS_CLASSES: Record<string, string> = {
  active: 'bg-interactive-selected border border-border-active text-blue-science',
  delivered: 'bg-teal-success/10 border-teal-success/30 text-teal-success',
  cancelled: 'bg-overlay-subtle border-overlay-strong text-text-muted',
  paused: 'bg-amber-manufacturing/10 border-amber-manufacturing/30 text-amber-manufacturing',
};

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

type JobsPanelProps = {
  agreementId: number;
};

function JobsPanel({ agreementId }: JobsPanelProps) {
  const [jobs, setJobs] = useState<AgreementJob[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchJobs = async () => {
      setLoading(true);
      try {
        const response = await fetch(`/api/job-slots/agreements/${agreementId}/jobs`);
        if (response.ok) {
          const data = await response.json();
          setJobs(data || []);
        }
      } catch (err) {
        console.error('Failed to fetch agreement jobs:', err);
      } finally {
        setLoading(false);
      }
    };
    fetchJobs();
  }, [agreementId]);

  if (loading) {
    return (
      <div className="flex justify-center items-center py-6">
        <Loader2 className="h-5 w-5 animate-spin text-primary" />
      </div>
    );
  }

  if (jobs.length === 0) {
    return (
      <div className="py-4 text-center text-sm text-text-muted">
        No active jobs found for this character.
      </div>
    );
  }

  return (
    <div className="overflow-x-auto rounded-sm border border-overlay-subtle mt-2">
      <Table>
        <TableHeader>
          <TableRow className="bg-background-void">
            <TableHead>Item</TableHead>
            <TableHead>Activity</TableHead>
            <TableHead className="text-right">Runs</TableHead>
            <TableHead>Start</TableHead>
            <TableHead>End</TableHead>
            <TableHead>Status</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {jobs.map((job) => (
            <TableRow key={job.jobId} className="hover:bg-interactive-hover">
              <TableCell className="text-text-emphasis">
                {job.productTypeName || job.blueprintTypeName}
              </TableCell>
              <TableCell>
                <Badge className="bg-interactive-selected border border-border-active text-blue-science hover:bg-interactive-active cursor-default">
                  {ACTIVITY_LABELS[job.activityType] || job.activityType}
                </Badge>
              </TableCell>
              <TableCell className="text-right text-text-emphasis">{job.runs}</TableCell>
              <TableCell className="text-text-secondary text-sm">{formatDate(job.startDate)}</TableCell>
              <TableCell className="text-text-secondary text-sm">{formatDate(job.endDate)}</TableCell>
              <TableCell>
                <Badge className={`border capitalize cursor-default ${JOB_STATUS_CLASSES[job.status] || JOB_STATUS_CLASSES.cancelled}`}>
                  {job.status}
                </Badge>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

type CancelDialogProps = {
  open: boolean;
  onClose: () => void;
  onConfirm: (reason: string) => void;
};

function CancelDialog({ open, onClose, onConfirm }: CancelDialogProps) {
  const [reason, setReason] = useState('');

  const handleConfirm = () => {
    onConfirm(reason);
    setReason('');
  };

  const handleClose = () => {
    setReason('');
    onClose();
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="bg-background-panel border-overlay-subtle">
        <DialogHeader>
          <DialogTitle className="text-text-emphasis">Cancel Agreement</DialogTitle>
        </DialogHeader>
        <div className="space-y-3 py-2">
          <Label htmlFor="cancel-reason" className="text-text-secondary">
            Cancellation Reason (optional)
          </Label>
          <textarea
            id="cancel-reason"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="Provide a reason for cancelling this agreement..."
            rows={3}
            className="flex w-full rounded-sm border border-overlay-subtle bg-background-void px-3 py-2 text-sm text-text-primary placeholder:text-text-muted focus-visible:outline-none focus-visible:border-blue-science resize-none"
          />
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={handleClose} className="border-overlay-strong text-text-secondary">
            Back
          </Button>
          <Button
            className="bg-rose-danger hover:bg-rose-danger/80 text-white"
            onClick={handleConfirm}
          >
            Cancel Agreement
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default function Agreements() {
  const { data: session } = useSession();
  const [tab, setTab] = useState('seller');
  const [sellerAgreements, setSellerAgreements] = useState<Agreement[]>([]);
  const [renterAgreements, setRenterAgreements] = useState<Agreement[]>([]);
  const [loading, setLoading] = useState(true);
  const [expandedJobsId, setExpandedJobsId] = useState<number | null>(null);
  const [cancelDialogId, setCancelDialogId] = useState<number | null>(null);

  useEffect(() => {
    if (session) {
      fetchAgreements();
    }
  }, [session]);

  const fetchAgreements = async () => {
    setLoading(true);
    try {
      const [sellerResponse, renterResponse] = await Promise.all([
        fetch('/api/job-slots/agreements?role=seller'),
        fetch('/api/job-slots/agreements?role=renter'),
      ]);

      if (sellerResponse.ok) {
        const data = await sellerResponse.json();
        setSellerAgreements(data || []);
      }

      if (renterResponse.ok) {
        const data = await renterResponse.json();
        setRenterAgreements(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch agreements:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleStatusUpdate = async (id: number, status: string, cancellationReason?: string) => {
    try {
      const body: { status: string; cancellationReason?: string } = { status };
      if (cancellationReason) body.cancellationReason = cancellationReason;

      const response = await fetch(`/api/job-slots/agreements/${id}/status`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });

      if (response.ok) {
        const label = status === 'completed' ? 'completed' : 'cancelled';
        toast.success(`Agreement ${label}`);
        fetchAgreements();
      } else {
        const error = await response.json();
        toast.error(error.error || `Failed to update agreement`);
      }
    } catch (err) {
      console.error('Agreement status update failed:', err);
      toast.error('Failed to update agreement');
    }
  };

  const handleMarkComplete = async (id: number) => {
    if (!confirm('Mark this agreement as complete?')) return;
    await handleStatusUpdate(id, 'completed');
  };

  const handleCancelConfirm = async (reason: string) => {
    if (cancelDialogId === null) return;
    const id = cancelDialogId;
    setCancelDialogId(null);
    await handleStatusUpdate(id, 'cancelled', reason);
  };

  const toggleJobs = (id: number) => {
    setExpandedJobsId((prev) => (prev === id ? null : id));
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
          <TabsTrigger value="seller">{`As Seller (${sellerAgreements.length})`}</TabsTrigger>
          <TabsTrigger value="renter">{`As Renter (${renterAgreements.length})`}</TabsTrigger>
        </TabsList>

        <TabsContent value="seller">
          {sellerAgreements.length === 0 ? (
            <div className="bg-background-panel rounded-sm border border-overlay-subtle p-8 text-center">
              <h3 className="text-lg font-semibold text-text-secondary">No seller agreements</h3>
              <p className="text-sm text-text-muted mt-1">
                Accept interest requests on your listings to create agreements.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
              <Table>
                <TableHeader>
                  <TableRow className="bg-background-void">
                    <TableHead>Renter</TableHead>
                    <TableHead>Activity</TableHead>
                    <TableHead>Character</TableHead>
                    <TableHead className="text-right">Slots</TableHead>
                    <TableHead className="text-right">Price</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Agreed</TableHead>
                    <TableHead>Expected End</TableHead>
                    <TableHead className="text-center">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {sellerAgreements.map((agreement) => (
                    <>
                      <TableRow key={agreement.id} className="hover:bg-interactive-hover">
                        <TableCell className="text-text-emphasis">{agreement.renterName}</TableCell>
                        <TableCell>
                          <Badge className="bg-interactive-selected border border-border-active text-blue-science hover:bg-interactive-active cursor-default">
                            {ACTIVITY_LABELS[agreement.activityType] || agreement.activityType}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-text-secondary">{agreement.characterName}</TableCell>
                        <TableCell className="text-right text-text-emphasis">{agreement.slotsAgreed}</TableCell>
                        <TableCell className="text-right text-text-secondary">
                          {`${formatISK(agreement.priceAmount)} ${PRICING_UNIT_LABELS[agreement.pricingUnit] || agreement.pricingUnit}`}
                        </TableCell>
                        <TableCell>
                          <Badge className={`border capitalize cursor-default ${STATUS_CLASSES[agreement.status] || STATUS_CLASSES.cancelled}`}>
                            {agreement.status}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-text-secondary text-sm">{formatDate(agreement.agreedAt)}</TableCell>
                        <TableCell className="text-text-secondary text-sm">
                          {agreement.expectedEndAt ? formatDate(agreement.expectedEndAt) : '-'}
                        </TableCell>
                        <TableCell className="text-center">
                          <div className="flex gap-1 justify-center flex-wrap">
                            {agreement.status === 'active' && (
                              <>
                                <Button
                                  size="sm"
                                  variant="outline"
                                  className="text-text-secondary border-overlay-strong"
                                  onClick={() => toggleJobs(agreement.id)}
                                >
                                  {expandedJobsId === agreement.id ? (
                                    <><ChevronUp className="h-3 w-3 mr-1" />Jobs</>
                                  ) : (
                                    <><ChevronDown className="h-3 w-3 mr-1" />Jobs</>
                                  )}
                                </Button>
                                <Button
                                  size="sm"
                                  className="bg-teal-success hover:bg-teal-success/80"
                                  onClick={() => handleMarkComplete(agreement.id)}
                                >
                                  Complete
                                </Button>
                                <Button
                                  variant="outline"
                                  size="sm"
                                  className="text-rose-danger border-rose-danger hover:bg-rose-danger/10"
                                  onClick={() => setCancelDialogId(agreement.id)}
                                >
                                  Cancel
                                </Button>
                              </>
                            )}
                          </div>
                        </TableCell>
                      </TableRow>
                      {expandedJobsId === agreement.id && (
                        <TableRow key={`${agreement.id}-jobs`}>
                          <TableCell colSpan={9} className="bg-background-void p-0">
                            <div className="px-4 py-3">
                              <JobsPanel agreementId={agreement.id} />
                            </div>
                          </TableCell>
                        </TableRow>
                      )}
                    </>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </TabsContent>

        <TabsContent value="renter">
          {renterAgreements.length === 0 ? (
            <div className="bg-background-panel rounded-sm border border-overlay-subtle p-8 text-center">
              <h3 className="text-lg font-semibold text-text-secondary">No renter agreements</h3>
              <p className="text-sm text-text-muted mt-1">
                Send interest requests on listings and wait for a seller to accept.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
              <Table>
                <TableHeader>
                  <TableRow className="bg-background-void">
                    <TableHead>Seller</TableHead>
                    <TableHead>Activity</TableHead>
                    <TableHead>Character</TableHead>
                    <TableHead className="text-right">Slots</TableHead>
                    <TableHead className="text-right">Price</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Agreed</TableHead>
                    <TableHead>Expected End</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {renterAgreements.map((agreement) => (
                    <TableRow key={agreement.id} className="hover:bg-interactive-hover">
                      <TableCell className="text-text-emphasis">{agreement.sellerName}</TableCell>
                      <TableCell>
                        <Badge className="bg-interactive-selected border border-border-active text-blue-science hover:bg-interactive-active cursor-default">
                          {ACTIVITY_LABELS[agreement.activityType] || agreement.activityType}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-text-secondary">{agreement.characterName}</TableCell>
                      <TableCell className="text-right text-text-emphasis">{agreement.slotsAgreed}</TableCell>
                      <TableCell className="text-right text-text-secondary">
                        {`${formatISK(agreement.priceAmount)} ${PRICING_UNIT_LABELS[agreement.pricingUnit] || agreement.pricingUnit}`}
                      </TableCell>
                      <TableCell>
                        <Badge className={`border capitalize cursor-default ${STATUS_CLASSES[agreement.status] || STATUS_CLASSES.cancelled}`}>
                          {agreement.status}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-text-secondary text-sm">{formatDate(agreement.agreedAt)}</TableCell>
                      <TableCell className="text-text-secondary text-sm">
                        {agreement.expectedEndAt ? formatDate(agreement.expectedEndAt) : '-'}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </TabsContent>
      </Tabs>

      <CancelDialog
        open={cancelDialogId !== null}
        onClose={() => setCancelDialogId(null)}
        onConfirm={handleCancelConfirm}
      />
    </div>
  );
}
