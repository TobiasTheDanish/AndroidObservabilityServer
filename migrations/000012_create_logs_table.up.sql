CREATE TABLE IF NOT EXISTS public.ob_logs (
	id SERIAL PRIMARY KEY,
	app_id INTEGER NOT NULL,
	session_id TEXT NOT NULL,
	message TEXT NOT NULL,
	data JSONB,
	created_at BIGINT NOT NULL,

	FOREIGN KEY (app_id) REFERENCES public.ob_applications (id) 
		ON DELETE CASCADE ON UPDATE NO ACTION,
	FOREIGN KEY (session_id) REFERENCES public.ob_sessions (id) 
		ON DELETE NO ACTION ON UPDATE NO ACTION
)
