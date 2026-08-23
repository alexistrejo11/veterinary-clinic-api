-- name: GetAddressByID :one
SELECT id, user_id, street, city, state, zip_code, country, building_type,
       building_outer_number, building_inner_number, is_default,
       created_at, updated_at, deleted_at, version
FROM addresses
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetAddressByIDAndUserID :one
SELECT id, user_id, street, city, state, zip_code, country, building_type,
       building_outer_number, building_inner_number, is_default,
       created_at, updated_at, deleted_at, version
FROM addresses
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: GetAddressesByUserID :many
SELECT id, user_id, street, city, state, zip_code, country, building_type,
       building_outer_number, building_inner_number, is_default,
       created_at, updated_at, deleted_at, version
FROM addresses
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY is_default DESC, created_at DESC;

-- name: FindAddressesBySpec :many
SELECT id, user_id, street, city, state, zip_code, country, building_type,
       building_outer_number, building_inner_number, is_default,
       created_at, updated_at, deleted_at, version
FROM addresses
WHERE deleted_at IS NULL
  AND ($1::int = 0 OR user_id = $1)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAddressesByUserID :one
SELECT COUNT(*) FROM addresses
WHERE user_id = $1 AND deleted_at IS NULL;

-- name: CountAddressesBySpec :one
SELECT COUNT(*) FROM addresses
WHERE deleted_at IS NULL
  AND ($1::int = 0 OR user_id = $1);

-- name: ExistsAddressByID :one
SELECT COUNT(*) > 0 FROM addresses
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateAddress :one
WITH demoted AS (
    UPDATE addresses
    SET is_default = FALSE, updated_at = CURRENT_TIMESTAMP
    WHERE user_id = $1 AND $10::BOOLEAN AND is_default AND deleted_at IS NULL
)
INSERT INTO addresses (
    user_id, street, city, state, zip_code, country, building_type,
    building_outer_number, building_inner_number, is_default
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING id, user_id, street, city, state, zip_code, country, building_type,
          building_outer_number, building_inner_number, is_default,
          created_at, updated_at, deleted_at, version;

-- name: UpdateAddress :one
WITH updated AS (
    UPDATE addresses
    SET
        street = COALESCE($2, street),
        city = COALESCE($3, city),
        state = COALESCE($4, state),
        zip_code = COALESCE($5, zip_code),
        country = COALESCE($6, country),
        building_type = COALESCE($7, building_type),
        building_outer_number = COALESCE($8, building_outer_number),
        building_inner_number = COALESCE($9, building_inner_number),
        is_default = COALESCE($10, is_default),
        updated_at = CURRENT_TIMESTAMP,
        version = version + 1
    WHERE addresses.id = $1 AND addresses.deleted_at IS NULL
    RETURNING id, user_id, street, city, state, zip_code, country, building_type,
              building_outer_number, building_inner_number, is_default,
              created_at, updated_at, deleted_at, version
), demoted AS (
    UPDATE addresses
    SET is_default = FALSE, updated_at = CURRENT_TIMESTAMP
    WHERE addresses.user_id = (SELECT user_id FROM updated)
      AND addresses.id <> $1
      AND $10::BOOLEAN IS TRUE
      AND addresses.is_default
      AND addresses.deleted_at IS NULL
)
SELECT * FROM updated;

-- name: SoftDeleteAddress :exec
UPDATE addresses
SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: RestoreAddress :exec
UPDATE addresses
SET deleted_at = NULL, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: HardDeleteAddress :exec
DELETE FROM addresses WHERE id = $1;

-- name: UpdateAddressDefaultFlag :exec
UPDATE addresses
SET is_default = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
