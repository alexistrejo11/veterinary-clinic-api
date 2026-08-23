-- 000010_canonical_user_roles.down.sql
-- The manager/employee to receptionist/veterinarian data conversion is irreversible.
-- This rollback restores the previous constraint only.

ALTER TABLE users DROP CONSTRAINT IF EXISTS check_user_role;
ALTER TABLE users
    ADD CONSTRAINT check_user_role
    CHECK (role IN ('veterinarian', 'receptionist', 'manager', 'customer', 'admin', 'superadmin'));
