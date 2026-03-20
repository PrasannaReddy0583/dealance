-- Auto-create databases for each microservice
-- This script runs on first PostgreSQL startup
-- Using CREATE IF NOT EXISTS pattern to be idempotent

SELECT 'CREATE DATABASE dealance_auth' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'dealance_auth')\gexec
SELECT 'CREATE DATABASE dealance_user' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'dealance_user')\gexec
SELECT 'CREATE DATABASE dealance_content' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'dealance_content')\gexec
SELECT 'CREATE DATABASE dealance_startup' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'dealance_startup')\gexec
SELECT 'CREATE DATABASE dealance_deal' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'dealance_deal')\gexec
SELECT 'CREATE DATABASE dealance_chat' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'dealance_chat')\gexec
SELECT 'CREATE DATABASE dealance_media' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'dealance_media')\gexec
SELECT 'CREATE DATABASE dealance_feed' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'dealance_feed')\gexec
SELECT 'CREATE DATABASE dealance_notify' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'dealance_notify')\gexec
SELECT 'CREATE DATABASE dealance_admin' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'dealance_admin')\gexec
SELECT 'CREATE DATABASE dealance_wallet' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'dealance_wallet')\gexec

-- Grant all privileges to the dealance user
GRANT ALL PRIVILEGES ON DATABASE dealance_auth TO dealance;
GRANT ALL PRIVILEGES ON DATABASE dealance_user TO dealance;
GRANT ALL PRIVILEGES ON DATABASE dealance_content TO dealance;
GRANT ALL PRIVILEGES ON DATABASE dealance_startup TO dealance;
GRANT ALL PRIVILEGES ON DATABASE dealance_deal TO dealance;
GRANT ALL PRIVILEGES ON DATABASE dealance_wallet TO dealance;
GRANT ALL PRIVILEGES ON DATABASE dealance_chat TO dealance;
GRANT ALL PRIVILEGES ON DATABASE dealance_media TO dealance;
GRANT ALL PRIVILEGES ON DATABASE dealance_feed TO dealance;
GRANT ALL PRIVILEGES ON DATABASE dealance_notify TO dealance;
GRANT ALL PRIVILEGES ON DATABASE dealance_admin TO dealance;
