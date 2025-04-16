BEGIN;

CREATE TABLE IF NOT EXISTS public.ob_installations (
	id TEXT PRIMARY KEY,
	sdk_version INTEGER DEFAULT -1,
	model TEXT DEFAULT '',
	brand TEXT DEFAULT '',
	app_id INTEGER NOT NULL,
	FOREIGN KEY (app_id) REFERENCES public.ob_applications (id) 
		ON DELETE CASCADE 
		ON UPDATE NO ACTION
);

CREATE TABLE IF NOT EXISTS public.ob_sessions (
	id TEXT PRIMARY KEY,
	installation_id TEXT NOT NULL,
	created_at BIGINT NOT NULL,
	crashed SMALLINT NOT NULL DEFAULT 0,
	app_id INTEGER NOT NULL,
	FOREIGN KEY (app_id) REFERENCES public.ob_applications (id) 
		ON DELETE CASCADE ON UPDATE NO ACTION
);

COMMIT;
