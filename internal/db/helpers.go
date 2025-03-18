package db

import "database/sql"

// Конвертирует sql.NullFloat64 → *float64
func toFloat64Ptr(n sql.NullFloat64) *float64 {
	if n.Valid {
		return &n.Float64
	}
	return nil
}

// Конвертирует sql.NullInt64 → *int64
func toInt64Ptr(n sql.NullInt64) *int64 {
	if n.Valid {
		return &n.Int64
	}
	return nil
}
