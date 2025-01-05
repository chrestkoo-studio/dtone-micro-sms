package cache

import (
	"context"
	"database/sql"
)

func InitLocalCache(ctx context.Context, db *sql.DB) {
	InitSmsConfig(ctx, db)
}
