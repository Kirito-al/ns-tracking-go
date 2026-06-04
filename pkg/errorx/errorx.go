package errorx

// Code 业务错误码定义
type Code int

const (
	Success Code = 200

	// 数据库错误（50001-50010）
	DatabaseError       Code = 50001
	DatabaseConnection  Code = 50002
	DatabaseQueryFailed Code = 50003
	UpsertFailed        Code = 50004

	// 业务逻辑错误（50011-50020）
	FormatError         Code = 50011
	InvalidStatus       Code = 50012
	TrackingNotFound    Code = 50013

	// Webhook 签名错误（40101-40110）
	SignatureInvalid    Code = 40101
	SignatureExpired    Code = 40102
	SignatureMissing    Code = 40103

	// 参数验证错误（40001-40010）
	InvalidParameter    Code = 40001
	MissingParameter    Code = 40002
	InvalidFormat       Code = 40003
)

// GetMessage 获取错误码对应的错误消息
func GetMessage(code Code) string {
	switch code {
	case Success:
		return "success"
	case DatabaseError:
		return "Database error"
	case DatabaseConnection:
		return "Database connection failed"
	case DatabaseQueryFailed:
		return "Database query failed"
	case UpsertFailed:
		return "Upsert operation failed"
	case FormatError:
		return "Format error"
	case InvalidStatus:
		return "Invalid tracking status"
	case TrackingNotFound:
		return "Tracking not found"
	case SignatureInvalid:
		return "Signature verification failed"
	case SignatureExpired:
		return "Signature expired"
	case SignatureMissing:
		return "Signature missing"
	case InvalidParameter:
		return "Invalid parameter"
	case MissingParameter:
		return "Missing required parameter"
	case InvalidFormat:
		return "Invalid format"
	default:
		return "Unknown error"
	}
}

// NewError 创建业务错误
func NewError(code Code) *CodeError {
	return &CodeError{
		Code:    int(code),
		Message: GetMessage(code),
	}
}

// CodeError 业务错误结构
type CodeError struct {
	Code    int
	Message string
}

// Error 实现error接口
func (e *CodeError) Error() string {
	return e.Message
}

// GetCode 获取错误码
func (e *CodeError) GetCode() int {
	return e.Code
}