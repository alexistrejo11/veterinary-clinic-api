-- name: FindSessionPrescriptionByID :one
SELECT * FROM session_prescriptions WHERE id = $1;

-- name: FindSessionPrescriptionsBySessionID :many
SELECT * FROM session_prescriptions WHERE session_id = $1 ORDER BY created_at;

-- name: FindActivePrescriptionsByPetID :many
SELECT sp.* FROM session_prescriptions sp
JOIN medical_sessions ms ON ms.id = sp.session_id
WHERE ms.pet_id = $1 AND ms.deleted_at IS NULL
  AND (cardinality(sqlc.arg(employee_ids)::int[]) = 0 OR ms.employee_id = ANY(sqlc.arg(employee_ids)::int[]))
  AND (sp.duration_days IS NULL OR (sp.start_date + (sp.duration_days || ' days')::interval) >= CURRENT_DATE)
ORDER BY sp.start_date DESC
LIMIT $2 OFFSET $3;

-- name: CountActivePrescriptionsByPetID :one
SELECT COUNT(*) FROM session_prescriptions sp
JOIN medical_sessions ms ON ms.id = sp.session_id
WHERE ms.pet_id = $1 AND ms.deleted_at IS NULL
  AND (cardinality(sqlc.arg(employee_ids)::int[]) = 0 OR ms.employee_id = ANY(sqlc.arg(employee_ids)::int[]))
  AND (sp.duration_days IS NULL OR (sp.start_date + (sp.duration_days || ' days')::interval) >= CURRENT_DATE);

-- name: CreateSessionPrescription :one
INSERT INTO session_prescriptions (
    session_id, medication_id, dosage, frequency, duration_days, route, instructions, start_date
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateSessionPrescription :one
UPDATE session_prescriptions
SET session_id = $2, medication_id = $3, dosage = $4, frequency = $5, duration_days = $6, route = $7, instructions = $8, start_date = $9
WHERE id = $1
RETURNING *;

-- name: DeleteSessionPrescriptionByID :exec
DELETE FROM session_prescriptions WHERE id = $1;

-- name: DeleteSessionPrescriptionsBySessionID :exec
DELETE FROM session_prescriptions WHERE session_id = $1;
