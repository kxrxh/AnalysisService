-- Queries for the objects table
-- name: GetObjectByID :one
SELECT *
FROM objects
WHERE id = sqlc.arg(id);

-- name: GetObjectsByAnalysisID :many
SELECT *
FROM objects
WHERE id_analysis = sqlc.arg(analysis_id)
ORDER BY id;

-- name: GetObjectsImages :many
SELECT id, id_analysis, file
FROM objects
WHERE id = ANY(sqlc.arg(ids)::int[]);

-- name: GetObjectsImagesForAnalysis :many
SELECT id, id_analysis, file
FROM objects
WHERE id_analysis = sqlc.arg(id_analysis);

-- name: GetObjectsMetadata :many
SELECT id, id_analysis, m_h, m_s, m_v, m_r, m_g, m_b, l_avg, w_avg, brt_avg, r_avg, g_avg, b_avg, h_avg, s_avg, v_avg, h, s, v, h_m, s_m, v_m, r_m, g_m, b_m, brt_m, w_m, l_m, l, w, l_w, pr, sq, brt, r, g, b, solid, min_h, min_s, min_v, max_h, max_s, max_v, entropy, id_image, color_rhs, geometry, sq_sqcrl, hu1, hu2, hu3, hu4, hu5, hu6
FROM objects
WHERE id = ANY(sqlc.arg(ids)::int[]);

-- name: GetObjectsMetadataForAnalysis :many
SELECT id, id_analysis, m_h, m_s, m_v, m_r, m_g, m_b, l_avg, w_avg, brt_avg, r_avg, g_avg, b_avg, h_avg, s_avg, v_avg, h, s, v, h_m, s_m, v_m, r_m, g_m, b_m, brt_m, w_m, l_m, l, w, l_w, pr, sq, brt, r, g, b, solid, min_h, min_s, min_v, max_h, max_s, max_v, entropy, id_image, color_rhs, geometry, sq_sqcrl, hu1, hu2, hu3, hu4, hu5, hu6
FROM objects
WHERE id_analysis = sqlc.arg(id_analysis);

-- name: GetObjectsByIDs :many
SELECT
    o.*,
    a.id AS analysis_id,
    a.date_time AS analysis_date_time,
    a.product AS analysis_product,
    a.color_rhs AS analysis_color_rhs,
    a.id_user AS analysis_id_user,
    a.telegram_link AS analysis_telegram_link,
    a.text AS analysis_text,
    a.scale_mm_pixel AS analysis_scale_mm_pixel,
    a.mass AS analysis_mass,
    a.area AS analysis_area,
    a.r AS analysis_r,
    a.g AS analysis_g,
    a.b AS analysis_b,
    a.h AS analysis_h,
    a.s AS analysis_s,
    a.v AS analysis_v,
    a.lab_l AS analysis_lab_l,
    a.lab_a AS analysis_lab_a,
    a.lab_b AS analysis_lab_b,
    a.w AS analysis_w,
    a.l AS analysis_l,
    a.t AS analysis_t,
    a.id_analysis AS analysis_id_analysis
FROM objects o
LEFT JOIN analysis a ON o.id_analysis = a.id
WHERE o.id = ANY(sqlc.arg(ids)::int[]);