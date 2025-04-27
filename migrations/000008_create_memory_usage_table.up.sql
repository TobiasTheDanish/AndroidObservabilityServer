CREATE TABLE IF NOT EXISTS public.ob_memory_usage (
	id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL REFERENCES public.ob_sessions(id) ON DELETE CASCADE,
	installation_id TEXT NOT NULL,
	app_id INTEGER NOT NULL REFERENCES public.ob_applications(id) ON DELETE CASCADE,
	used_memory BIGINT NOT NULL,
	free_memory BIGINT NOT NULL,
	max_memory BIGINT NOT NULL,
	total_memory BIGINT NOT NULL,
	available_heap_space BIGINT NOT NULL
);
