import { useState, useEffect, useRef } from 'react';
import { useSession } from "next-auth/react";
import Navbar from "@industry-tool/components/Navbar";
import Loading from "@industry-tool/components/loading";
import { Check, X, Trash2, Settings, UserPlus, Scale, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import {
  Table, TableHeader, TableBody, TableRow, TableHead, TableCell,
} from '@/components/ui/table';
import {
  Select, SelectTrigger, SelectValue, SelectContent, SelectItem,
} from '@/components/ui/select';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription,
} from '@/components/ui/dialog';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { toast } from '@/components/ui/sonner';
import PermissionsDialog from './PermissionsDialog';

export type Contact = {
  id: number;
  requesterUserId: number;
  recipientUserId: number;
  requesterName: string;
  recipientName: string;
  status: 'pending' | 'accepted' | 'rejected';
  requestedAt: string;
  respondedAt?: string;
  contactRuleId?: number | null;
};

type ContactRule = {
  id: number;
  userId: number;
  ruleType: 'corporation' | 'alliance' | 'everyone';
  entityId: number | null;
  entityName: string | null;
  permissions: string[];
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

const SERVICE_TYPES = [
  { type: 'for_sale_browse', label: 'Browse For-Sale Items' },
];

type SearchResult = {
  id: number;
  name: string;
};

export default function ContactsList() {
  const { data: session } = useSession();
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [contactRules, setContactRules] = useState<ContactRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [addContactOpen, setAddContactOpen] = useState(false);
  const [newContactCharacterName, setNewContactCharacterName] = useState('');
  const [permissionsDialogOpen, setPermissionsDialogOpen] = useState(false);
  const [selectedContact, setSelectedContact] = useState<Contact | null>(null);
  const hasFetchedRef = useRef(false);

  // Add Rule dialog state
  const [addRuleOpen, setAddRuleOpen] = useState(false);
  const [newRuleType, setNewRuleType] = useState<string>('corporation');
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<SearchResult[]>([]);
  const [selectedEntity, setSelectedEntity] = useState<SearchResult | null>(null);
  const [searchLoading, setSearchLoading] = useState(false);
  const searchTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [rulePermissions, setRulePermissions] = useState<string[]>(SERVICE_TYPES.map(s => s.type));

  const currentUserId = session?.providerAccountId ? parseInt(session.providerAccountId) : null;

  useEffect(() => {
    if (session && !hasFetchedRef.current) {
      hasFetchedRef.current = true;
      fetchContacts();
      fetchContactRules();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [session]);

  const fetchContacts = async () => {
    if (!session) return;

    setLoading(true);
    try {
      const response = await fetch('/api/contacts');
      if (response.ok) {
        const data: Contact[] = await response.json();
        setContacts(data || []);
      }
    } finally {
      setLoading(false);
    }
  };

  const fetchContactRules = async () => {
    if (!session) return;

    try {
      const response = await fetch('/api/contact-rules');
      if (response.ok) {
        const data: ContactRule[] = await response.json();
        setContactRules(data || []);
      }
    } catch {
      // Silently fail
    }
  };

  const handleAddContact = async () => {
    if (!newContactCharacterName || !session) return;

    try {
      const response = await fetch('/api/contacts', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ characterName: newContactCharacterName }),
      });

      if (response.ok) {
        toast.success('Contact request sent!');
        setAddContactOpen(false);
        setNewContactCharacterName('');
        await fetchContacts();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Failed to send contact request');
      }
    } catch {
      toast.error('Failed to send contact request');
    }
  };

  const handleAccept = async (contactId: number) => {
    try {
      const response = await fetch(`/api/contacts/${contactId}/accept`, { method: 'POST' });
      if (response.ok) {
        toast.success('Contact accepted!');
        await fetchContacts();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Failed to accept contact');
      }
    } catch {
      toast.error('Failed to accept contact');
    }
  };

  const handleReject = async (contactId: number) => {
    try {
      const response = await fetch(`/api/contacts/${contactId}/reject`, { method: 'POST' });
      if (response.ok) {
        toast.success('Contact rejected');
        await fetchContacts();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Failed to reject contact');
      }
    } catch {
      toast.error('Failed to reject contact');
    }
  };

  const handleDelete = async (contactId: number) => {
    try {
      const response = await fetch(`/api/contacts/${contactId}`, { method: 'DELETE' });
      if (response.ok) {
        toast.success('Contact removed');
        await fetchContacts();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Failed to remove contact');
      }
    } catch {
      toast.error('Failed to remove contact');
    }
  };

  const handleOpenPermissions = (contact: Contact) => {
    setSelectedContact(contact);
    setPermissionsDialogOpen(true);
  };

  const handleClosePermissions = () => {
    setPermissionsDialogOpen(false);
    setSelectedContact(null);
  };

  const handleSearchEntities = async (query: string) => {
    if (!query || query.length < 2) {
      setSearchResults([]);
      return;
    }

    setSearchLoading(true);
    try {
      const endpoint = newRuleType === 'corporation'
        ? `/api/contact-rules/corporations?q=${encodeURIComponent(query)}`
        : `/api/contact-rules/alliances?q=${encodeURIComponent(query)}`;

      const response = await fetch(endpoint);
      if (response.ok) {
        const data: SearchResult[] = await response.json();
        setSearchResults(data || []);
      }
    } catch {
      // Silently fail
    } finally {
      setSearchLoading(false);
    }
  };

  const handleSearchInputChange = (value: string) => {
    setSearchQuery(value);

    if (searchTimeoutRef.current) {
      clearTimeout(searchTimeoutRef.current);
    }

    searchTimeoutRef.current = setTimeout(() => {
      handleSearchEntities(value);
    }, 300);
  };

  const handleCreateRule = async () => {
    if (!session) return;

    if (newRuleType !== 'everyone' && !selectedEntity) {
      toast.error('Please select a corporation or alliance');
      return;
    }

    try {
      const body: { ruleType: string; entityId?: number; entityName?: string; permissions: string[] } = {
        ruleType: newRuleType,
        permissions: rulePermissions,
      };

      if (newRuleType !== 'everyone' && selectedEntity) {
        body.entityId = selectedEntity.id;
        body.entityName = selectedEntity.name;
      }

      const response = await fetch('/api/contact-rules', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });

      if (response.ok) {
        toast.success('Contact rule created! Auto-contacts are being generated.');
        setAddRuleOpen(false);
        setNewRuleType('corporation');
        setSelectedEntity(null);
        setSearchQuery('');
        setSearchResults([]);
        setRulePermissions(SERVICE_TYPES.map(s => s.type));
        await fetchContactRules();
        setTimeout(() => fetchContacts(), 2000);
      } else {
        const error = await response.json();
        toast.error(error.error || 'Failed to create contact rule');
      }
    } catch {
      toast.error('Failed to create contact rule');
    }
  };

  const handleDeleteRule = async (ruleId: number) => {
    try {
      const response = await fetch(`/api/contact-rules/${ruleId}`, { method: 'DELETE' });
      if (response.ok) {
        toast.success('Contact rule removed');
        await fetchContactRules();
        await fetchContacts();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Failed to remove contact rule');
      }
    } catch {
      toast.error('Failed to remove contact rule');
    }
  };

  const getRuleTypeVariant = (ruleType: string): 'default' | 'secondary' | 'success' => {
    switch (ruleType) {
      case 'corporation': return 'default';
      case 'alliance': return 'secondary';
      case 'everyone': return 'success';
      default: return 'default';
    }
  };

  const getRuleTypeLabel = (ruleType: string): string => {
    switch (ruleType) {
      case 'corporation': return 'Corporation';
      case 'alliance': return 'Alliance';
      case 'everyone': return 'Everyone';
      default: return ruleType;
    }
  };

  const myContacts = contacts.filter(c => c.status === 'accepted');
  const pendingRequests = contacts.filter(c =>
    c.status === 'pending' && c.recipientUserId === currentUserId
  );
  const sentRequests = contacts.filter(c =>
    c.status === 'pending' && c.requesterUserId === currentUserId
  );

  if (!session) return null;
  if (loading) return <Loading />;

  return (
    <>
      <Navbar />
      <div className="max-w-5xl mx-auto px-4 py-8">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-display font-semibold">Contacts</h1>
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => setAddRuleOpen(true)}>
              <Scale className="h-4 w-4 mr-2" />
              Add Rule
            </Button>
            <Button onClick={() => setAddContactOpen(true)}>
              <UserPlus className="h-4 w-4 mr-2" />
              Add Contact
            </Button>
          </div>
        </div>

        <Card>
          <Tabs defaultValue="contacts">
            <TabsList className="w-full justify-start border-b border-[var(--color-border-dim)] rounded-none bg-transparent p-0">
              <TabsTrigger value="contacts" className="rounded-none border-b-2 border-transparent data-[state=active]:border-[var(--color-primary-cyan)] data-[state=active]:shadow-none">
                My Contacts ({myContacts.length})
              </TabsTrigger>
              <TabsTrigger value="pending" className="rounded-none border-b-2 border-transparent data-[state=active]:border-[var(--color-primary-cyan)] data-[state=active]:shadow-none">
                Pending ({pendingRequests.length})
              </TabsTrigger>
              <TabsTrigger value="sent" className="rounded-none border-b-2 border-transparent data-[state=active]:border-[var(--color-primary-cyan)] data-[state=active]:shadow-none">
                Sent ({sentRequests.length})
              </TabsTrigger>
              <TabsTrigger value="rules" className="rounded-none border-b-2 border-transparent data-[state=active]:border-[var(--color-primary-cyan)] data-[state=active]:shadow-none">
                Rules ({contactRules.length})
              </TabsTrigger>
            </TabsList>

            <TabsContent value="contacts">
              <CardContent className="p-0">
                {myContacts.length === 0 ? (
                  <div className="p-8 text-center">
                    <p className="text-[var(--color-text-secondary)]">No contacts yet. Add a contact to get started!</p>
                  </div>
                ) : (
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Character Name</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Connected Since</TableHead>
                        <TableHead className="text-right">Actions</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {myContacts.map((contact) => {
                        const otherUserName = contact.requesterUserId === currentUserId
                          ? contact.recipientName
                          : contact.requesterName;
                        return (
                          <TableRow key={contact.id}>
                            <TableCell>
                              <span className="text-[var(--color-text-primary)]">{otherUserName}</span>
                              {contact.contactRuleId && <Badge variant="info" className="ml-2 text-[10px]">Auto</Badge>}
                            </TableCell>
                            <TableCell><Badge variant="success">Connected</Badge></TableCell>
                            <TableCell className="text-[var(--color-text-secondary)]">
                              {new Date(contact.respondedAt || contact.requestedAt).toLocaleDateString()}
                            </TableCell>
                            <TableCell className="text-right">
                              <Button variant="ghost" size="icon" onClick={() => handleOpenPermissions(contact)} title="Manage Permissions"><Settings className="h-4 w-4" /></Button>
                              <Button variant="ghost" size="icon" onClick={() => handleDelete(contact.id)} title="Remove Contact" className="text-[var(--color-danger-rose)] hover:text-[var(--color-danger-rose)]"><Trash2 className="h-4 w-4" /></Button>
                            </TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                )}
              </CardContent>
            </TabsContent>

            <TabsContent value="pending">
              <CardContent className="p-0">
                {pendingRequests.length === 0 ? (
                  <div className="p-8 text-center"><p className="text-[var(--color-text-secondary)]">No pending requests</p></div>
                ) : (
                  <Table>
                    <TableHeader><TableRow><TableHead>Character Name</TableHead><TableHead>Requested</TableHead><TableHead className="text-right">Actions</TableHead></TableRow></TableHeader>
                    <TableBody>
                      {pendingRequests.map((contact) => (
                        <TableRow key={contact.id}>
                          <TableCell className="text-[var(--color-text-primary)]">{contact.requesterName}</TableCell>
                          <TableCell className="text-[var(--color-text-secondary)]">{new Date(contact.requestedAt).toLocaleDateString()}</TableCell>
                          <TableCell className="text-right">
                            <Button variant="ghost" size="icon" onClick={() => handleAccept(contact.id)} title="Accept" className="text-[var(--color-success-teal)] hover:text-[var(--color-success-teal)]"><Check className="h-4 w-4" /></Button>
                            <Button variant="ghost" size="icon" onClick={() => handleReject(contact.id)} title="Reject" className="text-[var(--color-danger-rose)] hover:text-[var(--color-danger-rose)]"><X className="h-4 w-4" /></Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                )}
              </CardContent>
            </TabsContent>

            <TabsContent value="sent">
              <CardContent className="p-0">
                {sentRequests.length === 0 ? (
                  <div className="p-8 text-center"><p className="text-[var(--color-text-secondary)]">No sent requests</p></div>
                ) : (
                  <Table>
                    <TableHeader><TableRow><TableHead>Character Name</TableHead><TableHead>Sent</TableHead><TableHead className="text-right">Actions</TableHead></TableRow></TableHeader>
                    <TableBody>
                      {sentRequests.map((contact) => (
                        <TableRow key={contact.id}>
                          <TableCell className="text-[var(--color-text-primary)]">{contact.recipientName}</TableCell>
                          <TableCell className="text-[var(--color-text-secondary)]">{new Date(contact.requestedAt).toLocaleDateString()}</TableCell>
                          <TableCell className="text-right">
                            <Button variant="ghost" size="icon" onClick={() => handleDelete(contact.id)} title="Cancel Request" className="text-[var(--color-danger-rose)] hover:text-[var(--color-danger-rose)]"><X className="h-4 w-4" /></Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                )}
              </CardContent>
            </TabsContent>

            <TabsContent value="rules">
              <CardContent className="p-0">
                {contactRules.length === 0 ? (
                  <div className="p-8 text-center"><p className="text-[var(--color-text-secondary)]">No contact rules yet. Add a rule to automatically connect with corporations, alliances, or everyone.</p></div>
                ) : (
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Type</TableHead><TableHead>Entity</TableHead><TableHead>Permissions</TableHead><TableHead>Created</TableHead><TableHead className="text-right">Actions</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {contactRules.map((rule) => (
                        <TableRow key={rule.id}>
                          <TableCell><Badge variant={getRuleTypeVariant(rule.ruleType)}>{getRuleTypeLabel(rule.ruleType)}</Badge></TableCell>
                          <TableCell className="text-[var(--color-text-primary)]">{rule.ruleType === 'everyone' ? 'All Users' : (rule.entityName || `ID: ${rule.entityId}`)}</TableCell>
                          <TableCell>
                            <div className="flex gap-1 flex-wrap">
                              {(rule.permissions || []).map((perm) => {
                                const label = SERVICE_TYPES.find(s => s.type === perm)?.label || perm;
                                return <Badge key={perm} variant="outline" className="text-[10px]">{label}</Badge>;
                              })}
                            </div>
                          </TableCell>
                          <TableCell className="text-[var(--color-text-secondary)]">{new Date(rule.createdAt).toLocaleDateString()}</TableCell>
                          <TableCell className="text-right">
                            <Button variant="ghost" size="icon" onClick={() => handleDeleteRule(rule.id)} title="Remove Rule" className="text-[var(--color-danger-rose)] hover:text-[var(--color-danger-rose)]"><Trash2 className="h-4 w-4" /></Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                )}
              </CardContent>
            </TabsContent>
          </Tabs>
        </Card>
      </div>

      {/* Add Contact Dialog */}
      <Dialog open={addContactOpen} onOpenChange={setAddContactOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Add Contact</DialogTitle>
            <DialogDescription>Send a contact request to another player.</DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            <div>
              <Label htmlFor="charName">Character Name</Label>
              <Input id="charName" value={newContactCharacterName} onChange={(e) => setNewContactCharacterName(e.target.value)} placeholder="Enter character name..." autoFocus />
              <p className="text-xs text-[var(--color-text-muted)] mt-1">Enter the character name of the person you want to add</p>
            </div>
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setAddContactOpen(false)}>Cancel</Button>
            <Button onClick={handleAddContact}>Send Request</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Add Rule Dialog */}
      <Dialog open={addRuleOpen} onOpenChange={setAddRuleOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Add Contact Rule</DialogTitle>
            <DialogDescription>
              Contact rules automatically create connections with all members of a corporation, alliance, or everyone.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Rule Type</Label>
              <Select value={newRuleType} onValueChange={(value) => { setNewRuleType(value); setSelectedEntity(null); setSearchQuery(''); setSearchResults([]); }}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="corporation">Corporation</SelectItem>
                  <SelectItem value="alliance">Alliance</SelectItem>
                  <SelectItem value="everyone">Everyone</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {newRuleType !== 'everyone' && (
              <div>
                <Label>{newRuleType === 'corporation' ? 'Search Corporations' : 'Search Alliances'}</Label>
                <Input value={searchQuery} onChange={(e) => handleSearchInputChange(e.target.value)} placeholder={`Search for a ${newRuleType} by name...`} />
                {searchLoading && (
                  <div className="flex items-center gap-2 mt-2">
                    <Loader2 className="h-4 w-4 animate-spin text-[var(--color-primary-cyan)]" />
                    <span className="text-xs text-[var(--color-text-muted)]">Searching...</span>
                  </div>
                )}
                {searchResults.length > 0 && (
                  <div className="mt-2 border border-[var(--color-border-dim)] rounded-sm max-h-40 overflow-y-auto">
                    {searchResults.map(result => (
                      <button key={result.id} onClick={() => { setSelectedEntity(result); setSearchQuery(result.name); setSearchResults([]); }}
                        className="w-full text-left px-3 py-2 text-sm hover:bg-[var(--color-surface-elevated)] transition-colors cursor-pointer">{result.name}</button>
                    ))}
                  </div>
                )}
                {selectedEntity && <p className="text-xs text-[var(--color-primary-cyan)] mt-1">Selected: {selectedEntity.name}</p>}
              </div>
            )}

            {newRuleType === 'everyone' && (
              <Alert><AlertDescription>This will automatically connect you with every user in the system.</AlertDescription></Alert>
            )}

            <div>
              <Label className="mb-2 block">Permissions to Grant</Label>
              {SERVICE_TYPES.map((service) => (
                <div key={service.type} className="flex items-center gap-2 py-1">
                  <Checkbox id={`perm-${service.type}`} checked={rulePermissions.includes(service.type)}
                    onCheckedChange={(checked) => {
                      if (checked) { setRulePermissions([...rulePermissions, service.type]); }
                      else { setRulePermissions(rulePermissions.filter(p => p !== service.type)); }
                    }} />
                  <Label htmlFor={`perm-${service.type}`} className="cursor-pointer">{service.label}</Label>
                </div>
              ))}
            </div>
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setAddRuleOpen(false)}>Cancel</Button>
            <Button onClick={handleCreateRule} disabled={(newRuleType !== 'everyone' && !selectedEntity) || rulePermissions.length === 0}>Create Rule</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {selectedContact && (
        <PermissionsDialog open={permissionsDialogOpen} onClose={handleClosePermissions} contact={selectedContact} currentUserId={currentUserId || 0} />
      )}
    </>
  );
}
