-- 009_default_privileges.sql
-- Ensure dis_user can operate on public schema objects.

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'dis_user') THEN
        GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO dis_user;
        GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO dis_user;

        ALTER DEFAULT PRIVILEGES IN SCHEMA public
            GRANT ALL PRIVILEGES ON TABLES TO dis_user;
        ALTER DEFAULT PRIVILEGES IN SCHEMA public
            GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO dis_user;
    END IF;
END$$;
