-- name: FindAppointmentsBySpec :many
SELECT 
    id, clinic_service, scheduled_date, status, reason, notes, cancellation_reason,
    customer_id, employee_id, pet_id, created_at, updated_at
FROM appointments
WHERE 
    ($1 = 0 OR id = $1)
    AND ($2 = 0 OR customer_id = $2)
    AND ($3 = 0 OR employee_id = $3)
    AND ($4 = 0 OR pet_id = $4)
    AND (
        $5::TEXT IS NULL OR 
        $5::TEXT = '' OR 
        clinic_service = $5  
    )
    AND (
        $6::TEXT IS NULL OR 
        $6::TEXT = '' OR 
        status = $6::appointment_status      
    )
    AND ($7::timestamp IS NULL OR scheduled_date >= $7)
    AND ($8::timestamp IS NULL OR scheduled_date <= $8)
    AND ($9::timestamp IS NULL OR scheduled_date = $9)
    AND deleted_at IS NULL
ORDER BY scheduled_date DESC
LIMIT $10 OFFSET $11;

-- name: CountAppointmentsBySpec :one
SELECT COUNT(*) 
FROM appointments 
WHERE 
    ($1 = 0 OR id = $1)
    AND ($2 = 0 OR customer_id = $2)
    AND ($3 = 0 OR employee_id = $3)
    AND ($4 = 0 OR pet_id = $4)
    AND (
        $5::TEXT IS NULL OR 
        $5::TEXT = '' OR 
        clinic_service = $5
    )
    AND (
        $6::TEXT IS NULL OR 
        $6::TEXT = '' OR 
        status = $6::appointment_status    
    )
    AND ($7::timestamp IS NULL OR scheduled_date >= $7)
    AND ($8::timestamp IS NULL OR scheduled_date <= $8)
    AND ($9::timestamp IS NULL OR scheduled_date = $9)
    AND deleted_at IS NULL;



-- name: FindAppointmentByID :one
SELECT * FROM appointments
WHERE id = $1 
AND deleted_at IS NULL;

-- name: FindAppointmentByIDAndCustomerID :one
SELECT * FROM appointments
WHERE id = $1 
AND customer_id = $2
AND deleted_at IS NULL;

-- name: FindAppointmentByIDAndEmployeeID :one
SELECT * FROM appointments
WHERE id = $1 
AND employee_id = $2
AND deleted_at IS NULL;

-- name: CreateAppointment :one
INSERT INTO appointments (
    clinic_service, 
    scheduled_date, 
    status, 
    reason,
    notes, 
    cancellation_reason,
    customer_id, 
    employee_id,
    pet_id,
    created_at,
    updated_at,
    deleted_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, NULL
) RETURNING *;

-- name: UpdateAppointment :one
UPDATE appointments SET
    clinic_service = $2,
    scheduled_date = $3,
    status = $4,
    reason = $5,
    notes = $6,
    cancellation_reason = $7,
    customer_id = $8,
    employee_id = $9,
    pet_id = $10,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: SoftDeleteAppointment :exec
UPDATE appointments SET
    deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;


-- name: HardDeleteAppointment :exec
DELETE FROM appointments WHERE id = $1;

-- name: RestoreAppointment :exec
UPDATE appointments SET
    deleted_at = NULL
WHERE id = $1;

-- name: ExistsAppointmentID :one
SELECT COUNT(*) > 0 FROM appointments
WHERE id = $1 AND deleted_at IS NULL;


-- name: ExistsConflictingAppointment :one
SELECT COUNT(*) > 0 FROM appointments
WHERE (scheduled_date BETWEEN $1 AND $2) AND employee_id = $3;
