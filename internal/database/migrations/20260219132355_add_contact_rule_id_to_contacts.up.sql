alter table contacts add column contact_rule_id bigint references contact_rules(id) on delete cascade;
create index idx_contacts_rule on contacts(contact_rule_id) where contact_rule_id is not null;
