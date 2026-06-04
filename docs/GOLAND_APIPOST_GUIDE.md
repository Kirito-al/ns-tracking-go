========================================
GoLand + ApiPost 完整测试指南
========================================

【第一步】GoLand运行配置

1. 打开GoLand
2. 打开项目：D:\gohome\demo1-gozero\ns-tracking-go
3. 找到文件：app/main.go
4. 右键 → Edit Configurations

配置参数：
  - Name: ns-tracking-app
  - Run kind: File
  - Files: app/main.go
  - Program arguments: -f etc/app.yaml
  - Working directory: D:\gohome\demo1-gozero\ns-tracking-go\app
  - Environment: （可选）GOPATH=...

5. 点击Run按钮启动

预期输出：
  Starting server at 0.0.0.0:8082...

----------------------------------------

【第二步】启动中间件（必须先启动）

PowerShell：
  cd D:\gohome\demo1-gozero\ns-tracking-go
  docker-compose up -d postgres redis

验证：
  docker ps | Select-String "postgres|redis"

----------------------------------------

【第三步】初始化数据库（首次运行）

PowerShell：
  cd D:\gohome\demo1-gozero\ns-tracking-go
  docker exec -i tracking-postgres psql -U postgres -d ns_admin_webhook_development < database.sql

验证表创建：
  docker exec tracking-postgres psql -U postgres -d ns_admin_webhook_development -c "\dt"

----------------------------------------

【第四步】ApiPost测试配置

接口信息：
  方法：POST
  URL：http://localhost:8082/webhook/yunexpress/tracking/normal

Headers：
  Content-Type: application/json
  X-YunExpress-Signature: （开发环境跳过验签，可不传）

Body（完整示例）：

{
  "trackingNumber": "YT2606500704802225",
  "wayBillNumber": "YT2606500704802225",
  "trackingStatus": "20",
  "packageState": 2,
  "providerName": "云途物流",
  "countryCode": "US",
  "orderTrackingDetails": [
    {
      "processDate": "2026-06-04T10:00:00Z",
      "processLocation": "深圳仓库",
      "processContent": "快件电子信息已收到",
      "trackingStatus": "10"
    },
    {
      "processDate": "2026-06-04T12:30:00Z",
      "processLocation": "深圳口岸",
      "processContent": "快件已发出",
      "trackingStatus": "20"
    }
  ]
}

----------------------------------------

【第五步】查看测试结果

成功响应：
{
  "code": 200,
  "message": "success",
  "data": {
    "trackingNumber": "YT2606500704802225",
    "status": 20
  }
}

失败响应（数据库未启动）：
{
  "code": 500,
  "message": "Database error",
  "data": ""
}

----------------------------------------

【第六步】验证数据库写入

PowerShell：
  docker exec tracking-postgres psql -U postgres -d ns_admin_webhook_development -c "SELECT tracking_number, status, synced_at FROM tracking_details WHERE tracking_number='YT2606500704802225'"

预期结果：
  tracking_number      | status | synced_at
  YT2606500704802225   | 20     | 1780402275

----------------------------------------

【常见问题】

❌ 问题1：连接数据库失败
解决：确保docker-compose已启动postgres

❌ 问题2：端口8082被占用
解决：修改etc/app.yaml的Port为8083或其他端口

❌ 问题3：找不到tracking_details表
解决：执行database.sql初始化数据库

❌ 问题4：Redis连接失败
解决：docker-compose up -d redis

----------------------------------------

【新架构vs旧架构】

旧架构（service/tracking）：
  - 需启动API（8082）+ RPC（50051）两个服务
  - HTTP → gRPC → DB

新架构（DDD）：
  - 只启动app.exe一个服务
  - HTTP → Logic → DB（直接调用）
  - 性能更好（无gRPC网络开销）

========================================

✅ GoLand运行后，用ApiPost发送请求即可测试！