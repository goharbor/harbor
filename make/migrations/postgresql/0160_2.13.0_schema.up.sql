ALTER TABLE p2p_preheat_policy DROP COLUMN IF EXISTS scope;
ALTER TABLE p2p_preheat_policy ADD COLUMN IF NOT EXISTS extra_attrs text;
ALTER TABLE replication_policy ADD COLUMN IF NOT EXISTS skip_if_running boolean;
