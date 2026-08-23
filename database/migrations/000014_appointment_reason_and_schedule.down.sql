ALTER TABLE appointments
    DROP CONSTRAINT IF EXISTS appointments_employee_schedule_excl,
    DROP CONSTRAINT IF EXISTS appointments_reason_check,
    DROP COLUMN IF EXISTS cancellation_reason,
    DROP COLUMN IF EXISTS reason;

DROP TRIGGER IF EXISTS appointments_hourly_capacity_check ON appointments;
DROP FUNCTION IF EXISTS enforce_appointment_hourly_capacity();
DROP FUNCTION IF EXISTS appointment_schedule_range(TIMESTAMP WITH TIME ZONE);
