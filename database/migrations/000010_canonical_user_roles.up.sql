-- 000010_canonical_user_roles.up.sql
-- Replace legacy authorization roles with the four canonical application roles.

UPDATE users
SET role = CASE role
    WHEN 'manager' THEN 'receptionist'
    WHEN 'employee' THEN 'veterinarian'
    ELSE role
END
WHERE role IN ('manager', 'employee');

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM users
        WHERE role NOT IN ('admin', 'veterinarian', 'receptionist', 'customer')
    ) THEN
        RAISE EXCEPTION
            'cannot migrate users with unsupported roles; resolve them before applying 000010_canonical_user_roles';
    END IF;
END
$$;

ALTER TABLE users DROP CONSTRAINT IF EXISTS check_user_role;
ALTER TABLE users
    ADD CONSTRAINT check_user_role
    CHECK (role IN ('admin', 'veterinarian', 'receptionist', 'customer'));
