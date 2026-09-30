-- Phase 6B expands the gatepass lifecycle to distinguish physical departure,
-- return-in-progress, overdue returns and terminal completion.
ALTER TABLE gatepasses
    MODIFY COLUMN status ENUM(
        'draft',
        'submitted',
        'pending_approval',
        'approved',
        'rejected',
        'cancelled',
        'expired',
        'checked_out',
        'awaiting_return',
        'partially_returned',
        'return_overdue',
        'checked_in',
        'completed'
    ) NOT NULL DEFAULT 'pending_approval';
