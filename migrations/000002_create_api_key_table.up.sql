CREATE TABLE IF NOT EXISTS public.ob_api_keys (
	key TEXT PRIMARY KEY,
	app_id INTEGER NOT NULL,
	FOREIGN KEY (app_id) REFERENCES public.ob_applications (id)
		ON DELETE CASCADE 
		ON UPDATE NO ACTION
);
