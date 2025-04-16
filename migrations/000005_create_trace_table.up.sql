CREATE TABLE IF NOT EXISTS public.ob_trace (
	trace_id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL,
	group_id TEXT NOT NULL,
	parent_id TEXT DEFAULT '',
	name TEXT NOT NULL,
	status TEXT NOT NULL,
	error_message TEXT DEFAULT '',
	started_at BIGINT NOT NULL,
	ended_at BIGINT NOT NULL DEFAULT 0,
	has_ended INTEGER NOT NULL DEFAULT 0,
	app_id INTEGER NOT NULL,
	FOREIGN KEY (app_id) REFERENCES public.ob_applications (id) 
		ON DELETE CASCADE ON UPDATE NO ACTION,
	FOREIGN KEY (session_id) REFERENCES public.ob_sessions (id) 
		ON DELETE NO ACTION ON UPDATE NO ACTION
);
