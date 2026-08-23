ALTER TABLE appointments
    ADD COLUMN reason VARCHAR(50),
    ADD COLUMN cancellation_reason TEXT;

ALTER TABLE appointments
    ADD CONSTRAINT appointments_reason_check CHECK (
        reason IS NULL OR reason IN (
            'routine_checkup', 'vaccination', 'injury', 'illness', 'dental',
            'grooming', 'behavior', 'nutrition', 'follow_up', 'emergency'
        )
    );

CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Appointments have a fixed 30-minute duration. Marking this wrapper immutable
-- makes the range usable in the GiST exclusion index.
CREATE FUNCTION appointment_schedule_range(scheduled_at TIMESTAMP WITH TIME ZONE)
RETURNS tstzrange
LANGUAGE SQL
IMMUTABLE
PARALLEL SAFE
AS $$
    SELECT tstzrange(scheduled_at, scheduled_at + INTERVAL '30 minutes', '[)');
$$;

CREATE FUNCTION enforce_appointment_hourly_capacity()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    hour_start TIMESTAMP WITH TIME ZONE;
    active_appointments INTEGER;
BEGIN
    IF NEW.deleted_at IS NOT NULL
        OR NEW.status IN ('cancelled', 'not_presented') THEN
        RETURN NEW;
    END IF;

    hour_start := date_trunc('hour', NEW.scheduled_date);
    -- Serialise writes for the hour so concurrent requests cannot overbook it.
    PERFORM pg_advisory_xact_lock(EXTRACT(EPOCH FROM hour_start)::BIGINT);

    SELECT COUNT(*) INTO active_appointments
    FROM appointments
    WHERE id IS DISTINCT FROM NEW.id
        AND scheduled_date >= hour_start
        AND scheduled_date < hour_start + INTERVAL '1 hour'
        AND deleted_at IS NULL
        AND status NOT IN ('cancelled', 'not_presented');

    IF active_appointments >= 5 THEN
        RAISE EXCEPTION 'maximum appointment capacity reached for hour %', hour_start
            USING ERRCODE = 'check_violation';
    END IF;

    RETURN NEW;
END;
$$;

ALTER TABLE appointments
    ADD CONSTRAINT appointments_employee_schedule_excl EXCLUDE USING gist (
        employee_id WITH =,
        appointment_schedule_range(scheduled_date) WITH &&
    ) WHERE (
        deleted_at IS NULL
        AND employee_id IS NOT NULL
        AND status NOT IN ('cancelled', 'not_presented')
    );

CREATE TRIGGER appointments_hourly_capacity_check
BEFORE INSERT OR UPDATE OF scheduled_date, status, deleted_at ON appointments
FOR EACH ROW EXECUTE FUNCTION enforce_appointment_hourly_capacity();
