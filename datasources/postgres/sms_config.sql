CREATE TABLE sms_config
(
    id            SERIAL PRIMARY KEY,               -- Unique identifier
    country_id    INT          NOT NULL,            -- Country ID
    mobile_prefix INT          NOT NULL,            -- Mobile number prefix
    cost          BIGINT       NOT NULL DEFAULT 0,  -- Cost for SMS
    status        INT          NOT NULL DEFAULT 0,  -- Status of the configuration
    created_at    BIGINT                DEFAULT 0,  -- Creation timestamp
    created_by    VARCHAR(255) NOT NULL DEFAULT '', -- User who created the entry
    updated_at    BIGINT                DEFAULT 0,  -- Last update timestamp
    updated_by    VARCHAR(255) NOT NULL DEFAULT ''  -- User who updated the entry
);

insert into sms_config (country_id, mobile_prefix, cost, status, created_at, created_by, updated_at, updated_by)
values (1, 60, 1, 1, EXTRACT(EPOCH FROM NOW()), 'AUTO', 0, 'AUTO');