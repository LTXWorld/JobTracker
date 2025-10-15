-- 精简 Offer 阶段状态并对既有数据进行迁移
-- 1. 将“已拒绝”枚举值重命名为“已拒绝offer”
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_type t
        JOIN pg_enum e ON t.oid = e.enumtypid
        WHERE t.typname = 'application_status'
          AND e.enumlabel = '已拒绝'
    ) THEN
        ALTER TYPE application_status RENAME VALUE '已拒绝' TO '已拒绝offer';
    END IF;
END;
$$;

-- 2. 统一历史状态（job_status_history）
UPDATE job_status_history
SET new_status = 'HR面通过'::application_status
WHERE new_status IN ('待发offer', '已收到offer');

UPDATE job_status_history
SET old_status = 'HR面通过'::application_status
WHERE old_status IN ('待发offer', '已收到offer');

WITH latest_finish AS (
    SELECT DISTINCT ON (job_application_id, user_id)
        job_application_id,
        user_id,
        COALESCE(old_status, '已接受offer'::application_status) AS prev_status
    FROM job_status_history
    WHERE new_status = '流程结束'
    ORDER BY job_application_id, user_id, status_changed_at DESC, id DESC
)
UPDATE job_status_history jsh
SET new_status = CASE latest_finish.prev_status
        WHEN '已接受offer'::application_status THEN '已接受offer'::application_status
        WHEN '已拒绝offer'::application_status THEN '已拒绝offer'::application_status
        WHEN '待发offer'::application_status THEN 'HR面通过'::application_status
        WHEN '已收到offer'::application_status THEN 'HR面通过'::application_status
        WHEN 'HR面通过'::application_status THEN 'HR面通过'::application_status
        ELSE '已接受offer'::application_status
    END
FROM latest_finish
WHERE jsh.job_application_id = latest_finish.job_application_id
  AND jsh.user_id = latest_finish.user_id
  AND jsh.new_status = '流程结束'::application_status;

UPDATE job_status_history
SET new_status = '已接受offer'::application_status
WHERE new_status = '流程结束'::application_status;

UPDATE job_status_history
SET old_status = '已接受offer'::application_status
WHERE old_status = '流程结束'::application_status;

-- 3. 统一当前岗位状态（job_applications）
UPDATE job_applications
SET status = 'HR面通过'::application_status
WHERE status IN ('待发offer', '已收到offer');

WITH latest_finish AS (
    SELECT DISTINCT ON (ja.id)
        ja.id,
        COALESCE(jsh.old_status, '已接受offer'::application_status) AS prev_status
    FROM job_applications ja
    LEFT JOIN job_status_history jsh
      ON jsh.job_application_id = ja.id
     AND jsh.new_status = '流程结束'
    WHERE ja.status = '流程结束'
    ORDER BY ja.id, jsh.status_changed_at DESC, jsh.id DESC
)
UPDATE job_applications ja
SET status = CASE latest_finish.prev_status
        WHEN '已接受offer'::application_status THEN '已接受offer'::application_status
        WHEN '已拒绝offer'::application_status THEN '已拒绝offer'::application_status
        WHEN '待发offer'::application_status THEN 'HR面通过'::application_status
        WHEN '已收到offer'::application_status THEN 'HR面通过'::application_status
        WHEN 'HR面通过'::application_status THEN 'HR面通过'::application_status
        ELSE '已接受offer'::application_status
    END
FROM latest_finish
WHERE ja.id = latest_finish.id
  AND ja.status = '流程结束'::application_status;

UPDATE job_applications
SET status = '已接受offer'::application_status
WHERE status = '流程结束'::application_status;

-- 4. 清理 JSON 历史/统计缓存中的旧状态
UPDATE job_applications ja
SET status_history = translated.new_history
FROM (
    SELECT id,
           REPLACE(
               REPLACE(
                   REPLACE(
                       REPLACE(status_history::text, '待发offer', 'HR面通过'),
                       '已收到offer', 'HR面通过'
                   ),
                   '已拒绝"', '已拒绝offer"'
               ),
               '流程结束', '已接受offer'
           )::jsonb AS new_history
    FROM job_applications
    WHERE status_history::text LIKE '%待发offer%'
       OR status_history::text LIKE '%已收到offer%'
       OR status_history::text LIKE '%已拒绝"%'
       OR status_history::text LIKE '%流程结束%'
) AS translated
WHERE ja.id = translated.id;

UPDATE job_applications ja
SET status_duration_stats = translated.new_stats
FROM (
    SELECT id,
           REPLACE(
               REPLACE(
                   REPLACE(
                       REPLACE(status_duration_stats::text, '待发offer', 'HR面通过'),
                       '已收到offer', 'HR面通过'
                   ),
                   '已拒绝"', '已拒绝offer"'
               ),
               '流程结束', '已接受offer'
           )::jsonb AS new_stats
    FROM job_applications
    WHERE status_duration_stats::text LIKE '%待发offer%'
       OR status_duration_stats::text LIKE '%已收到offer%'
       OR status_duration_stats::text LIKE '%已拒绝"%'
       OR status_duration_stats::text LIKE '%流程结束%'
) AS translated
WHERE ja.id = translated.id;

-- 5. 更新状态流转模板与过渡表
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'status_transitions') THEN
        UPDATE status_transitions
        SET from_status = 'HR面通过'
        WHERE from_status IN ('待发offer', '已收到offer');

        UPDATE status_transitions
        SET to_status = 'HR面通过'
        WHERE to_status IN ('待发offer', '已收到offer');

        UPDATE status_transitions
        SET from_status = '已接受offer'
        WHERE from_status = '流程结束';

        UPDATE status_transitions
        SET to_status = '已接受offer'
        WHERE to_status = '流程结束';

        UPDATE status_transitions
        SET from_status = '已拒绝offer'
        WHERE from_status = '已拒绝';

        UPDATE status_transitions
        SET to_status = '已拒绝offer'
        WHERE to_status = '已拒绝';
    END IF;
END;
$$;

-- 全量刷新默认模板为新的基础配置
UPDATE status_flow_templates
SET flow_config = '{
        "transitions": {
            "已投递": ["简历筛选中", "简历筛选未通过", "已拒绝offer"],
            "简历筛选中": ["笔试中", "简历筛选未通过"],
            "简历筛选未通过": [],
            "笔试中": ["笔试通过", "笔试未通过"],
            "笔试通过": ["一面中"],
            "笔试未通过": [],
            "一面中": ["一面通过", "一面未通过", "HR面中"],
            "一面通过": ["二面中", "三面中", "HR面中"],
            "一面未通过": [],
            "二面中": ["二面通过", "二面未通过"],
            "二面通过": ["三面中", "HR面中"],
            "二面未通过": [],
            "三面中": ["三面通过", "三面未通过"],
            "三面通过": ["HR面中"],
            "三面未通过": [],
            "HR面中": ["HR面通过", "HR面未通过"],
            "HR面通过": ["已接受offer", "已拒绝offer"],
            "HR面未通过": [],
            "已接受offer": [],
            "已拒绝offer": []
        },
        "rules": {
            "auto_transitions": {
                "笔试通过": "一面中",
                "一面通过": "二面中",
                "二面通过": "三面中",
                "三面通过": "HR面中"
            },
            "require_confirmation": ["已拒绝offer"],
            "time_limits": {
                "简历筛选中": 7,
                "笔试中": 3,
                "一面中": 1,
                "二面中": 1,
                "三面中": 1,
                "HR面中": 1
            }
        }
    }'::jsonb,
    updated_at = NOW()
WHERE is_default = TRUE;

-- 非默认模板进行关键字替换，避免遗留旧状态
UPDATE status_flow_templates
SET flow_config = REPLACE(
        REPLACE(
            REPLACE(
                REPLACE(flow_config::text, '待发offer', 'HR面通过'),
                '已收到offer', 'HR面通过'
            ),
            '已拒绝"', '已拒绝offer"'
        ),
        '流程结束', '已接受offer'
    )::jsonb,
    updated_at = NOW()
WHERE is_default = FALSE
  AND (
        flow_config::text LIKE '%待发offer%'
     OR flow_config::text LIKE '%已收到offer%'
     OR flow_config::text LIKE '%已拒绝"%'
     OR flow_config::text LIKE '%流程结束%'
  );

-- 6. 重建 application_status 枚举以移除旧值
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_type t WHERE t.typname = 'application_status_new'
    ) THEN
        DROP TYPE application_status_new;
    END IF;
END;
$$;

CREATE TYPE application_status_new AS ENUM (
    '已投递',
    '简历筛选中',
    '笔试中',
    '笔试通过',
    '笔试未通过',
    '一面中',
    '一面通过',
    '一面未通过',
    '二面中',
    '二面通过',
    '二面未通过',
    '三面中',
    '三面通过',
    '三面未通过',
    'HR面中',
    'HR面通过',
    'HR面未通过',
    '已接受offer',
    '已拒绝offer'
);

ALTER TABLE job_status_history
    ALTER COLUMN new_status DROP DEFAULT,
    ALTER COLUMN new_status TYPE application_status_new USING new_status::text::application_status_new;

ALTER TABLE job_status_history
    ALTER COLUMN old_status TYPE application_status_new USING (CASE WHEN old_status IS NULL THEN NULL ELSE old_status::text::application_status_new END);

ALTER TABLE job_applications
    ALTER COLUMN status DROP DEFAULT,
    ALTER COLUMN status TYPE application_status_new USING status::text::application_status_new;

ALTER TYPE application_status RENAME TO application_status_old;
ALTER TYPE application_status_new RENAME TO application_status;

ALTER TABLE job_applications
    ALTER COLUMN status SET DEFAULT '已投递'::application_status;

DROP TYPE application_status_old;
