-- name: FindEmployeeByID :one
SELECT * FROM employees
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindEmployeeByUserID :one
SELECT * FROM employees
WHERE user_id = $1 AND deleted_at IS NULL;

-- name: FindEmployees :many
SELECT * FROM employees
WHERE deleted_at IS NULL
  AND ($1::int = -1 OR ($1 = 1 AND is_active = true) OR ($1 = 0 AND is_active = false))
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: FindActiveEmployees :many
SELECT * FROM employees
WHERE is_active = TRUE AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: FindEmployeesBySpeciality :many
SELECT * FROM employees
WHERE speciality = $1 AND deleted_at IS NULL
  AND ($2::int = -1 OR ($2 = 1 AND is_active = true) OR ($2 = 0 AND is_active = false))
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CreateEmployee :one
INSERT INTO employees (
    first_name, last_name, photo, license_number, speciality,
    years_of_experience, is_active, user_id, schedule_json, gender, date_of_birth
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, $11
)
RETURNING *;

-- name: UpdateEmployee :one
UPDATE employees
SET
    first_name = COALESCE($2, first_name),
    last_name = COALESCE($3, last_name),
    photo = COALESCE($4, photo),
    license_number = COALESCE($5, license_number),
    speciality = COALESCE($6, speciality),
    years_of_experience = COALESCE($7, years_of_experience),
    is_active = COALESCE($8, is_active),
    user_id = COALESCE($9, user_id),
    schedule_json = COALESCE($10::jsonb, schedule_json),
    gender = COALESCE($11, gender),
    date_of_birth = COALESCE($12, date_of_birth),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteEmployee :exec
UPDATE employees
SET
    deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP,
    is_active = FALSE
WHERE id = $1;

-- name: HardDeleteEmployee :exec
DELETE FROM employees WHERE id = $1;

-- name: RestoreEmployee :one
UPDATE employees
SET deleted_at = NULL, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NOT NULL
RETURNING id;

-- name: ExistsEmployeeByID :one
SELECT COUNT(*) > 0 FROM employees
WHERE id = $1 AND deleted_at IS NULL;

-- name: ExistsEmployeeByUserID :one
SELECT COUNT(*) > 0 FROM employees
WHERE user_id = $1 AND deleted_at IS NULL;

-- name: CountActiveEmployees :one
SELECT COUNT(*) FROM employees
WHERE is_active = TRUE AND deleted_at IS NULL;

-- name: CountAllEmployees :one
SELECT COUNT(*) FROM employees
WHERE deleted_at IS NULL;

-- name: CountEmployeesByActive :one
SELECT COUNT(*) FROM employees
WHERE deleted_at IS NULL
  AND ($1::int = -1 OR ($1 = 1 AND is_active = true) OR ($1 = 0 AND is_active = false));


-- name: CountEmployeesBySpeciality :one
SELECT COUNT(*) FROM employees
WHERE speciality = $1 AND deleted_at IS NULL
  AND ($2::int = -1 OR ($2 = 1 AND is_active = true) OR ($2 = 0 AND is_active = false));
