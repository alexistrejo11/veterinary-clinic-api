-- name: FindSessionVaccinationByID :one
SELECT * FROM session_vaccinations WHERE id = $1;

-- name: FindSessionVaccinationsBySessionID :many
SELECT * FROM session_vaccinations WHERE session_id = $1 ORDER BY dose_number, created_at;

-- name: CreateSessionVaccination :one
INSERT INTO session_vaccinations (
    session_id, vaccine_catalog_id, batch_number, dose_number,
    expiration_date, site_of_injection, next_dose_date, reaction_notes, administered_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateSessionVaccination :one
UPDATE session_vaccinations
SET session_id = $2, vaccine_catalog_id = $3, batch_number = $4, dose_number = $5,
    expiration_date = $6, site_of_injection = $7, next_dose_date = $8, reaction_notes = $9, administered_by = $10
WHERE id = $1
RETURNING *;

-- name: DeleteSessionVaccinationByID :exec
DELETE FROM session_vaccinations WHERE id = $1;

-- name: DeleteSessionVaccinationsBySessionID :exec
DELETE FROM session_vaccinations WHERE session_id = $1;

-- name: FindVaccinationHistoryBySpec :many
SELECT sv.* FROM session_vaccinations sv
JOIN medical_sessions ms ON ms.id = sv.session_id
WHERE ms.deleted_at IS NULL
  AND (cardinality(sqlc.arg(pet_ids)::int[]) = 0 OR ms.pet_id = ANY(sqlc.arg(pet_ids)::int[]))
  AND (cardinality(sqlc.arg(employee_ids)::int[]) = 0 OR ms.employee_id = ANY(sqlc.arg(employee_ids)::int[]))
  AND (cardinality(sqlc.arg(vaccine_catalog_ids)::int[]) = 0 OR sv.vaccine_catalog_id = ANY(sqlc.arg(vaccine_catalog_ids)::int[]))
  AND (sqlc.narg(date_from)::timestamptz IS NULL OR ms.visit_date >= sqlc.narg(date_from)::timestamptz)
  AND (sqlc.narg(date_to)::timestamptz IS NULL OR ms.visit_date <= sqlc.narg(date_to)::timestamptz)
ORDER BY ms.visit_date DESC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountVaccinationHistoryBySpec :one
SELECT COUNT(*) FROM session_vaccinations sv
JOIN medical_sessions ms ON ms.id = sv.session_id
WHERE ms.deleted_at IS NULL
  AND (cardinality(sqlc.arg(pet_ids)::int[]) = 0 OR ms.pet_id = ANY(sqlc.arg(pet_ids)::int[]))
  AND (cardinality(sqlc.arg(employee_ids)::int[]) = 0 OR ms.employee_id = ANY(sqlc.arg(employee_ids)::int[]))
  AND (cardinality(sqlc.arg(vaccine_catalog_ids)::int[]) = 0 OR sv.vaccine_catalog_id = ANY(sqlc.arg(vaccine_catalog_ids)::int[]))
  AND (sqlc.narg(date_from)::timestamptz IS NULL OR ms.visit_date >= sqlc.narg(date_from)::timestamptz)
  AND (sqlc.narg(date_to)::timestamptz IS NULL OR ms.visit_date <= sqlc.narg(date_to)::timestamptz);
