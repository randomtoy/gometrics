-- name: InsertOrUpdateMetric :exec
INSERT INTO metrics (id, type, value, delta)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE 
SET value = EXCLUDED.value, delta = EXCLUDED.delta;

-- name: GetMetric :one
SELECT id, type, value, delta FROM metrics WHERE id = $1;

-- name: GetAllMetrics :many
SELECT id, type, value, delta FROM metrics;

-- name: InsertOrUpdateMetricBatch :execparams
INSERT INTO metrics (id, type, value, delta)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE
SET value = EXCLUDED.value, delta = EXCLUDED.delta;