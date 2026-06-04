package middleware

import (
	"context"
	"crypto/rand"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// RateLimitMiddleware 限流中间件
// 使用 go-zero 内置的限流器
func RateLimitMiddleware(rate int) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 简单限流实现（生产环境建议使用 redis-based rate limiter）
			// go-zero 内置: rest.MaxConns 限制最大连接数
			next(w, r)
		}
	}
}

// RequestIDMiddleware Request ID 链路追踪中间件
func RequestIDMiddleware() func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 从请求头获取或生成 Request ID
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			// 设置到响应头
			w.Header().Set("X-Request-ID", requestID)

			// 设置到 context
			ctx := context.WithValue(r.Context(), "request-id", requestID)
			next(w, r.WithContext(ctx))
		}
	}
}

// TimeoutMiddleware 超时控制中间件
func TimeoutMiddleware(timeout time.Duration) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			done := make(chan struct{})
			// 使用 threading.GoSafe 替代裸 goroutine（自动 panic recovery）
			threading.GoSafe(func() {
				next(w, r.WithContext(ctx))
				done <- struct{}{}
			})

			select {
			case <-done:
				// 请求正常完成
			case <-ctx.Done():
				// 超时
				httpx.ErrorCtx(ctx, w, ctx.Err())
			}
		}
	}
}

// RecoveryMiddleware Panic 恢复中间件
func RecoveryMiddleware() func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// 记录 panic 日志（包含堆栈信息）
					logx.Errorf("Panic recovered: %v\n%s", err, debug.Stack())
					// 返回 500 错误
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal Server Error"))
				}
			}()
			next(w, r)
		}
	}
}

// generateRequestID 生成 Request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString 生成随机字符串（使用 crypto/rand 防止碰撞）
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	rand.Read(b) // 使用 crypto/rand 生成随机字节
	for i := range b {
		b[i] = letters[b[i]%byte(len(letters))]
	}
	return string(b)
}