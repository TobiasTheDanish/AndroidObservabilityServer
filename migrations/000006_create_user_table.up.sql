CREATE TABLE IF NOT EXISTS public.ob_teams (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS public.ob_users (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	pw_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS public.ob_team_users (
	id SERIAL PRIMARY KEY,
	team_id INTEGER REFERENCES public.ob_teams (id) ON DELETE CASCADE,
	user_id INTEGER REFERENCES public.ob_users (id) ON DELETE CASCADE,
	role TEXT NOT NULL DEFAULT 'owner'
);

ALTER TABLE public.ob_applications
	ADD COLUMN IF NOT EXISTS team_id INTEGER REFERENCES public.ob_teams(id);
