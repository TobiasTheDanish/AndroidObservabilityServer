CREATE TABLE IF NOT EXISTS public.ob_events (
	id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL,
	created_at BIGINT NOT NULL,
	type TEXT NOT NULL,
	serialized_data TEXT DEFAULT '',
	app_id INTEGER NOT NULL,
	FOREIGN KEY (app_id) REFERENCES public.ob_applications (id) 
		ON DELETE CASCADE ON UPDATE NO ACTION,
	FOREIGN KEY (session_id) REFERENCES public.ob_sessions (id) 
		ON DELETE NO ACTION ON UPDATE NO ACTION
);
