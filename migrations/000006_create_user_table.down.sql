ALTER TABLE public.ob_applications
	DROP COLUMN team_id;

DROP TABLE IF EXISTS public.ob_team_users;
DROP TABLE IF EXISTS public.ob_users;
DROP TABLE IF EXISTS public.ob_teams;
