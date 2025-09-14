package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationMiddleware 验证中间件
type ValidationMiddleware struct {
	validator *validator.Validate
}

// NewValidationMiddleware 创建验证中间件
func NewValidationMiddleware() *ValidationMiddleware {
	validate := validator.New()

	// 注册自定义验证规则
	validate.RegisterValidation("phone", validatePhone)
	validate.RegisterValidation("password_strength", validatePasswordStrength)

	return &ValidationMiddleware{
		validator: validate,
	}
}

// ValidateStruct 验证结构体
func (v *ValidationMiddleware) ValidateStruct(data interface{}) []string {
	var errors []string

	if err := v.validator.Struct(data); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, formatValidationError(err))
		}
	}

	return errors
}

// formatValidationError 格式化验证错误
func formatValidationError(err validator.FieldError) string {
	field := strings.ToLower(err.Field())

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s 字段是必需的", field)
	case "email":
		return fmt.Sprintf("%s 必须是有效的邮箱地址", field)
	case "min":
		return fmt.Sprintf("%s 长度不能少于 %s 个字符", field, err.Param())
	case "max":
		return fmt.Sprintf("%s 长度不能超过 %s 个字符", field, err.Param())
	case "phone":
		return fmt.Sprintf("%s 必须是有效的手机号码", field)
	case "password_strength":
		return fmt.Sprintf("%s 必须包含大写字母、小写字母和数字，且至少8个字符", field)
	case "uuid":
		return fmt.Sprintf("%s 必须是有效的UUID格式", field)
	case "oneof":
		return fmt.Sprintf("%s 必须是以下值之一：%s", field, err.Param())
	default:
		return fmt.Sprintf("%s 字段验证失败", field)
	}
}

// validatePhone 自定义手机号码验证
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if len(phone) != 11 {
		return false
	}
	// 简单的中国手机号码验证
	return strings.HasPrefix(phone, "1") && len(phone) == 11
}

// validatePasswordStrength 密码强度验证
func validatePasswordStrength(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// 至少8个字符
	if len(password) < 8 {
		return false
	}

	// 包含大写字母、小写字母、数字
	hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
	hasDigit := strings.ContainsAny(password, "0123456789")

	return hasUpper && hasLower && hasDigit
}

// ValidationErrorResponse 验证错误响应
// func Validationdto.ErrorResponse(errors []string) fiber.Map {
// 	return fiber.Map{
// 		"code":    400,
// 		"message": "验证失败",
// 		"errors":  errors,
// 	}
// }

// // ValidateJSON 中间件：验证JSON请求体
// func ValidateJSON(validator *ValidationMiddleware, target interface{}) fiber.Handler {
// 	return func(c *fiber.Ctx) error {
// 		if err := c.BodyParser(target); err != nil {
// 			return c.Status(400).JSON(fiber.Map{
// 				"code":    400,
// 				"message": "请求体格式错误",
// 				"error":   err.Error(),
// 			})
// 		}

// 		if errors := validator.ValidateStruct(target); len(errors) > 0 {
// 			return c.Status(400).JSON(Validationdto.ErrorResponse(errors))
// 		}

// 		// 将验证后的数据存储在上下文中
// 		c.Locals("validated_data", target)
// 		return c.Next()
// 	}
// }
