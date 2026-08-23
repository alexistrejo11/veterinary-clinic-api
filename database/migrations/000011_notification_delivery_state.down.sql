DROP INDEX IF EXISTS idx_notifications_status;
DROP INDEX IF EXISTS idx_notifications_sender_id;
ALTER TABLE notifications
    DROP CONSTRAINT IF EXISTS chk_notification_status,
    DROP COLUMN IF EXISTS attempts,
    DROP COLUMN IF EXISTS sent_at,
    DROP COLUMN IF EXISTS provider_reference,
    DROP COLUMN IF EXISTS failure_message,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS sender_id;
