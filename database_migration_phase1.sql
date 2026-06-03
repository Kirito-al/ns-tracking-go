-- ========================================
-- Phase 1 数据库优化迁移脚本
-- ========================================
-- 目标：
-- 1. detail 字段从 text 改成 JSONB（已创建则跳过）
-- 2. 删除冗余的 created_time 字段（与 created_at 重复）
-- 3. 删除冗余的 updated_time 字段（与 updated_at 重复）
-- ========================================

-- 设置数据库
\c ns_admin_webhook_development

-- ========== 1. tracking_details 表优化 ==========

-- 1.1 删除 created_time 字段（如果存在）
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'tracking_details' 
        AND column_name = 'created_time'
    ) THEN
        ALTER TABLE tracking_details DROP COLUMN created_time;
        RAISE NOTICE 'Dropped tracking_details.created_time';
    ELSE
        RAISE NOTICE 'tracking_details.created_time does not exist, skipping';
    END IF;
END $$;

-- 1.2 删除 updated_time 字段（如果存在）
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'tracking_details' 
        AND column_name = 'updated_time'
    ) THEN
        ALTER TABLE tracking_details DROP COLUMN updated_time;
        RAISE NOTICE 'Dropped tracking_details.updated_time';
    ELSE
        RAISE NOTICE 'tracking_details.updated_time does not exist, skipping';
    END IF;
END $$;

-- 1.3 确保 detail 字段是 JSONB 类型（已创建则跳过）
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'tracking_details' 
        AND column_name = 'detail'
        AND data_type != 'jsonb'
    ) THEN
        ALTER TABLE tracking_details 
        ALTER COLUMN detail TYPE jsonb USING detail::jsonb;
        RAISE NOTICE 'Converted tracking_details.detail to JSONB';
    ELSE
        RAISE NOTICE 'tracking_details.detail is already JSONB, skipping';
    END IF;
END $$;

-- 1.4 确保 last_detail 字段是 JSONB 类型
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'tracking_details' 
        AND column_name = 'last_detail'
        AND data_type != 'jsonb'
    ) THEN
        ALTER TABLE tracking_details 
        ALTER COLUMN last_detail TYPE jsonb USING last_detail::jsonb;
        RAISE NOTICE 'Converted tracking_details.last_detail to JSONB';
    ELSE
        RAISE NOTICE 'tracking_details.last_detail is already JSONB, skipping';
    END IF;
END $$;

-- 1.5 确保 replace_detail 字段是 JSONB 类型
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'tracking_details' 
        AND column_name = 'replace_detail'
        AND data_type != 'jsonb'
    ) THEN
        ALTER TABLE tracking_details 
        ALTER COLUMN replace_detail TYPE jsonb USING replace_detail::jsonb;
        RAISE NOTICE 'Converted tracking_details.replace_detail to JSONB';
    ELSE
        RAISE NOTICE 'tracking_details.replace_detail is already JSONB, skipping';
    END IF;
END $$;

-- ========== 2. tracking_cache_logs 表优化 ==========

-- 2.1 确保 detail 字段是 JSONB 类型
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'tracking_cache_logs' 
        AND column_name = 'detail'
        AND data_type != 'jsonb'
    ) THEN
        ALTER TABLE tracking_cache_logs 
        ALTER COLUMN detail TYPE jsonb USING detail::jsonb;
        RAISE NOTICE 'Converted tracking_cache_logs.detail to JSONB';
    ELSE
        RAISE NOTICE 'tracking_cache_logs.detail is already JSONB, skipping';
    END IF;
END $$;

-- ========== 3. tracking_replaces 表优化 ==========

-- 3.1 确保 old_detail 字段是 JSONB 类型
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'tracking_replaces' 
        AND column_name = 'old_detail'
        AND data_type != 'jsonb'
    ) THEN
        ALTER TABLE tracking_replaces 
        ALTER COLUMN old_detail TYPE jsonb USING old_detail::jsonb;
        RAISE NOTICE 'Converted tracking_replaces.old_detail to JSONB';
    ELSE
        RAISE NOTICE 'tracking_replaces.old_detail is already JSONB, skipping';
    END IF;
END $$;

-- 3.2 确保 new_detail 字段是 JSONB 类型
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'tracking_replaces' 
        AND column_name = 'new_detail'
        AND data_type != 'jsonb'
    ) THEN
        ALTER TABLE tracking_replaces 
        ALTER COLUMN new_detail TYPE jsonb USING new_detail::jsonb;
        RAISE NOTICE 'Converted tracking_replaces.new_detail to JSONB';
    ELSE
        RAISE NOTICE 'tracking_replaces.new_detail is already JSONB, skipping';
    END IF;
END $$;

-- ========== 迁移完成验证 ==========

SELECT 'Migration completed!' AS status;

-- 验证 tracking_details 表结构
SELECT 
    column_name, 
    data_type, 
    is_nullable
FROM information_schema.columns
WHERE table_name = 'tracking_details'
ORDER BY ordinal_position;

-- 验证 JSONB 字段
SELECT 
    column_name, 
    data_type
FROM information_schema.columns
WHERE table_name = 'tracking_details'
AND column_name IN ('detail', 'last_detail', 'replace_detail');
