-- Migration: Add export tasks table for Excel export functionality
-- File: 010_add_export_tasks_table.sql
-- Description: Create export_tasks table to track Excel export task status and files
-- Author: PACT Backend Coder
-- Date: 2025-09-09

-- Create export_tasks table
CREATE TABLE IF NOT EXISTS export_tasks (
    id SERIAL PRIMARY KEY,
    task_id VARCHAR(100) UNIQUE NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    export_type VARCHAR(20) NOT NULL DEFAULT 'xlsx',
    filename VARCHAR(255),
    file_path VARCHAR(500),
    file_size BIGINT,
    total_records INTEGER,
    processed_records INTEGER DEFAULT 0,
    progress INTEGER DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
    filters JSONB,
    options JSONB,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    expires_at TIMESTAMP,
    CONSTRAINT valid_status CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled', 'expired')),
    CONSTRAINT valid_export_type CHECK (export_type IN ('xlsx', 'csv')),
    CONSTRAINT valid_progress_range CHECK (progress >= 0 AND progress <= 100),
    CONSTRAINT valid_records_count CHECK (
        (total_records IS NULL OR total_records >= 0) AND
        (processed_records >= 0) AND
        (total_records IS NULL OR processed_records <= total_records)
    )
);

-- Add indexes for performance
CREATE INDEX idx_export_tasks_user_id ON export_tasks(user_id);
CREATE INDEX idx_export_tasks_status ON export_tasks(status);
CREATE INDEX idx_export_tasks_created_at ON export_tasks(created_at);
CREATE INDEX idx_export_tasks_task_id ON export_tasks(task_id);
CREATE INDEX idx_export_tasks_expires_at ON export_tasks(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_export_tasks_user_status ON export_tasks(user_id, status);

-- Add comments for documentation
COMMENT ON TABLE export_tasks IS 'Tracks Excel export task status and metadata';
COMMENT ON COLUMN export_tasks.task_id IS 'Unique identifier for the export task';
COMMENT ON COLUMN export_tasks.user_id IS 'User who initiated the export';
COMMENT ON COLUMN export_tasks.status IS 'Current status: pending, processing, completed, failed, cancelled, expired';
COMMENT ON COLUMN export_tasks.export_type IS 'Export format: xlsx, csv';
COMMENT ON COLUMN export_tasks.filename IS 'Generated filename for download';
COMMENT ON COLUMN export_tasks.file_path IS 'Server file path for generated file';
COMMENT ON COLUMN export_tasks.file_size IS 'File size in bytes';
COMMENT ON COLUMN export_tasks.total_records IS 'Total number of records to export';
COMMENT ON COLUMN export_tasks.processed_records IS 'Number of records processed so far';
COMMENT ON COLUMN export_tasks.progress IS 'Progress percentage (0-100)';
COMMENT ON COLUMN export_tasks.filters IS 'JSON filters applied to the export';
COMMENT ON COLUMN export_tasks.options IS 'JSON export options and preferences';
COMMENT ON COLUMN export_tasks.error_message IS 'Error message if export failed';
COMMENT ON COLUMN export_tasks.expires_at IS 'When the generated file expires';

-- Create function to automatically update progress based on records
CREATE OR REPLACE FUNCTION update_export_progress()
RETURNS TRIGGER AS $$
BEGIN
    -- Auto-calculate progress if total_records is set
    IF NEW.total_records IS NOT NULL AND NEW.total_records > 0 THEN
        NEW.progress = (NEW.processed_records * 100) / NEW.total_records;
    END IF;
    
    -- Auto-set started_at when status changes to processing
    IF OLD.status != 'processing' AND NEW.status = 'processing' AND NEW.started_at IS NULL THEN
        NEW.started_at = CURRENT_TIMESTAMP;
    END IF;
    
    -- Auto-set completed_at when status changes to completed or failed
    IF OLD.status NOT IN ('completed', 'failed') AND NEW.status IN ('completed', 'failed') AND NEW.completed_at IS NULL THEN
        NEW.completed_at = CURRENT_TIMESTAMP;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for automatic progress updates
CREATE TRIGGER trigger_update_export_progress
    BEFORE UPDATE ON export_tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_export_progress();

-- Create function to clean up expired export tasks
CREATE OR REPLACE FUNCTION cleanup_expired_export_tasks()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- Delete expired completed tasks and their files
    DELETE FROM export_tasks 
    WHERE status = 'completed' 
      AND expires_at IS NOT NULL
      AND expires_at < CURRENT_TIMESTAMP;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    -- Also clean up failed tasks older than 7 days
    DELETE FROM export_tasks
    WHERE status = 'failed'
      AND created_at < CURRENT_TIMESTAMP - INTERVAL '7 days';
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Create view for export task statistics
CREATE OR REPLACE VIEW export_task_stats AS
SELECT 
    user_id,
    COUNT(*) as total_tasks,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_tasks,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_tasks,
    COUNT(CASE WHEN status = 'processing' THEN 1 END) as active_tasks,
    COALESCE(SUM(total_records), 0) as total_exported_records,
    COALESCE(SUM(file_size), 0) as total_file_size,
    MAX(created_at) as last_export_at,
    AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) as avg_processing_time_seconds
FROM export_tasks
WHERE status IN ('completed', 'failed', 'processing')
GROUP BY user_id;

COMMENT ON VIEW export_task_stats IS 'Statistics view for export tasks per user';

-- Insert initial test data (only in development)
-- This will be used for testing the export functionality
DO $$
BEGIN
    -- Check if we're in a development environment by looking for test users
    IF EXISTS (SELECT 1 FROM users WHERE username = 'testuser') THEN
        -- Insert a sample completed export task for testing
        INSERT INTO export_tasks (
            task_id, user_id, status, export_type, filename, 
            total_records, processed_records, progress,
            filters, options, created_at, started_at, completed_at, expires_at
        ) VALUES (
            'export_20250909_000000_user1',
            (SELECT id FROM users WHERE username = 'testuser' LIMIT 1),
            'completed',
            'xlsx',
            '求职投递记录_测试用户_20250909_000000.xlsx',
            50,
            50,
            100,
            '{"status": ["已投递", "面试中"]}',
            '{"includeStatistics": true, "filename": "测试导出"}',
            CURRENT_TIMESTAMP - INTERVAL '1 hour',
            CURRENT_TIMESTAMP - INTERVAL '55 minutes',
            CURRENT_TIMESTAMP - INTERVAL '50 minutes',
            CURRENT_TIMESTAMP + INTERVAL '23 hours'
        );
        
        RAISE NOTICE 'Sample export task data inserted for testing';
    END IF;
EXCEPTION 
    WHEN OTHERS THEN
        RAISE NOTICE 'Could not insert sample export task data: %', SQLERRM;
END $$;

-- Create indexes on JSONB columns for filter performance
CREATE INDEX IF NOT EXISTS idx_export_tasks_filters_gin ON export_tasks USING GIN (filters);
CREATE INDEX IF NOT EXISTS idx_export_tasks_options_gin ON export_tasks USING GIN (options);

-- Performance optimization: Create partial indexes
CREATE INDEX idx_export_tasks_active_tasks ON export_tasks(user_id, created_at) 
WHERE status IN ('pending', 'processing');

CREATE INDEX idx_export_tasks_completed_recent ON export_tasks(user_id, completed_at DESC) 
WHERE status = 'completed' AND completed_at > CURRENT_TIMESTAMP - INTERVAL '7 days';

-- Add constraint to ensure file_path is provided for completed tasks
ALTER TABLE export_tasks ADD CONSTRAINT check_completed_task_has_file
CHECK (status != 'completed' OR file_path IS NOT NULL);

RAISE NOTICE 'Export tasks table and related objects created successfully';
RAISE NOTICE 'Added indexes for performance optimization';
RAISE NOTICE 'Created automatic cleanup function: cleanup_expired_export_tasks()';
RAISE NOTICE 'Created statistics view: export_task_stats';