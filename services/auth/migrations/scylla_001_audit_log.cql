-- ScyllaDB Keyspace and Audit Log Table for Auth Service

CREATE KEYSPACE IF NOT EXISTS dealance_auth
    WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

CREATE TABLE IF NOT EXISTS dealance_auth.security_audit_log (
    user_id    UUID,
    event_at   TIMESTAMP,
    event_id   UUID,
    device_id  TEXT,
    event_type TEXT,
    event_data TEXT,
    ip_address TEXT,
    risk_score DECIMAL,
    PRIMARY KEY ((user_id), event_at, event_id)
) WITH CLUSTERING ORDER BY (event_at DESC, event_id DESC)
  AND default_time_to_live = 7776000;
