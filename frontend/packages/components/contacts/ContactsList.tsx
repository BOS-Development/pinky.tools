import { useState, useEffect, useRef } from 'react';
import { useSession } from "next-auth/react";
import Navbar from "@industry-tool/components/Navbar";
import Loading from "@industry-tool/components/loading";
import Container from '@mui/material/Container';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Button from '@mui/material/Button';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import DeleteIcon from '@mui/icons-material/Delete';
import CheckIcon from '@mui/icons-material/Check';
import CloseIcon from '@mui/icons-material/Close';
import SettingsIcon from '@mui/icons-material/Settings';
import PersonAddIcon from '@mui/icons-material/PersonAdd';
import RuleIcon from '@mui/icons-material/Rule';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import Autocomplete from '@mui/material/Autocomplete';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
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
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

type SearchResult = {
  id: number;
  name: string;
};

export default function ContactsList() {
  const { data: session } = useSession();
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [contactRules, setContactRules] = useState<ContactRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [tabIndex, setTabIndex] = useState(0);
  const [snackbarOpen, setSnackbarOpen] = useState(false);
  const [snackbarMessage, setSnackbarMessage] = useState('');
  const [snackbarSeverity, setSnackbarSeverity] = useState<'success' | 'error'>('success');
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
        showSnackbar('Contact request sent!', 'success');
        setAddContactOpen(false);
        setNewContactCharacterName('');
        await fetchContacts();
      } else {
        const error = await response.json();
        showSnackbar(error.error || 'Failed to send contact request', 'error');
      }
    } catch (err) {
      showSnackbar('Failed to send contact request', 'error');
    }
  };

  const handleAccept = async (contactId: number) => {
    try {
      const response = await fetch(`/api/contacts/${contactId}/accept`, {
        method: 'POST',
      });

      if (response.ok) {
        showSnackbar('Contact accepted!', 'success');
        await fetchContacts();
      } else {
        const error = await response.json();
        showSnackbar(error.error || 'Failed to accept contact', 'error');
      }
    } catch (err) {
      showSnackbar('Failed to accept contact', 'error');
    }
  };

  const handleReject = async (contactId: number) => {
    try {
      const response = await fetch(`/api/contacts/${contactId}/reject`, {
        method: 'POST',
      });

      if (response.ok) {
        showSnackbar('Contact rejected', 'success');
        await fetchContacts();
      } else {
        const error = await response.json();
        showSnackbar(error.error || 'Failed to reject contact', 'error');
      }
    } catch (err) {
      showSnackbar('Failed to reject contact', 'error');
    }
  };

  const handleDelete = async (contactId: number) => {
    try {
      const response = await fetch(`/api/contacts/${contactId}`, {
        method: 'DELETE',
      });

      if (response.ok) {
        showSnackbar('Contact removed', 'success');
        await fetchContacts();
      } else {
        const error = await response.json();
        showSnackbar(error.error || 'Failed to remove contact', 'error');
      }
    } catch (err) {
      showSnackbar('Failed to remove contact', 'error');
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

  const handleSearchInputChange = (_event: React.SyntheticEvent, value: string) => {
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
      showSnackbar('Please select a corporation or alliance', 'error');
      return;
    }

    try {
      const body: { ruleType: string; entityId?: number; entityName?: string } = {
        ruleType: newRuleType,
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
        showSnackbar('Contact rule created! Auto-contacts are being generated.', 'success');
        setAddRuleOpen(false);
        setNewRuleType('corporation');
        setSelectedEntity(null);
        setSearchQuery('');
        setSearchResults([]);
        await fetchContactRules();
        // Refresh contacts after a short delay to show auto-created ones
        setTimeout(() => fetchContacts(), 2000);
      } else {
        const error = await response.json();
        showSnackbar(error.error || 'Failed to create contact rule', 'error');
      }
    } catch (err) {
      showSnackbar('Failed to create contact rule', 'error');
    }
  };

  const handleDeleteRule = async (ruleId: number) => {
    try {
      const response = await fetch(`/api/contact-rules/${ruleId}`, {
        method: 'DELETE',
      });

      if (response.ok) {
        showSnackbar('Contact rule removed', 'success');
        await fetchContactRules();
        await fetchContacts();
      } else {
        const error = await response.json();
        showSnackbar(error.error || 'Failed to remove contact rule', 'error');
      }
    } catch (err) {
      showSnackbar('Failed to remove contact rule', 'error');
    }
  };

  const showSnackbar = (message: string, severity: 'success' | 'error') => {
    setSnackbarMessage(message);
    setSnackbarSeverity(severity);
    setSnackbarOpen(true);
  };

  const getRuleTypeColor = (ruleType: string): 'primary' | 'secondary' | 'success' => {
    switch (ruleType) {
      case 'corporation': return 'primary';
      case 'alliance': return 'secondary';
      case 'everyone': return 'success';
      default: return 'primary';
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

  // Filter contacts by tab
  const myContacts = contacts.filter(c => c.status === 'accepted');
  const pendingRequests = contacts.filter(c =>
    c.status === 'pending' && c.recipientUserId === currentUserId
  );
  const sentRequests = contacts.filter(c =>
    c.status === 'pending' && c.requesterUserId === currentUserId
  );

  if (!session) {
    return null;
  }

  if (loading) {
    return <Loading />;
  }

  return (
    <>
      <Navbar />
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
          <Typography variant="h4">Contacts</Typography>
          <Box sx={{ display: 'flex', gap: 1 }}>
            <Button
              variant="outlined"
              startIcon={<RuleIcon />}
              onClick={() => setAddRuleOpen(true)}
            >
              Add Rule
            </Button>
            <Button
              variant="contained"
              startIcon={<PersonAddIcon />}
              onClick={() => setAddContactOpen(true)}
            >
              Add Contact
            </Button>
          </Box>
        </Box>

        <Card>
          <Tabs value={tabIndex} onChange={(_, newValue) => setTabIndex(newValue)}>
            <Tab label={`My Contacts (${myContacts.length})`} />
            <Tab label={`Pending Requests (${pendingRequests.length})`} />
            <Tab label={`Sent Requests (${sentRequests.length})`} />
            <Tab label={`Contact Rules (${contactRules.length})`} />
          </Tabs>

          <CardContent>
            {/* My Contacts Tab */}
            {tabIndex === 0 && (
              <TableContainer component={Paper} variant="outlined">
                {myContacts.length === 0 ? (
                  <Box sx={{ p: 4, textAlign: 'center' }}>
                    <Typography color="text.secondary">
                      No contacts yet. Add a contact to get started!
                    </Typography>
                  </Box>
                ) : (
                  <Table>
                    <TableHead>
                      <TableRow>
                        <TableCell>Character Name</TableCell>
                        <TableCell>Status</TableCell>
                        <TableCell>Connected Since</TableCell>
                        <TableCell align="right">Actions</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {myContacts.map((contact) => {
                        const otherUserName = contact.requesterUserId === currentUserId
                          ? contact.recipientName
                          : contact.requesterName;

                        return (
                          <TableRow key={contact.id} hover>
                            <TableCell>
                              {otherUserName}
                              {contact.contactRuleId && (
                                <Chip
                                  label="Auto"
                                  size="small"
                                  color="info"
                                  variant="outlined"
                                  sx={{ ml: 1 }}
                                />
                              )}
                            </TableCell>
                            <TableCell>
                              <Chip label="Connected" color="success" size="small" />
                            </TableCell>
                            <TableCell>
                              {new Date(contact.respondedAt || contact.requestedAt).toLocaleDateString()}
                            </TableCell>
                            <TableCell align="right">
                              <IconButton
                                size="small"
                                onClick={() => handleOpenPermissions(contact)}
                                title="Manage Permissions"
                              >
                                <SettingsIcon />
                              </IconButton>
                              <IconButton
                                size="small"
                                onClick={() => handleDelete(contact.id)}
                                title="Remove Contact"
                                color="error"
                              >
                                <DeleteIcon />
                              </IconButton>
                            </TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                )}
              </TableContainer>
            )}

            {/* Pending Requests Tab */}
            {tabIndex === 1 && (
              <TableContainer component={Paper} variant="outlined">
                {pendingRequests.length === 0 ? (
                  <Box sx={{ p: 4, textAlign: 'center' }}>
                    <Typography color="text.secondary">
                      No pending requests
                    </Typography>
                  </Box>
                ) : (
                  <Table>
                    <TableHead>
                      <TableRow>
                        <TableCell>Character Name</TableCell>
                        <TableCell>Requested</TableCell>
                        <TableCell align="right">Actions</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {pendingRequests.map((contact) => (
                        <TableRow key={contact.id} hover>
                          <TableCell>{contact.requesterName}</TableCell>
                          <TableCell>
                            {new Date(contact.requestedAt).toLocaleDateString()}
                          </TableCell>
                          <TableCell align="right">
                            <IconButton
                              size="small"
                              onClick={() => handleAccept(contact.id)}
                              title="Accept"
                              color="success"
                            >
                              <CheckIcon />
                            </IconButton>
                            <IconButton
                              size="small"
                              onClick={() => handleReject(contact.id)}
                              title="Reject"
                              color="error"
                            >
                              <CloseIcon />
                            </IconButton>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                )}
              </TableContainer>
            )}

            {/* Sent Requests Tab */}
            {tabIndex === 2 && (
              <TableContainer component={Paper} variant="outlined">
                {sentRequests.length === 0 ? (
                  <Box sx={{ p: 4, textAlign: 'center' }}>
                    <Typography color="text.secondary">
                      No sent requests
                    </Typography>
                  </Box>
                ) : (
                  <Table>
                    <TableHead>
                      <TableRow>
                        <TableCell>Character Name</TableCell>
                        <TableCell>Sent</TableCell>
                        <TableCell align="right">Actions</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {sentRequests.map((contact) => (
                        <TableRow key={contact.id} hover>
                          <TableCell>{contact.recipientName}</TableCell>
                          <TableCell>
                            {new Date(contact.requestedAt).toLocaleDateString()}
                          </TableCell>
                          <TableCell align="right">
                            <IconButton
                              size="small"
                              onClick={() => handleDelete(contact.id)}
                              title="Cancel Request"
                              color="error"
                            >
                              <CloseIcon />
                            </IconButton>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                )}
              </TableContainer>
            )}

            {/* Contact Rules Tab */}
            {tabIndex === 3 && (
              <TableContainer component={Paper} variant="outlined">
                {contactRules.length === 0 ? (
                  <Box sx={{ p: 4, textAlign: 'center' }}>
                    <Typography color="text.secondary">
                      No contact rules yet. Add a rule to automatically connect with corporations, alliances, or everyone.
                    </Typography>
                  </Box>
                ) : (
                  <Table>
                    <TableHead>
                      <TableRow>
                        <TableCell>Type</TableCell>
                        <TableCell>Entity</TableCell>
                        <TableCell>Created</TableCell>
                        <TableCell align="right">Actions</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {contactRules.map((rule) => (
                        <TableRow key={rule.id} hover>
                          <TableCell>
                            <Chip
                              label={getRuleTypeLabel(rule.ruleType)}
                              color={getRuleTypeColor(rule.ruleType)}
                              size="small"
                            />
                          </TableCell>
                          <TableCell>
                            {rule.ruleType === 'everyone' ? 'All Users' : (rule.entityName || `ID: ${rule.entityId}`)}
                          </TableCell>
                          <TableCell>
                            {new Date(rule.createdAt).toLocaleDateString()}
                          </TableCell>
                          <TableCell align="right">
                            <IconButton
                              size="small"
                              onClick={() => handleDeleteRule(rule.id)}
                              title="Remove Rule"
                              color="error"
                            >
                              <DeleteIcon />
                            </IconButton>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                )}
              </TableContainer>
            )}
          </CardContent>
        </Card>
      </Container>

      {/* Add Contact Dialog */}
      <Dialog open={addContactOpen} onClose={() => setAddContactOpen(false)}>
        <DialogTitle>Add Contact</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Character Name"
            type="text"
            fullWidth
            value={newContactCharacterName}
            onChange={(e) => setNewContactCharacterName(e.target.value)}
            helperText="Enter the character name of the person you want to add"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setAddContactOpen(false)}>Cancel</Button>
          <Button onClick={handleAddContact} variant="contained">
            Send Request
          </Button>
        </DialogActions>
      </Dialog>

      {/* Add Rule Dialog */}
      <Dialog open={addRuleOpen} onClose={() => setAddRuleOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Add Contact Rule</DialogTitle>
        <DialogContent>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Contact rules automatically create connections with all members of a corporation, alliance, or everyone.
            You will grant them permission to browse your for-sale items.
          </Typography>

          <FormControl fullWidth sx={{ mb: 2, mt: 1 }}>
            <InputLabel>Rule Type</InputLabel>
            <Select
              value={newRuleType}
              label="Rule Type"
              onChange={(e) => {
                setNewRuleType(e.target.value);
                setSelectedEntity(null);
                setSearchQuery('');
                setSearchResults([]);
              }}
            >
              <MenuItem value="corporation">Corporation</MenuItem>
              <MenuItem value="alliance">Alliance</MenuItem>
              <MenuItem value="everyone">Everyone</MenuItem>
            </Select>
          </FormControl>

          {newRuleType !== 'everyone' && (
            <Autocomplete
              options={searchResults}
              getOptionLabel={(option) => option.name}
              value={selectedEntity}
              onChange={(_event, newValue) => setSelectedEntity(newValue)}
              inputValue={searchQuery}
              onInputChange={handleSearchInputChange}
              loading={searchLoading}
              noOptionsText={searchQuery.length < 2 ? "Type to search..." : "No results found"}
              isOptionEqualToValue={(option, value) => option.id === value.id}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label={newRuleType === 'corporation' ? 'Search Corporations' : 'Search Alliances'}
                  helperText={`Search for a ${newRuleType} by name`}
                />
              )}
            />
          )}

          {newRuleType === 'everyone' && (
            <Alert severity="info">
              This will automatically connect you with every user in the system and grant them permission to browse your for-sale items.
            </Alert>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setAddRuleOpen(false)}>Cancel</Button>
          <Button
            onClick={handleCreateRule}
            variant="contained"
            disabled={newRuleType !== 'everyone' && !selectedEntity}
          >
            Create Rule
          </Button>
        </DialogActions>
      </Dialog>

      {/* Permissions Dialog */}
      {selectedContact && (
        <PermissionsDialog
          open={permissionsDialogOpen}
          onClose={handleClosePermissions}
          contact={selectedContact}
          currentUserId={currentUserId || 0}
        />
      )}

      {/* Snackbar */}
      <Snackbar
        open={snackbarOpen}
        autoHideDuration={3000}
        onClose={() => setSnackbarOpen(false)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert
          onClose={() => setSnackbarOpen(false)}
          severity={snackbarSeverity}
          sx={{ width: '100%' }}
        >
          {snackbarMessage}
        </Alert>
      </Snackbar>
    </>
  );
}
