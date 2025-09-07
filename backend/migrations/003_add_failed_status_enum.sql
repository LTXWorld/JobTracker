-- 添加面试失败状态枚举值
-- 创建时间: 2025-09-06
-- 目的: 为每个面试环节添加"未通过"状态，实现细粒度状态跟踪

-- 添加简历筛选未通过状态
ALTER TYPE application_status ADD VALUE '简历筛选未通过';

-- 添加笔试未通过状态  
ALTER TYPE application_status ADD VALUE '笔试未通过';

-- 添加一面未通过状态
ALTER TYPE application_status ADD VALUE '一面未通过';

-- 添加二面未通过状态
ALTER TYPE application_status ADD VALUE '二面未通过';

-- 添加三面未通过状态
ALTER TYPE application_status ADD VALUE '三面未通过';

-- 添加HR面未通过状态
ALTER TYPE application_status ADD VALUE 'HR面未通过';

-- 添加注释说明新状态的用途
COMMENT ON TYPE application_status IS '求职申请状态枚举，包含各阶段的进行中、通过、未通过状态';