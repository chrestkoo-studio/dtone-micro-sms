CREATE TABLE sms_log
(
    id          BIGSERIAL PRIMARY KEY,             -- Auto-incrementing unique identifier
    partner_id  BIGINT       NOT NULL,             -- Foreign key for partner
    doc_no      VARCHAR(255) NOT NULL,             -- Document number
    mobile_no   VARCHAR(100) NOT NULL,             -- Username, assuming max length of 100
    message     TEXT         NOT NULL,             -- Message stored as text
    url         VARCHAR(255) NOT NULL,             -- url stored as text
    input       TEXT         NOT NULL,             -- Input, stored as text
    output      TEXT NULL,                         -- Output, stored as text
    retry_count INT          NOT NULL,             -- RetryCount, stored as a 32-bit integer
    status      INT          NOT NULL DEFAULT (0), -- Status, stored as a 32-bit integer
    created_at  BIGINT       NOT NULL DEFAULT (0), -- Created Unix timestamp, stored as an unsigned 64-bit integer
    updated_at  BIGINT       NOT NULL DEFAULT (0)  -- Updated Unix timestamp, stored as an unsigned 64-bit integer
);