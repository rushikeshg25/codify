CREATE DATABASE IF NOT EXISTS codify;

USE codify;

CREATE TABLE IF NOT EXISTS users (
    id INT NOT NULL PRIMARY KEY  AUTO_INCREMENT,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS codegrounds (
    id CHAR(36) NOT NULL PRIMARY KEY,
    user_id INT NOT NULL, 
    name VARCHAR(255) NOT NULL,
    codeground_type ENUM('REACT', 'NODE') NOT NULL,
    -- Lifecycle state, written by the codeground-manager as it provisions.
    -- Without this the frontend cannot tell "still starting" from "failed",
    -- so a broken codeground looks identical to a slow one.
    status ENUM('PROVISIONING', 'RUNNING', 'STOPPED', 'FAILED') NOT NULL DEFAULT 'PROVISIONING',
    status_detail VARCHAR(255) NULL,
    -- Refreshed by the browser while a codeground is open; the reaper uses it
    -- to find codegrounds nobody is using.
    last_active_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Existing installs: this file only runs for a fresh database, so apply the
-- new columns by hand.
--
-- ALTER TABLE codegrounds
--   ADD COLUMN status ENUM('PROVISIONING','RUNNING','STOPPED','FAILED') NOT NULL DEFAULT 'PROVISIONING',
--   ADD COLUMN status_detail VARCHAR(255) NULL,
--   ADD COLUMN last_active_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
