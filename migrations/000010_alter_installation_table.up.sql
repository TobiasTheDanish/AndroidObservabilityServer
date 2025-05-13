ALTER TABLE IF EXISTS public.ob_installations
	ADD COLUMN IF NOT EXISTS created_at BIGINT NOT NULL DEFAULT(extract(epoch from now()));
