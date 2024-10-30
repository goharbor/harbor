ALTER TABLE p2p_preheat_policy DROP COLUMN IF EXISTS scope;
ALTER TABLE p2p_preheat_policy ADD COLUMN IF NOT EXISTS extra_attrs text;
