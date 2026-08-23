-- ========================================
-- CREATE & UPDATE OPERATIONS
-- ========================================

-- name: CreateUser :one
INSERT INTO users (
    email, 
    phone_number, 
    hashed_password, 
    role, 
    status,
    name,
    gender,
    photo_url,
    bio,
    date_of_birth,
    oauth_provider,
    oauth_provider_id,
    oauth_access_token,
    oauth_refresh_token,
    oauth_token_expiry,
    email_verified,
    two_fa_method,
    two_fa_secret,
    two_fa_enabled,
    two_fa_enabled_at,
    two_fa_backup_codes,
    two_fa_backup_codes_generated_at,
    last_2fa_code_used_at,
    last_login,
    login_attempts,
    locked_until,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
    $21, $22, $23, $24, $25, $26,
    CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET 
    email = COALESCE($2, email),
    phone_number = COALESCE($3, phone_number),
    hashed_password = COALESCE($4, hashed_password),
    role = COALESCE($5, role),
    status = COALESCE($6, status),
    name = COALESCE($7, name),
    gender = COALESCE($8, gender),
    photo_url = COALESCE($9, photo_url),
    bio = COALESCE($10, bio),
    date_of_birth = COALESCE($11, date_of_birth),
    oauth_provider = COALESCE($12, oauth_provider),
    oauth_provider_id = COALESCE($13, oauth_provider_id),
    oauth_access_token = COALESCE($14, oauth_access_token),
    oauth_refresh_token = COALESCE($15, oauth_refresh_token),
    oauth_token_expiry = COALESCE($16, oauth_token_expiry),
    email_verified = COALESCE($17, email_verified),
    two_fa_method = COALESCE($18, two_fa_method),
    two_fa_secret = COALESCE($19, two_fa_secret),
    two_fa_enabled = COALESCE($20, two_fa_enabled),
    two_fa_enabled_at = COALESCE($21, two_fa_enabled_at),
    two_fa_backup_codes = COALESCE($22, two_fa_backup_codes),
    two_fa_backup_codes_generated_at = COALESCE($23, two_fa_backup_codes_generated_at),
    last_2fa_code_used_at = COALESCE($24, last_2fa_code_used_at),
    last_login = COALESCE($25, last_login),
    login_attempts = COALESCE($26, login_attempts),
    locked_until = COALESCE($27, locked_until),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- ========================================
-- FIND OPERATIONS
-- ========================================

-- name: FindUserByID :one
SELECT * FROM users 
WHERE id = $1 
AND deleted_at IS NULL;

-- name: FindUserByEmail :one
SELECT * FROM users
WHERE email = $1
AND deleted_at IS NULL;

-- name: FindUserByPhoneNumber :one
SELECT * FROM users
WHERE phone_number = $1
AND deleted_at IS NULL;

-- name: FindUserByCustomerID :one
SELECT u.*
FROM users u
INNER JOIN customers c ON u.id = c.user_id
WHERE c.id = $1
AND u.deleted_at IS NULL
AND c.deleted_at IS NULL;

-- name: FindUserByEmployeeID :one
SELECT u.*
FROM users u
INNER JOIN employees e ON u.id = e.user_id
WHERE e.id = $1
AND u.deleted_at IS NULL
AND e.deleted_at IS NULL;

-- name: FindUserByOAuthProvider :one
SELECT * FROM users
WHERE oauth_provider = $1
AND oauth_provider_id = $2
AND deleted_at IS NULL;

-- name: FindUsersBySpec :many
SELECT * FROM users
WHERE 
    ($1::INT IS NULL OR id = $1)
    AND ($2::TEXT IS NULL OR $2 = '' OR email = $2)
    AND ($3::TEXT IS NULL OR $3 = '' OR phone_number = $3)
    AND ($4::TEXT IS NULL OR $4 = '' OR role = $4)
    AND ($5::TEXT IS NULL OR $5 = '' OR status = $5)
    AND ($6::TEXT IS NULL OR $6 = '' OR oauth_provider = $6)
    AND ($7::BOOLEAN IS NULL OR email_verified = $7)
    AND ($8::BOOLEAN IS NULL OR two_fa_enabled = $8)
    AND ($9::TEXT IS NULL OR $9 = '' OR name ILIKE '%' || $9 || '%')
    AND ($10::TIMESTAMP IS NULL OR created_at >= $10)
    AND ($11::TIMESTAMP IS NULL OR created_at <= $11)
    AND ($12::TIMESTAMP IS NULL OR last_login >= $12)
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $13 OFFSET $14;

-- name: CountUsersBySpec :one
SELECT COUNT(*) FROM users
WHERE 
    ($1::INT IS NULL OR id = $1)
    AND ($2::TEXT IS NULL OR $2 = '' OR email = $2)
    AND ($3::TEXT IS NULL OR $3 = '' OR phone_number = $3)
    AND ($4::TEXT IS NULL OR $4 = '' OR role = $4)
    AND ($5::TEXT IS NULL OR $5 = '' OR status = $5)
    AND ($6::TEXT IS NULL OR $6 = '' OR oauth_provider = $6)
    AND ($7::BOOLEAN IS NULL OR email_verified = $7)
    AND ($8::BOOLEAN IS NULL OR two_fa_enabled = $8)
    AND ($9::TEXT IS NULL OR $9 = '' OR name ILIKE '%' || $9 || '%')
    AND ($10::TIMESTAMP IS NULL OR created_at >= $10)
    AND ($11::TIMESTAMP IS NULL OR created_at <= $11)
    AND ($12::TIMESTAMP IS NULL OR last_login >= $12)
    AND deleted_at IS NULL;

-- ========================================
-- EXISTS OPERATIONS
-- ========================================

-- name: ExistsUserByID :one
SELECT COUNT(*) > 0 FROM users
WHERE id = $1
AND deleted_at IS NULL;

-- name: ExistsUserByEmail :one
SELECT COUNT(*) > 0 FROM users
WHERE email = $1
AND deleted_at IS NULL;

-- name: ExistsUserByPhoneNumber :one
SELECT COUNT(*) > 0 FROM users
WHERE phone_number = $1
AND deleted_at IS NULL;

-- name: ExistsUserByCustomerID :one
SELECT COUNT(*) > 0
FROM users u
INNER JOIN customers c ON u.id = c.user_id
WHERE c.id = $1
AND u.deleted_at IS NULL
AND c.deleted_at IS NULL;

-- name: ExistsUserByEmployeeID :one
SELECT COUNT(*) > 0
FROM users u
INNER JOIN employees e ON u.id = e.user_id
WHERE e.id = $1
AND u.deleted_at IS NULL
AND e.deleted_at IS NULL;

-- ========================================
-- DELETE & RESTORE OPERATIONS
-- ========================================

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: HardDeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: RestoreUser :exec
UPDATE users
SET deleted_at = NULL
WHERE id = $1;

-- name: IsDeletedUserByID :one
SELECT deleted_at IS NOT NULL
FROM users
WHERE id = $1;
