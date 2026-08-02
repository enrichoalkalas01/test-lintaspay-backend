-- 0001_init.up.sql
-- Initial schema: users, refresh_tokens, disbursements, idempotency_keys, audit_logs.
-- Safe to re-run: uses IF NOT EXISTS and idempotent seed inserts.

CREATE TABLE IF NOT EXISTS users (
    id CHAR(36) NOT NULL,
    username VARCHAR(50) NOT NULL,
    password_hash VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uq_users_username (username)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id CHAR(36) NOT NULL,
    user_id CHAR(36) NOT NULL,
    token_hash CHAR(64) NOT NULL,
    expires_at DATETIME(3) NOT NULL,
    revoked_at DATETIME(3) NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uq_refresh_tokens_token_hash (token_hash),
    KEY idx_refresh_tokens_user_id (user_id),
    CONSTRAINT fk_refresh_tokens_user FOREIGN KEY (user_id) REFERENCES users (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS disbursements (
    id CHAR(36) NOT NULL,
    recipient_name VARCHAR(255) NOT NULL,
    account_number VARCHAR(50) NOT NULL,
    bank_code VARCHAR(20) NOT NULL,
    amount BIGINT NOT NULL,
    admin_fee BIGINT NOT NULL,
    note VARCHAR(500) NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_by VARCHAR(50) NOT NULL,
    approved_by VARCHAR(50) NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    PRIMARY KEY (id),
    KEY idx_disbursements_status (status),
    KEY idx_disbursements_created_at (created_at),
    KEY idx_disbursements_recipient_name (recipient_name),
    KEY idx_disbursements_deleted_at (deleted_at),
    CONSTRAINT chk_disbursements_amount CHECK (amount >= 10000),
    CONSTRAINT chk_disbursements_status CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED'))
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

-- primary key on `key` is what makes the idempotency reservation atomic:
-- a concurrent insert with the same key fails with a duplicate entry error
CREATE TABLE IF NOT EXISTS idempotency_keys (
    `key` CHAR(36) NOT NULL,
    user_id CHAR(36) NOT NULL,
    request_hash CHAR(64) NOT NULL,
    response_code INT NOT NULL,
    response_body JSON NULL,
    expires_at DATETIME(3) NOT NULL,
    created_at DATETIME(3) NOT NULL,
    PRIMARY KEY (`key`),
    KEY idx_idempotency_keys_expires_at (expires_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS audit_logs (
    id CHAR(36) NOT NULL,
    entity_id CHAR(36) NOT NULL,
    action VARCHAR(30) NOT NULL,
    actor VARCHAR(50) NOT NULL,
    `before` JSON NULL,
    `after` JSON NULL,
    created_at DATETIME(3) NOT NULL,
    PRIMARY KEY (id),
    KEY idx_audit_logs_entity_id (entity_id),
    KEY idx_audit_logs_action (action),
    KEY idx_audit_logs_created_at (created_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

-- seed users (passwords: superadmin123 / admin123 / operator123)
INSERT INTO users (id, username, password_hash, role) VALUES
    ('0d9f2f61-4c2c-4a3e-9c26-6c9edcecae01', 'superadmin', '$2a$10$rs5TMQrlN1SNXlO4jctn2uEF7jxhwR.GCO8SOCDmoToNG6OJSmZ/C', 'superadmin'),
    ('0d9f2f61-4c2c-4a3e-9c26-6c9edcecae02', 'admin', '$2a$10$91CATpMpml7JdQAxldhJWeKsq2yljXIxBmIst531YsPHg0Tp4eHTy', 'admin'),
    ('0d9f2f61-4c2c-4a3e-9c26-6c9edcecae03', 'operator', '$2a$10$tAVWusve6o.QeIxkmQ85qeYapEKELxoFfjLGyX0jxC4GxQoPDCIT6', 'operator')
ON DUPLICATE KEY UPDATE username = VALUES(username);
