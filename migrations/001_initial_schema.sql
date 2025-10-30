-- Migration 001: Initial Schema
-- Creates the core tables for requests and history

-- Enable foreign key support
PRAGMA foreign_keys = ON;

-- Requests table: stores saved HTTP requests
CREATE TABLE IF NOT EXISTS requests (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    method TEXT NOT NULL,
    url TEXT NOT NULL,
    headers TEXT,          -- JSON serialized map[string]string
    query_params TEXT,     -- JSON serialized map[string]string
    body TEXT,
    auth_type TEXT,        -- Type of authentication: "none", "basic", "bearer", "apikey"
    auth_config TEXT,      -- JSON serialized auth configuration (will be encrypted in Phase 3)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- History table: tracks request execution history
CREATE TABLE IF NOT EXISTS history (
    id TEXT PRIMARY KEY,
    request_id TEXT,           -- Reference to requests table (nullable for ad-hoc requests)
    executed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status_code INTEGER,       -- HTTP status code
    status TEXT,               -- HTTP status text (e.g., "200 OK")
    response_time_ms INTEGER,  -- Response time in milliseconds
    response_headers TEXT,     -- JSON serialized response headers
    response_body TEXT,        -- Response body content
    error TEXT,                -- Error message if request failed (NULL on success)
    FOREIGN KEY (request_id) REFERENCES requests(id) ON DELETE SET NULL
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_history_executed_at ON history(executed_at DESC);
CREATE INDEX IF NOT EXISTS idx_history_request_id ON history(request_id);
CREATE INDEX IF NOT EXISTS idx_requests_name ON requests(name);
CREATE INDEX IF NOT EXISTS idx_requests_created_at ON requests(created_at DESC);
