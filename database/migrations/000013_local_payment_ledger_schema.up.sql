-- Phase 7: payments are an internal clinic ledger. No gateway state is stored.
-- Relies on enum values added in 000012 (must be a separate transaction).

ALTER TABLE payments DROP CONSTRAINT IF EXISTS chk_paid_date;
ALTER TABLE payments DROP CONSTRAINT IF EXISTS chk_refund_date;
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_transaction_id_key;
ALTER TABLE payments RENAME COLUMN transaction_id TO reference;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS appointment_id INT REFERENCES appointments(id) ON DELETE SET NULL;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS reason TEXT;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS receipt_number VARCHAR(80);

UPDATE payments SET status = 'paid' WHERE status = 'completed';
UPDATE payments SET status = 'authorized' WHERE status = 'pending';
ALTER TABLE payments ALTER COLUMN status SET DEFAULT 'authorized';

-- Existing completed rows have no receipt_number; satisfy the new invariant.
UPDATE payments
SET receipt_number = COALESCE(receipt_number, 'LEGACY-' || lpad(id::text, 6, '0'))
WHERE status = 'paid' AND receipt_number IS NULL;

UPDATE payments
SET paid_at = COALESCE(paid_at, updated_at, created_at)
WHERE status = 'paid' AND paid_at IS NULL;

ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_reference_key;
ALTER TABLE payments ADD CONSTRAINT payments_reference_key UNIQUE (reference);
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_receipt_number_key;
ALTER TABLE payments ADD CONSTRAINT payments_receipt_number_key UNIQUE (receipt_number);
ALTER TABLE payments DROP CONSTRAINT IF EXISTS chk_payment_receipt;
ALTER TABLE payments ADD CONSTRAINT chk_payment_receipt CHECK (
    (status = 'paid' AND paid_at IS NOT NULL AND receipt_number IS NOT NULL)
    OR (status <> 'paid')
);
CREATE INDEX IF NOT EXISTS idx_payments_appointment_id ON payments(appointment_id);
