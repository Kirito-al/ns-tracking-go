-- ========================================
-- 四张核心表结构（对标 Ruby ns-tracking 项目）
-- ========================================

-- 1. tracking_logs（原始真实轨迹表）
-- 用途：存储原始运单信息，用于GE渠道判断、Token路由
CREATE TABLE IF NOT EXISTS tracking_logs (
    id SERIAL PRIMARY KEY,
    source_tracking_number VARCHAR(255),      -- 主单号（云途运单号）
    tracking_number VARCHAR(255),              -- 内部系统单号
    client_order_number VARCHAR(255),          -- 客户订单号
    
    -- 渠道信息（关键：用于GE渠道判断）
    channel_alias VARCHAR(255),                -- 渠道别名（如：GE-云途）
    shipping_agent VARCHAR(255),               -- 物流代理商（如：云途）
    shipping_channel VARCHAR(255),             -- 物流渠道名称
    mabang_logistics_code VARCHAR(255),        -- 马帮物流代码
    
    -- 包裹信息
    package_number VARCHAR(255),               -- 集包号
    country_code VARCHAR(50),                  -- 目的国家代码（如：US）
    inner_number VARCHAR(255),                 -- 内部编号
    source_number VARCHAR(255),                -- 来源编号
    
    -- 状态信息
    status INTEGER DEFAULT 0,                  -- 0=packing,1=shipped,2=failed,3=intercept_pending,4=intercept_completed
    track_status INTEGER,                      -- 0=待揽收,1=运输中,2=已签收,3=异常
    
    -- 时间戳
    synced_at BIGINT,                          -- 最后同步时间（Unix timestamp）
    fulfill_at BIGINT,                         -- 发货时间（Unix timestamp）
    received_at BIGINT,                        -- 揽收时间（Unix timestamp）
    delivered_at BIGINT,                       -- 签收时间（Unix timestamp）
    tracked_at BIGINT,                         -- 最后轨迹时间（Unix timestamp）
    created_at BIGINT NOT NULL,                -- 创建时间（Unix timestamp）
    updated_at BIGINT NOT NULL                 -- 更新时间（Unix timestamp）
);

-- tracking_logs 索引
CREATE INDEX IF NOT EXISTS idx_source_tracking_number ON tracking_logs(source_tracking_number);
CREATE INDEX IF NOT EXISTS idx_tracking_number ON tracking_logs(tracking_number);
CREATE INDEX IF NOT EXISTS idx_client_order_number ON tracking_logs(client_order_number);
CREATE INDEX IF NOT EXISTS idx_package_number ON tracking_logs(package_number);
CREATE INDEX IF NOT EXISTS idx_track_status ON tracking_logs(track_status);
CREATE INDEX IF NOT EXISTS idx_synced_at ON tracking_logs(synced_at);

COMMENT ON TABLE tracking_logs IS '原始运单轨迹表（用于GE渠道判断、状态同步）';
COMMENT ON COLUMN tracking_logs.channel_alias IS '渠道别名（关键：用于GE-云途判断）';
COMMENT ON COLUMN tracking_logs.shipping_agent IS '物流代理商（关键：用于GE渠道判断）';
COMMENT ON COLUMN tracking_logs.shipping_channel IS '物流渠道（关键：用于GE渠道判断）';

-- 2. tracking_details（用户展示表）- 扩展字段
-- 注意：go-zero 使用 sqlx，字段名必须与 Go 结构体一致
CREATE TABLE IF NOT EXISTS tracking_details (
    id SERIAL PRIMARY KEY,
    tracking_number VARCHAR(255) NOT NULL UNIQUE,
    tracking_log_id INTEGER,                   -- 关联 tracking_logs.id
    tracking_replace_id INTEGER,               -- 关联 tracking_replaces.id（新增）
    
    service_class VARCHAR(255),                -- 服务类名（如：YunExpressService）
    status INTEGER,                            -- 状态码（0-1001）
    
    -- JSON 数据存储
    detail JSONB NOT NULL,                     -- 当前轨迹详情（response + metadata）
    last_detail JSONB,                         -- 上次轨迹详情（备份，用于回滚）
    replace_detail JSONB,                      -- 假轨迹替换数据（tracking_replace生成）
    
    -- 时间戳
    synced_at BIGINT,                          -- 同步时间（Unix timestamp）
    auto_delivered_at BIGINT,                  -- 自动签收时间（Unix timestamp）
    error_message TEXT,                        -- 错误信息
    
    created_at BIGINT NOT NULL,                -- 创建时间（Unix timestamp）
    updated_at BIGINT NOT NULL                 -- 更新时间（Unix timestamp）
);

-- tracking_details 索引
CREATE INDEX IF NOT EXISTS idx_tracking_number ON tracking_details(tracking_number);
CREATE INDEX IF NOT EXISTS idx_status ON tracking_details(status);
CREATE INDEX IF NOT EXISTS idx_synced_at ON tracking_details(synced_at);
CREATE INDEX IF NOT EXISTS idx_tracking_log_id ON tracking_details(tracking_log_id);
CREATE INDEX IF NOT EXISTS idx_tracking_replace_id ON tracking_details(tracking_replace_id);

COMMENT ON TABLE tracking_details IS '轨迹展示表（核心：用户查询数据源）';
COMMENT ON COLUMN tracking_details.last_detail IS '上次轨迹备份（用于数据回滚）';
COMMENT ON COLUMN tracking_details.replace_detail IS '假轨迹数据（tracking_replace生成）';

-- 3. tracking_cache_logs（同步缓冲表）
-- 用途：临时缓存待同步的运单，避免重复查询API
CREATE TABLE IF NOT EXISTS tracking_cache_logs (
    id SERIAL PRIMARY KEY,
    tracking_number VARCHAR(255) NOT NULL UNIQUE,
    service_class VARCHAR(255),                -- 服务类名
    
    -- 状态信息（镜像 tracking_detail）
    track_status INTEGER,                      -- 0=待揽收,1=运输中,2=已签收,3=异常
    delivered_at BIGINT,                       -- 签收时间（Unix timestamp）
    
    -- 同步控制
    synced_at BIGINT,                          -- 最后同步时间（Unix timestamp）
    retry_count INTEGER DEFAULT 0,             -- 重试次数
    
    created_at BIGINT NOT NULL,                -- 创建时间（Unix timestamp）
    updated_at BIGINT NOT NULL                 -- 更新时间（Unix timestamp）
);

-- tracking_cache_logs 索引
CREATE INDEX IF NOT EXISTS idx_tracking_number_cache ON tracking_cache_logs(tracking_number);
CREATE INDEX IF NOT EXISTS idx_track_status_cache ON tracking_cache_logs(track_status);
CREATE INDEX IF NOT EXISTS idx_synced_at_cache ON tracking_cache_logs(synced_at);

COMMENT ON TABLE tracking_cache_logs IS '轨迹同步缓冲表（避免重复API查询）';
COMMENT ON COLUMN tracking_cache_logs.retry_count IS '同步失败重试次数';

-- 4. tracking_replaces（数据回滚备份表）
-- 用途：生成假轨迹配置，用于缩短轨迹、数据回滚
CREATE TABLE IF NOT EXISTS tracking_replaces (
    id SERIAL PRIMARY KEY,
    
    -- 假轨迹生成参数（对标 Ruby tracking_replace.rb）
    first_start_at BIGINT,                     -- 第一个轨迹点开始时间（Unix timestamp）
    second_hour_from INTEGER,                  -- 第二个轨迹点小时范围起点
    second_hour_to INTEGER,                    -- 第二个轨迹点小时范围终点
    third_hour_from INTEGER,                   -- 第三个轨迹点小时范围起点
    third_hour_to INTEGER,                     -- 第三个轨迹点小时范围终点
    
    created_at BIGINT NOT NULL,                -- 创建时间（Unix timestamp）
    updated_at BIGINT NOT NULL                 -- 更新时间（Unix timestamp）
);

COMMENT ON TABLE tracking_replaces IS '假轨迹生成配置表（用于缩短轨迹、数据回滚）';
COMMENT ON COLUMN tracking_replaces.first_start_at IS '首轨迹点起始时间（Unix timestamp）';

-- 5. webhook_request_logs（Webhook请求日志表）
-- 用途：两阶段写入，记录Webhook请求的完整信息（替代raw_events的追溯性）
-- 技术方案参考：tms-talk/app/services/logistics/webhook_request_logs.py
CREATE TABLE IF NOT EXISTS webhook_request_logs (
    id SERIAL PRIMARY KEY,
    
    -- 基础信息
    provider_code VARCHAR(50),                 -- 服务商代码（如：yunexpress）
    tracking_number VARCHAR(255),               -- 运单号（可能多条轨迹）
    idempotency_key VARCHAR(64),                -- 幂等key（SHA256前16字节）
    
    -- 请求信息（两阶段写入：接收时记录）
    headers JSONB,                              -- 请求头（JSON格式）
    raw_body TEXT,                              -- 原始请求体（完整payload）
    client_ip VARCHAR(50),                      -- 客户端IP（云途推送服务器）
    
    -- 处理状态（两阶段写入：处理完成后更新）
    status VARCHAR(20) DEFAULT 'received',      -- 状态：received/processing/success/failed
    error TEXT,                                 -- 错误信息（失败时记录）
    response TEXT,                              -- 响应内容（成功时记录）
    
    -- 时间戳
    created_at BIGINT NOT NULL,                 -- 创建时间（接收请求时，Unix timestamp）
    updated_at BIGINT NOT NULL                  -- 更新时间（处理完成时，Unix timestamp）
);

-- webhook_request_logs 索引
CREATE INDEX IF NOT EXISTS idx_provider_code ON webhook_request_logs(provider_code);
CREATE INDEX IF NOT EXISTS idx_tracking_number_webhook ON webhook_request_logs(tracking_number);
CREATE INDEX IF NOT EXISTS idx_idempotency_key ON webhook_request_logs(idempotency_key);
CREATE INDEX IF NOT EXISTS idx_status_webhook ON webhook_request_logs(status);
CREATE INDEX IF NOT EXISTS idx_created_at_webhook ON webhook_request_logs(created_at);

COMMENT ON TABLE webhook_request_logs IS 'Webhook请求日志表（两阶段写入：接收时记录请求信息，处理完成后更新状态）';
COMMENT ON COLUMN webhook_request_logs.idempotency_key IS '幂等key（用于去重重复推送）';
COMMENT ON COLUMN webhook_request_logs.status IS '处理状态：received（接收）→ processing（处理中）→ success（成功）/ failed（失败）';
COMMENT ON COLUMN webhook_request_logs.raw_body IS '原始请求体（完整payload，用于问题排查）';

-- ========================================
-- 测试数据（模拟真实业务场景）
-- ========================================

-- 测试 tracking_logs（模拟GE渠道和普通渠道）
INSERT INTO tracking_logs (
    source_tracking_number,
    tracking_number,
    channel_alias,
    shipping_agent,
    shipping_channel,
    country_code,
    track_status,
    fulfill_at,
    created_at,
    updated_at
) VALUES 
-- GE渠道测试数据
(
    'YT2606500704802225',
    'YT2606500704802225',
    'GE-云途-标准',
    'GE-云途',
    'GE-YT-Standard',
    'US',
    1,
    EXTRACT(EPOCH FROM NOW() - INTERVAL '5 days'),
    EXTRACT(EPOCH FROM NOW()),
    EXTRACT(EPOCH FROM NOW())
),
-- 普通云途渠道测试数据
(
    'YT2606500704802226',
    'YT2606500704802226',
    '云途-标准',
    '云途',
    'YT-Standard',
    'US',
    1,
    EXTRACT(EPOCH FROM NOW() - INTERVAL '5 days'),
    EXTRACT(EPOCH FROM NOW()),
    EXTRACT(EPOCH FROM NOW())
)
ON CONFLICT DO NOTHING;

-- 测试 tracking_details（模拟完整轨迹）
INSERT INTO tracking_details (
    tracking_number,
    tracking_log_id,
    detail,
    status,
    service_class,
    synced_at,
    created_at,
    updated_at
) VALUES 
(
    'YT2606500704802225',
    (SELECT id FROM tracking_logs WHERE source_tracking_number = 'YT2606500704802225'),
    '{"response":{"Item":{"TrackingNumber":"USPS123456789","WayBillNumber":"YT2606500704802225","CarrierName":"云途","ProviderName":"云途物流","TrackingStatus":"20","PackageState":"2","OrderTrackingDetails":[{"ProcessDate":"2025-05-20T08:00:00Z","ProcessLocation":"深圳","ProcessContent":"快件电子信息已收到","TrackingStatus":"10"},{"ProcessDate":"2025-05-20T12:00:00Z","ProcessLocation":"广州","ProcessContent":"快件已发出","TrackingStatus":"20"},{"ProcessDate":"2025-05-21T08:00:00Z","ProcessLocation":"香港","ProcessContent":"到达转运中心","TrackingStatus":"20"}]}},"synced_at":"2025-05-21T10:00:00Z"}',
    20,
    'YunExpressService',
    EXTRACT(EPOCH FROM NOW()),
    EXTRACT(EPOCH FROM NOW()),
    EXTRACT(EPOCH FROM NOW())
),
(
    'YT2606500704802226',
    (SELECT id FROM tracking_logs WHERE source_tracking_number = 'YT2606500704802226'),
    '{"response":{"Item":{"TrackingNumber":"USPS987654321","WayBillNumber":"YT2606500704802226","CarrierName":"云途","ProviderName":"云途物流","TrackingStatus":"30","PackageState":"2","OrderTrackingDetails":[{"ProcessDate":"2025-05-20T08:00:00Z","ProcessLocation":"深圳","ProcessContent":"快件电子信息已收到","TrackingStatus":"10"},{"ProcessDate":"2025-05-20T12:00:00Z","ProcessLocation":"广州","ProcessContent":"快件已发出","TrackingStatus":"20"},{"ProcessDate":"2025-05-22T14:00:00Z","ProcessLocation":"US","ProcessContent":"到达目的地国家","TrackingStatus":"30"}]}},"synced_at":"2025-05-22T15:00:00Z"}',
    30,
    'YunExpressService',
    EXTRACT(EPOCH FROM NOW()),
    EXTRACT(EPOCH FROM NOW()),
    EXTRACT(EPOCH FROM NOW())
)
ON CONFLICT (tracking_number) DO UPDATE SET
    detail = EXCLUDED.detail,
    status = EXCLUDED.status,
    synced_at = EXCLUDED.synced_at,
    updated_at = EXCLUDED.updated_at;

-- 测试 tracking_cache_logs（模拟缓冲数据）
INSERT INTO tracking_cache_logs (
    tracking_number,
    service_class,
    track_status,
    synced_at,
    created_at,
    updated_at
) VALUES 
(
    'YT2606500704802225',
    'YunExpressService',
    1,
    EXTRACT(EPOCH FROM NOW()),
    EXTRACT(EPOCH FROM NOW()),
    EXTRACT(EPOCH FROM NOW())
)
ON CONFLICT DO NOTHING;

-- 查询测试数据验证
SELECT '=== tracking_logs ===' AS table_name;
SELECT source_tracking_number, channel_alias, shipping_agent, country_code, track_status 
FROM tracking_logs 
WHERE source_tracking_number IN ('YT2606500704802225', 'YT2606500704802226');

SELECT '=== tracking_details ===' AS table_name;
SELECT tracking_number, status, service_class, synced_at 
FROM tracking_details 
WHERE tracking_number IN ('YT2606500704802225', 'YT2606500704802226');

SELECT '=== tracking_cache_logs ===' AS table_name;
SELECT tracking_number, track_status, synced_at 
FROM tracking_cache_logs 
WHERE tracking_number = 'YT2606500704802225';