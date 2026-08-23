ALTER TABLE notifications
    ADD COLUMN sender_id INT REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN status VARCHAR(16) NOT NULL DEFAULT 'pending',
    ADD COLUMN failure_message TEXT,
    ADD COLUMN provider_reference VARCHAR(255),
    ADD COLUMN sent_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN attempts INT NOT NULL DEFAULT 0,
    ADD CONSTRAINT chk_notification_status CHECK (status IN ('pending', 'sent', 'failed'));

CREATE INDEX idx_notifications_sender_id ON notifications (sender_id);
CREATE INDEX idx_notifications_status ON notifications (status);
