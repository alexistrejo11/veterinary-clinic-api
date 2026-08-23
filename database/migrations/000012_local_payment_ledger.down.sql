-- PostgreSQL cannot drop enum values. The extra payment_status labels remain
-- after a down; 000013 reverses the table/constraint changes that use them.
SELECT 1;
