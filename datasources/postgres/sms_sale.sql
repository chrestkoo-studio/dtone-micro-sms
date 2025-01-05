CREATE TABLE sms_sale
(
    id           BIGSERIAL PRIMARY KEY,            -- Auto-incrementing Unique identifier
    partner_id   BIGINT       NOT NULL,            -- Foreign key for partner
    message      TEXT         NOT NULL,            -- SMS message content
    total_cost   BIGINT       NOT NULL DEFAULT 0,  -- Total cost of the sms sale
    status       INT          NOT NULL DEFAULT 0,  -- Status of the sale
    created_at   BIGINT                DEFAULT 0,  -- Creation timestamp
    created_by   VARCHAR(255) NOT NULL DEFAULT '', -- User who created the entry
    updated_at   BIGINT                DEFAULT 0,  -- Last update timestamp
    updated_by   VARCHAR(255) NOT NULL DEFAULT '', -- User who updated the entry
    approved_at  BIGINT                DEFAULT 0,  -- Approval timestamp
    approved_by  VARCHAR(255) NOT NULL DEFAULT '',-- User who approved the entry
    rejected_at  BIGINT                DEFAULT 0,  -- Rejection timestamp
    rejected_by  VARCHAR(255) NOT NULL DEFAULT '',-- User who rejected the entry
    cancelled_at BIGINT                DEFAULT 0,  -- Cancellation timestamp
    cancelled_by VARCHAR(255) NOT NULL DEFAULT ''  -- User who cancelled the entry
);