CREATE TABLE IF NOT EXISTS public.ob_auth_sessions (
	id TEXT PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES public.ob_users(id),
	expiry BIGINT NOT NULL
);
