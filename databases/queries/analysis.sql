-- Queries for the analysis table

-- name: GetAnalysisByID :one
SELECT *
FROM analysis
WHERE id_analysis = @id_analysis;

-- name: GetAnalysesByUserTelegramIDPagination :many
SELECT *
FROM analysis
WHERE id_user = @id_user
  AND (@product::TEXT IS NULL OR @product = '' OR product = @product)
  AND (@id_analysis::TEXT IS NULL OR @id_analysis = '' OR CAST(id_analysis AS TEXT) LIKE '%' || @id_analysis || '%')
ORDER BY
    CASE WHEN @sort_by = 'date_time' AND @sort_order = 'asc' THEN date_time END ASC,
    CASE WHEN @sort_by = 'date_time' AND @sort_order = 'desc' THEN date_time END DESC,
    CASE WHEN @sort_by = 'id' AND @sort_order = 'asc' THEN id END ASC,
    CASE WHEN @sort_by = 'id' AND @sort_order = 'desc' THEN id END DESC,
    CASE WHEN @sort_by = 'product' AND @sort_order = 'asc' THEN product END ASC,
    CASE WHEN @sort_by = 'product' AND @sort_order = 'desc' THEN product END DESC
LIMIT sqlc.arg('limit')::int
OFFSET sqlc.arg('offset')::int;

-- name: CountAnalysesByUserID :one
SELECT COUNT(*)
FROM analysis
WHERE id_user = @id_user
  AND (@product::TEXT IS NULL OR @product = '' OR product = @product)
  AND (@id_analysis::TEXT IS NULL OR @id_analysis = '' OR CAST(id_analysis AS TEXT) LIKE '%' || @id_analysis || '%');

-- name: GetAnalysesByIDs :many
SELECT *
FROM analysis
WHERE id = ANY(@ids::int[]);