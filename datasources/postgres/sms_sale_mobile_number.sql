CREATE TABLE sms_sale_mobile_number
(
    id            BIGSERIAL PRIMARY KEY,            -- Auto-incrementing Unique identifier
    sms_sale_id   BIGINT       NOT NULL,            -- Foreign key for sms_sale.id
    mobile_number VARCHAR(100) NOT NULL,            -- Mobile number
    message       TEXT         NOT NULL,            -- SMS message content
    total_cost    BIGINT       NOT NULL DEFAULT 0,  -- Total cost per message
    retry_count   INT          NOT NULL DEFAULT 0,  -- Retry count (default is 0)
    status        INT          NOT NULL DEFAULT 0,  -- Status of the sale
    created_at    BIGINT                DEFAULT 0,  -- Creation timestamp
    created_by    VARCHAR(255) NOT NULL DEFAULT '', -- User who created the entry
    updated_at    BIGINT                DEFAULT 0,  -- Last update timestamp
    updated_by    VARCHAR(255) NOT NULL DEFAULT '', -- User who updated the entry
    completed_at  BIGINT                DEFAULT 0,  -- Completed update timestamp
    completed_by  VARCHAR(255) NOT NULL DEFAULT '', -- User who Completed the entry
    rejected_at   BIGINT                DEFAULT 0,  -- Rejection timestamp
    rejected_by   VARCHAR(255) NOT NULL DEFAULT '', -- User who rejected the entry
    cancelled_at  BIGINT                DEFAULT 0,  -- Cancellation timestamp
    cancelled_by  VARCHAR(255) NOT NULL DEFAULT ''  -- User who cancelled the entry
);