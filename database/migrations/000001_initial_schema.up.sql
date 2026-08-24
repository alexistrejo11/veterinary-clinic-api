-- Complete development baseline. Later changes must be new migrations.

CREATE TYPE person_gender AS ENUM ('male', 'female', 'not_specified');
CREATE TYPE veterinarian_speciality AS ENUM (
    'unknown_specialty', 'general_practice', 'surgery', 'internal_medicine', 'dentistry',
    'dermatology', 'oncology', 'cardiology', 'neurology', 'ophthalmology', 'radiology',
    'emergency_critical_care', 'anesthesiology', 'pathology', 'preventive_medicine',
    'exotic_animal_medicine', 'equine_medicine', 'avian_medicine', 'zoo_animal_medicine',
    'food_animal_medicine', 'public_health'
);
CREATE TYPE currency AS ENUM ('USD', 'MXN', 'EUR', 'GBP', 'JPY', 'AUD', 'CAD', 'CHF', 'CNY', 'SEK', 'NZD');
CREATE TYPE payment_status AS ENUM ('pending', 'completed', 'failed', 'refunded', 'authorized', 'paid', 'cancelled', 'overdue');
CREATE TYPE payment_method AS ENUM ('cash', 'credit_card', 'debit_card', 'bank_transfer', 'paypal', 'stripe', 'check');
CREATE TYPE appointment_status AS ENUM ('pending', 'cancelled', 'confirmed', 'rescheduled', 'completed', 'not_presented');

CREATE TABLE users (
    id SERIAL PRIMARY KEY, email VARCHAR(255) UNIQUE, phone_number VARCHAR(255) UNIQUE,
    hashed_password VARCHAR(255), role VARCHAR(50) NOT NULL, status VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL, gender VARCHAR(50), photo_url TEXT DEFAULT 'https://example.com/default-profile.png',
    bio TEXT DEFAULT '', date_of_birth DATE, oauth_provider VARCHAR(50) NOT NULL DEFAULT 'local',
    oauth_provider_id VARCHAR(255), oauth_access_token TEXT, oauth_refresh_token TEXT,
    oauth_token_expiry TIMESTAMPTZ, email_verified BOOLEAN DEFAULT FALSE, two_fa_method VARCHAR(20) NOT NULL DEFAULT 'none',
    two_fa_secret VARCHAR(255), two_fa_enabled BOOLEAN DEFAULT FALSE, two_fa_enabled_at TIMESTAMPTZ,
    two_fa_backup_codes TEXT[], two_fa_backup_codes_generated_at TIMESTAMPTZ, last_2fa_code_used_at TIMESTAMPTZ,
    last_login TIMESTAMPTZ, login_attempts INTEGER DEFAULT 0, locked_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ, CONSTRAINT unique_oauth_provider_id UNIQUE (oauth_provider, oauth_provider_id),
    CONSTRAINT check_oauth_provider CHECK (oauth_provider IN ('google', 'facebook', 'apple', 'microsoft', 'github', 'local')),
    CONSTRAINT check_two_fa_method CHECK (two_fa_method IN ('totp', 'sms', 'email', 'none')),
    CONSTRAINT check_user_role CHECK (role IN ('admin', 'veterinarian', 'receptionist', 'customer')),
    CONSTRAINT check_user_status CHECK (status IN ('active', 'inactive', 'pending', 'banned', 'deleted'))
);
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_phone_number ON users(phone_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role ON users(role) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status ON users(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_oauth_provider ON users(oauth_provider, oauth_provider_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users(created_at DESC);
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_users_email_verified ON users(email_verified) WHERE email_verified = FALSE;

CREATE TABLE customers (
    id SERIAL PRIMARY KEY, first_name VARCHAR(255) NOT NULL, last_name VARCHAR(255) NOT NULL,
    photo VARCHAR(255) NOT NULL DEFAULT 'https://www.gravatar.com/avatar/', date_of_birth DATE NOT NULL,
    gender person_gender NOT NULL, user_id INT, is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, deleted_at TIMESTAMP
);
CREATE TABLE employees (
    id SERIAL PRIMARY KEY, first_name VARCHAR(255) NOT NULL, last_name VARCHAR(255) NOT NULL,
    gender person_gender NOT NULL, date_of_birth DATE NOT NULL, photo VARCHAR(500) NOT NULL,
    license_number VARCHAR(20) NOT NULL UNIQUE, speciality veterinarian_speciality NOT NULL,
    years_of_experience INT NOT NULL, is_active BOOLEAN NOT NULL DEFAULT TRUE, user_id INT, schedule_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE TABLE pets (
    id SERIAL PRIMARY KEY, name VARCHAR(255) NOT NULL, photo TEXT, species VARCHAR(100) NOT NULL, breed VARCHAR(100),
    age SMALLINT, gender VARCHAR(20) CHECK (gender IN ('male', 'female', 'unknown', 'other')), color VARCHAR(50),
    microchip VARCHAR(50) UNIQUE, blood_type VARCHAR(10), is_neutered BOOLEAN DEFAULT FALSE, customer_id INT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE, allergies TEXT, current_medications TEXT, special_needs TEXT,
    feeding_instructions TEXT, behavioral_notes TEXT, veterinary_contact TEXT, emergency_contact_name VARCHAR(255),
    emergency_contact_phone VARCHAR(20), created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at TIMESTAMPTZ,
    FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE
);
CREATE TABLE pet_feeding_instructions (
    id SERIAL PRIMARY KEY, pet_id INT NOT NULL REFERENCES pets(id) ON DELETE CASCADE, food_brand VARCHAR(255),
    food_type VARCHAR(100), amount_per_serving VARCHAR(100), frequency VARCHAR(100), special_instructions TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE pet_behavioral_notes (
    id SERIAL PRIMARY KEY, pet_id INT NOT NULL REFERENCES pets(id) ON DELETE CASCADE, note TEXT NOT NULL,
    noted_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, noted_by INT
);
CREATE TABLE pet_vaccinations (
    id SERIAL PRIMARY KEY, pet_id INT NOT NULL REFERENCES pets(id) ON DELETE CASCADE, vaccine_name VARCHAR(255) NOT NULL,
    administered_date DATE NOT NULL, next_due_date DATE, administered_by INT REFERENCES employees(id) ON DELETE SET NULL,
    batch_number VARCHAR(100), vaccine_type VARCHAR(100) NOT NULL DEFAULT 'core', notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE pet_deworming (
    id SERIAL PRIMARY KEY, pet_id INT NOT NULL REFERENCES pets(id) ON DELETE CASCADE, medication_name VARCHAR(255) NOT NULL,
    administered_date DATE NOT NULL, next_due_date DATE, administered_by INT REFERENCES employees(id) ON DELETE SET NULL,
    notes TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, deleted_at TIMESTAMPTZ
);
CREATE TABLE pet_chip_implants (
    id SERIAL PRIMARY KEY, pet_id INT NOT NULL REFERENCES pets(id) ON DELETE CASCADE, implant_date DATE NOT NULL,
    implant_location VARCHAR(100), chip_number VARCHAR(50) UNIQUE, implanted_by INT REFERENCES employees(id) ON DELETE SET NULL,
    notes TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE TABLE appointments (
    id SERIAL PRIMARY KEY,
    clinic_service VARCHAR(50) NOT NULL CHECK (clinic_service IN ('general_consultation', 'vaccination', 'surgery', 'dental_care', 'emergency_care', 'grooming', 'nutrition_consult', 'behavior_consult', 'wellness_exam', 'other')),
    scheduled_date TIMESTAMPTZ NOT NULL, status appointment_status NOT NULL, notes TEXT DEFAULT '', customer_id INT NOT NULL,
    pet_id INT NOT NULL, employee_id INT, created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at TIMESTAMPTZ, reason VARCHAR(50), cancellation_reason TEXT,
    FOREIGN KEY (pet_id) REFERENCES pets(id) ON DELETE CASCADE, FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    CONSTRAINT appointments_reason_check CHECK (reason IS NULL OR reason IN ('routine_checkup', 'vaccination', 'injury', 'illness', 'dental', 'grooming', 'behavior', 'nutrition', 'follow_up', 'emergency'))
);
CREATE FUNCTION appointment_schedule_range(scheduled_at TIMESTAMPTZ) RETURNS tstzrange LANGUAGE SQL IMMUTABLE PARALLEL SAFE AS $$
    SELECT tstzrange(scheduled_at, scheduled_at + INTERVAL '30 minutes', '[)');
$$;
CREATE FUNCTION enforce_appointment_hourly_capacity() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE hour_start TIMESTAMPTZ; active_appointments INTEGER;
BEGIN
    IF NEW.deleted_at IS NOT NULL OR NEW.status IN ('cancelled', 'not_presented') THEN RETURN NEW; END IF;
    hour_start := date_trunc('hour', NEW.scheduled_date);
    PERFORM pg_advisory_xact_lock(EXTRACT(EPOCH FROM hour_start)::BIGINT);
    SELECT COUNT(*) INTO active_appointments FROM appointments WHERE id IS DISTINCT FROM NEW.id
        AND scheduled_date >= hour_start AND scheduled_date < hour_start + INTERVAL '1 hour'
        AND deleted_at IS NULL AND status NOT IN ('cancelled', 'not_presented');
    IF active_appointments >= 5 THEN RAISE EXCEPTION 'maximum appointment capacity reached for hour %', hour_start USING ERRCODE = 'check_violation'; END IF;
    RETURN NEW;
END;
$$;
ALTER TABLE appointments ADD CONSTRAINT appointments_employee_schedule_excl EXCLUDE USING gist (
    employee_id WITH =, appointment_schedule_range(scheduled_date) WITH &&
) WHERE (deleted_at IS NULL AND employee_id IS NOT NULL AND status NOT IN ('cancelled', 'not_presented'));
CREATE TRIGGER appointments_hourly_capacity_check BEFORE INSERT OR UPDATE OF scheduled_date, status, deleted_at ON appointments
FOR EACH ROW EXECUTE FUNCTION enforce_appointment_hourly_capacity();

CREATE TABLE medical_sessions (
    id SERIAL PRIMARY KEY, pet_id INT NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    customer_id INT NOT NULL REFERENCES customers(id) ON DELETE CASCADE, employee_id INT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    appointment_id INT REFERENCES appointments(id) ON DELETE SET NULL,
    clinic_service VARCHAR(50) NOT NULL CHECK (clinic_service IN ('general_consultation', 'vaccination', 'surgery', 'dental_care', 'emergency_care', 'grooming', 'nutrition_consult', 'behavior_consult', 'wellness_exam', 'other')),
    visit_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, visit_type VARCHAR(50) NOT NULL DEFAULT '', diagnosis TEXT, notes TEXT,
    treatment TEXT, condition VARCHAR(254), weight DECIMAL(4,2), temperature DECIMAL(3,1), heart_rate INT, respiratory_rate INT,
    symptoms TEXT, medications TEXT, follow_up_date TIMESTAMPTZ, is_emergency BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at TIMESTAMPTZ
);
CREATE TABLE vaccine_catalog (id SERIAL PRIMARY KEY, name VARCHAR(120) NOT NULL, manufacturer VARCHAR(120), species VARCHAR(50), disease_target VARCHAR(120), total_doses INT DEFAULT 1, schedule_days INT[], notes TEXT, is_active BOOLEAN DEFAULT TRUE, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE TABLE medications (id SERIAL PRIMARY KEY, name VARCHAR(120) NOT NULL, active_ingredient VARCHAR(120), presentation VARCHAR(80), unit VARCHAR(30), requires_prescription BOOLEAN DEFAULT FALSE, species_warnings TEXT, is_active BOOLEAN DEFAULT TRUE, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE TABLE service_catalog (id SERIAL PRIMARY KEY, name VARCHAR(120) NOT NULL, category VARCHAR(50) NOT NULL CHECK (category IN ('consultation', 'vaccination', 'surgery', 'dental', 'grooming', 'laboratory', 'imaging', 'emergency', 'nutrition', 'behavior', 'wellness', 'other')), description TEXT, base_price NUMERIC(10,2), duration_min INT, requires_fasting BOOLEAN DEFAULT FALSE, is_active BOOLEAN DEFAULT TRUE, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE TABLE session_vaccinations (id SERIAL PRIMARY KEY, session_id INT NOT NULL REFERENCES medical_sessions(id) ON DELETE CASCADE, vaccine_catalog_id INT NOT NULL REFERENCES vaccine_catalog(id), batch_number VARCHAR(60), dose_number INT NOT NULL DEFAULT 1, expiration_date DATE, site_of_injection VARCHAR(60), next_dose_date DATE, reaction_notes TEXT, administered_by INT REFERENCES employees(id), created_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE TABLE session_surgeries (id SERIAL PRIMARY KEY, session_id INT NOT NULL REFERENCES medical_sessions(id) ON DELETE CASCADE, procedure_name VARCHAR(150) NOT NULL, anesthesia_type VARCHAR(80), anesthesia_agent VARCHAR(120), pre_op_notes TEXT, intra_op_notes TEXT, post_op_notes TEXT, duration_minutes INT, outcome VARCHAR(50) CHECK (outcome IN ('successful', 'complicated', 'aborted', 'pending')), surgeon_id INT REFERENCES employees(id), created_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE TABLE session_prescriptions (id SERIAL PRIMARY KEY, session_id INT NOT NULL REFERENCES medical_sessions(id) ON DELETE CASCADE, medication_id INT NOT NULL REFERENCES medications(id), dosage VARCHAR(80) NOT NULL, frequency VARCHAR(80) NOT NULL, duration_days INT, route VARCHAR(50), instructions TEXT, start_date DATE DEFAULT CURRENT_DATE, end_date DATE GENERATED ALWAYS AS (start_date + duration_days) STORED, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE TABLE session_attachments (id SERIAL PRIMARY KEY, session_id INT NOT NULL REFERENCES medical_sessions(id) ON DELETE CASCADE, file_type VARCHAR(40) NOT NULL CHECK (file_type IN ('image', 'xray', 'lab_result', 'ecg', 'pdf', 'other')), file_url TEXT NOT NULL, description VARCHAR(255), uploaded_by INT REFERENCES employees(id), created_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE TABLE session_services (id SERIAL PRIMARY KEY, session_id INT NOT NULL REFERENCES medical_sessions(id) ON DELETE CASCADE, service_catalog_id INT NOT NULL REFERENCES service_catalog(id), quantity NUMERIC(6,2) NOT NULL DEFAULT 1, price_applied NUMERIC(10,2), notes TEXT, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE VIEW pet_vaccination_history AS SELECT ms.pet_id, sv.id AS vaccination_id, ms.id AS session_id, ms.visit_date, vc.name AS vaccine_name, vc.disease_target, sv.dose_number, sv.batch_number, sv.next_dose_date, sv.reaction_notes, ms.employee_id AS vet_id FROM session_vaccinations sv JOIN medical_sessions ms ON ms.id = sv.session_id JOIN vaccine_catalog vc ON vc.id = sv.vaccine_catalog_id ORDER BY ms.pet_id, ms.visit_date;

CREATE TABLE payments (
    id SERIAL PRIMARY KEY, amount NUMERIC(10,2) NOT NULL CHECK (amount > 0), currency VARCHAR(10) NOT NULL DEFAULT 'MXN',
    status payment_status NOT NULL DEFAULT 'authorized', method payment_method NOT NULL DEFAULT 'cash', med_session_id INT REFERENCES medical_sessions(id) ON DELETE SET NULL,
    reference VARCHAR(255) UNIQUE, description TEXT, appointment_id INT REFERENCES appointments(id) ON DELETE SET NULL, reason TEXT,
    due_date TIMESTAMPTZ, paid_at TIMESTAMPTZ, refunded_at TIMESTAMPTZ, is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at TIMESTAMPTZ,
    paid_by_customer_id INT REFERENCES customers(id) ON DELETE SET NULL, invoice_id VARCHAR(100), refund_amount NUMERIC(10,2) CHECK (refund_amount >= 0),
    failure_reason TEXT, receipt_number VARCHAR(80) UNIQUE, CONSTRAINT chk_refund_amount CHECK (refund_amount IS NULL OR refund_amount <= amount),
    CONSTRAINT chk_payment_receipt CHECK ((status = 'paid' AND paid_at IS NOT NULL AND receipt_number IS NOT NULL) OR status <> 'paid')
);
CREATE INDEX idx_payments_status ON payments(status); CREATE INDEX idx_payments_customer_id ON payments(paid_by_customer_id);
CREATE INDEX idx_payments_method ON payments(method); CREATE INDEX idx_payments_created_at ON payments(created_at);
CREATE INDEX idx_payments_due_date ON payments(due_date); CREATE INDEX idx_payments_paid_at ON payments(paid_at);
CREATE INDEX idx_payments_invoice_id ON payments(invoice_id); CREATE INDEX idx_payments_appointment_id ON payments(appointment_id);
CREATE INDEX idx_payments_is_active ON payments(is_active) WHERE is_active = TRUE; CREATE INDEX idx_payments_deleted_at ON payments(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_payments_customer_status ON payments(paid_by_customer_id, status); CREATE INDEX idx_payments_status_date ON payments(status, created_at);
CREATE INDEX idx_payments_customer_date ON payments(paid_by_customer_id, created_at); CREATE INDEX idx_payments_due_status ON payments(due_date, status) WHERE status != 'completed';

CREATE TABLE addresses (
    id SERIAL PRIMARY KEY, user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE, street VARCHAR(255) NOT NULL, city VARCHAR(255) NOT NULL,
    state VARCHAR(255) NOT NULL, zip_code VARCHAR(255) NOT NULL, country VARCHAR(255) NOT NULL, building_type VARCHAR(255) NOT NULL,
    building_outer_number VARCHAR(255) NOT NULL, building_inner_number VARCHAR(255) NOT NULL, is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at TIMESTAMPTZ,
    version INT NOT NULL DEFAULT 1, CONSTRAINT chk_building_type CHECK (building_type IN ('house', 'apartment', 'office', 'other')),
    CONSTRAINT chk_country CHECK (country IN ('USA', 'Mexico', 'Canada'))
);
CREATE INDEX idx_addresses_user_id ON addresses(user_id); CREATE INDEX idx_addresses_street ON addresses(street); CREATE INDEX idx_addresses_city ON addresses(city); CREATE INDEX idx_addresses_state ON addresses(state); CREATE INDEX idx_addresses_zip_code ON addresses(zip_code); CREATE INDEX idx_addresses_country ON addresses(country); CREATE INDEX idx_addresses_building_type ON addresses(building_type); CREATE INDEX idx_addresses_building_outer_number ON addresses(building_outer_number); CREATE INDEX idx_addresses_building_inner_number ON addresses(building_inner_number); CREATE INDEX idx_addresses_is_default ON addresses(is_default);

CREATE TABLE notifications (
    id SERIAL PRIMARY KEY, user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE, sender_id INT REFERENCES users(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL, subject VARCHAR(255) NOT NULL, message TEXT NOT NULL, token VARCHAR(255) NOT NULL,
    notification_type VARCHAR(255) NOT NULL, channel VARCHAR(255) NOT NULL, status VARCHAR(16) NOT NULL DEFAULT 'pending', failure_message TEXT,
    provider_reference VARCHAR(255), sent_at TIMESTAMPTZ, attempts INT NOT NULL DEFAULT 0, created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_notification_type CHECK (notification_type IN ('alert', 'reminder', 'info', 'verification_token', 'activation_token', 'password_reset', 'appointment_confirm', 'appointment_remind', 'appointment_cancel', 'payment_receipt', 'payment_reminder', 'welcome', 'security', 'promotional', 'unknown')),
    CONSTRAINT chk_channel CHECK (channel IN ('email', 'sms', 'push', 'in_app', 'voice', 'whatsapp', 'unknown')),
    CONSTRAINT chk_notification_status CHECK (status IN ('pending', 'sent', 'failed'))
);
CREATE INDEX idx_notifications_user_id ON notifications(user_id); CREATE INDEX idx_notifications_sender_id ON notifications(sender_id);
CREATE INDEX idx_notifications_notification_type ON notifications(notification_type); CREATE INDEX idx_notifications_channel ON notifications(channel);
CREATE INDEX idx_notifications_status ON notifications(status); CREATE INDEX idx_notifications_created_at ON notifications(created_at); CREATE INDEX idx_notifications_updated_at ON notifications(updated_at);
