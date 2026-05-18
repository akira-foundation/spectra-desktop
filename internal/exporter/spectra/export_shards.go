package spectra

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type shardPayload struct {
	tableName string
	jsonBytes []byte
}

func dumpJSONShards(ctx context.Context, dbPath string, tables []string) ([]shardPayload, error) {
	dsn := fmt.Sprintf("file:%s?mode=ro&_pragma=journal_mode(WAL)", dbPath)
	rawDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	defer rawDB.Close()

	shards := make([]shardPayload, 0, len(tables))
	for _, table := range tables {
		rows, err := rawDB.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %s", table))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", table, err)
		}
		columns, err := rows.Columns()
		if err != nil {
			rows.Close()
			return nil, err
		}
		records := []map[string]any{}
		for rows.Next() {
			values := make([]any, len(columns))
			scanTargets := make([]any, len(columns))
			for i := range values {
				scanTargets[i] = &values[i]
			}
			if err := rows.Scan(scanTargets...); err != nil {
				rows.Close()
				return nil, err
			}
			record := map[string]any{}
			for i, col := range columns {
				record[col] = normalizeScannedValue(values[i])
			}
			records = append(records, record)
		}
		rows.Close()
		raw, err := json.MarshalIndent(records, "", "  ")
		if err != nil {
			return nil, err
		}
		shards = append(shards, shardPayload{tableName: table, jsonBytes: raw})
	}
	return shards, nil
}

func normalizeScannedValue(v any) any {
	switch x := v.(type) {
	case []byte:
		return string(x)
	default:
		return x
	}
}
