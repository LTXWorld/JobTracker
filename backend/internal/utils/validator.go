// /Users/lutao/GolandProjects/jobView/backend/internal/utils/validator.go
// 输入验证工具类，提供各种数据验证和清理功能
// 防止XSS、SQL注入等安全漏洞，确保数据完整性和安全性

package utils

import (
	"fmt"
	"html"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

var (
	// 常用正则表达式
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,50}$`)
	phoneRegex    = regexp.MustCompile(`^1[3-9]\d{9}$`)
	sqlKeywords   = []string{"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER", "EXEC", "UNION", "SCRIPT"}
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors 多个验证错误
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// ValidateUsername 验证用户名
func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)
	
	if username == "" {
		return ValidationError{Field: "username", Message: "用户名不能为空"}
	}
	
	if len(username) < 3 {
		return ValidationError{Field: "username", Message: "用户名长度不能少于3位"}
	}
	
	if len(username) > 50 {
		return ValidationError{Field: "username", Message: "用户名长度不能超过50位"}
	}
	
	if !usernameRegex.MatchString(username) {
		return ValidationError{Field: "username", Message: "用户名只能包含字母、数字、下划线和短横线"}
	}
	
	// 检查敏感词
	if containsSQLKeywords(username) {
		return ValidationError{Field: "username", Message: "用户名包含非法字符"}
	}
	
	return nil
}

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	
	if email == "" {
		return ValidationError{Field: "email", Message: "邮箱不能为空"}
	}
	
	if len(email) > 254 {
		return ValidationError{Field: "email", Message: "邮箱长度不能超过254位"}
	}
	
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ValidationError{Field: "email", Message: "邮箱格式不正确"}
	}
	
	// 检查常见的临时邮箱域名
	tempDomains := []string{"10minutemail.com", "guerrillamail.com", "mailinator.com"}
	for _, domain := range tempDomains {
		if strings.HasSuffix(email, "@"+domain) {
			return ValidationError{Field: "email", Message: "不支持临时邮箱"}
		}
	}
	
	return nil
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ValidationError{Field: "password", Message: "密码长度不能少于8位"}
	}
	
	if len(password) > 128 {
		return ValidationError{Field: "password", Message: "密码长度不能超过128位"}
	}
	
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false
	
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}
	
	var errors []string
	if !hasUpper {
		errors = append(errors, "至少包含一个大写字母")
	}
	if !hasLower {
		errors = append(errors, "至少包含一个小写字母")
	}
	if !hasDigit {
		errors = append(errors, "至少包含一个数字")
	}
	if !hasSpecial {
		errors = append(errors, "至少包含一个特殊字符")
	}
	
	if len(errors) > 0 {
		return ValidationError{
			Field:   "password",
			Message: "密码必须" + strings.Join(errors, "、"),
		}
	}
	
	return nil
}

// ValidatePhone 验证手机号码
func ValidatePhone(phone string) error {
	if phone == "" {
		return nil // 手机号可选
	}
	
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	
	if !phoneRegex.MatchString(phone) {
		return ValidationError{Field: "phone", Message: "手机号码格式不正确"}
	}
	
	return nil
}

// ValidateCompanyName 验证公司名称
func ValidateCompanyName(companyName string) error {
	companyName = strings.TrimSpace(companyName)
	
	if companyName == "" {
		return ValidationError{Field: "company_name", Message: "公司名称不能为空"}
	}
	
	if len(companyName) > 255 {
		return ValidationError{Field: "company_name", Message: "公司名称长度不能超过255位"}
	}
	
	if containsSQLKeywords(companyName) {
		return ValidationError{Field: "company_name", Message: "公司名称包含非法字符"}
	}
	
	return nil
}

// ValidatePositionTitle 验证职位名称
func ValidatePositionTitle(positionTitle string) error {
	positionTitle = strings.TrimSpace(positionTitle)
	
	if positionTitle == "" {
		return ValidationError{Field: "position_title", Message: "职位名称不能为空"}
	}
	
	if len(positionTitle) > 255 {
		return ValidationError{Field: "position_title", Message: "职位名称长度不能超过255位"}
	}
	
	if containsSQLKeywords(positionTitle) {
		return ValidationError{Field: "position_title", Message: "职位名称包含非法字符"}
	}
	
	return nil
}

// ValidateDate 验证日期格式（YYYY-MM-DD）
func ValidateDate(dateStr string) error {
	if dateStr == "" {
		return nil // 日期可选
	}
	
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return ValidationError{Field: "date", Message: "日期格式不正确，请使用YYYY-MM-DD格式"}
	}
	
	return nil
}

// ValidateSalaryRange 验证薪资范围
func ValidateSalaryRange(salaryRange string) error {
	if salaryRange == "" {
		return nil // 薪资范围可选
	}
	
	if len(salaryRange) > 100 {
		return ValidationError{Field: "salary_range", Message: "薪资范围长度不能超过100位"}
	}
	
	return nil
}

// SanitizeInput 清理输入，防止XSS攻击
func SanitizeInput(input string) string {
	// HTML转义
	input = html.EscapeString(input)
	
	// 移除多余的空白字符
	input = strings.TrimSpace(input)
	
	// 移除NULL字符
	input = strings.ReplaceAll(input, "\x00", "")
	
	return input
}

// SanitizeHTML 清理HTML内容，保留安全的标签
func SanitizeHTML(input string) string {
	// 简单的HTML清理，移除script、iframe等危险标签
	dangerousTags := []string{
		"<script", "</script>",
		"<iframe", "</iframe>",
		"<object", "</object>",
		"<embed", "</embed>",
		"<form", "</form>",
		"<input", "</input>",
		"javascript:",
		"vbscript:",
		"onload=", "onclick=", "onerror=",
	}
	
	input = strings.ToLower(input)
	for _, tag := range dangerousTags {
		input = strings.ReplaceAll(input, tag, "")
	}
	
	return input
}

// ValidateLength 验证字符串长度
func ValidateLength(field, value string, min, max int) error {
	length := len(strings.TrimSpace(value))
	
	if length < min {
		return ValidationError{
			Field:   field,
			Message: fmt.Sprintf("长度不能少于%d位", min),
		}
	}
	
	if length > max {
		return ValidationError{
			Field:   field,
			Message: fmt.Sprintf("长度不能超过%d位", max),
		}
	}
	
	return nil
}

// ValidateRequired 验证必填字段
func ValidateRequired(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return ValidationError{
			Field:   field,
			Message: "这是必填字段",
		}
	}
	return nil
}

// ValidateOptionalText 验证可选文本字段
func ValidateOptionalText(field, value string, maxLength int) error {
	if value == "" {
		return nil
	}
	
	if len(value) > maxLength {
		return ValidationError{
			Field:   field,
			Message: fmt.Sprintf("长度不能超过%d位", maxLength),
		}
	}
	
	if containsSQLKeywords(value) {
		return ValidationError{
			Field:   field,
			Message: "内容包含非法字符",
		}
	}
	
	return nil
}

// containsSQLKeywords 检查是否包含SQL关键词
func containsSQLKeywords(input string) bool {
	upperInput := strings.ToUpper(input)
	for _, keyword := range sqlKeywords {
		if strings.Contains(upperInput, keyword) {
			return true
		}
	}
	return false
}

// ValidateStruct 结构体验证（简化版）
func ValidateStruct(v interface{}) ValidationErrors {
	var errors ValidationErrors
	
	// 这里可以使用反射来验证结构体字段
	// 为了简化，这里返回空的错误列表
	// 实际项目中建议使用 go-playground/validator 库
	
	return errors
}

// IsValidURL 验证URL格式
func IsValidURL(urlStr string) bool {
	if urlStr == "" {
		return true // 空URL被认为是有效的（可选字段）
	}
	
	// 简单的URL验证
	return strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://")
}

// ValidateWorkLocation 验证工作地点
func ValidateWorkLocation(location string) error {
	return ValidateOptionalText("work_location", location, 255)
}

// ValidateNotes 验证备注信息
func ValidateNotes(notes string) error {
	return ValidateOptionalText("notes", notes, 2000)
}

// ValidateContactInfo 验证联系信息
func ValidateContactInfo(contactInfo string) error {
	return ValidateOptionalText("contact_info", contactInfo, 500)
}