-- 014_add_process_type.sql
-- 为 job_applications 表增加求职周期字段，支持秋招/春招/社招三种独立记录

BEGIN;

ALTER TABLE job_applications
    ADD COLUMN IF NOT EXISTS process_type VARCHAR(16);

UPDATE job_applications
SET process_type = COALESCE(process_type, '秋招');

ALTER TABLE job_applications
    ALTER COLUMN process_type SET NOT NULL,
    ALTER COLUMN process_type SET DEFAULT '秋招';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE table_name = 'job_applications'
          AND constraint_name = 'chk_job_applications_process_type'
    ) THEN
        ALTER TABLE job_applications
        ADD CONSTRAINT chk_job_applications_process_type
        CHECK (process_type IN ('秋招','春招','社招'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_job_applications_user_process_type
    ON job_applications(user_id, process_type);

COMMIT;
