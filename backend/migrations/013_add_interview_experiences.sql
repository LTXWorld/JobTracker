-- 新增面试体验记录表，支持记录各面试阶段的评分与备注
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name = 'interview_experiences'
    ) THEN
        CREATE TABLE interview_experiences (
            id BIGSERIAL PRIMARY KEY,
            application_id INTEGER NOT NULL REFERENCES job_applications(id) ON DELETE CASCADE,
            from_status application_status NOT NULL,
            to_status application_status NOT NULL,
            rating VARCHAR(16),
            note VARCHAR(200),
            skip BOOLEAN NOT NULL DEFAULT FALSE,
            skip_reason VARCHAR(200),
            recorded_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            recorded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
            created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
            CONSTRAINT chk_interview_experiences_rating CHECK (
                rating IS NULL OR rating IN ('good', 'average', 'bad')
            )
        );

        CREATE INDEX idx_interview_experiences_application_id ON interview_experiences(application_id, recorded_at DESC);
    END IF;
END $$;
