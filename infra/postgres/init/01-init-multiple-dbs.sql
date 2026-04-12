-- Creates separate databases and roles for GuildBanker and Keycloak.
-- This script is executed ONLY on first container init (empty volume).

-- App DB + role
DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'guildbanker') THEN
			CREATE ROLE guildbanker LOGIN PASSWORD 'guildbanker';
		END IF;
	END $$;

SELECT 'CREATE DATABASE guildbanker OWNER guildbanker'
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'guildbanker')\gexec

-- Keycloak DB + role
DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'keycloak') THEN
			CREATE ROLE keycloak LOGIN PASSWORD 'keycloak';
		END IF;
	END $$;

SELECT 'CREATE DATABASE keycloak OWNER keycloak'
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'keycloak')\gexec
