package validators

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	
	// Common regex patterns
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	uuidRegex  = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

func init() {
	validate = validator.New()
	
	// Register custom validators
	validate.RegisterValidation("strong_password", validateStrongPassword)
	validate.RegisterValidation("no_xss", validateNoXSS)
	validate.RegisterValidation("safe_string", validateSafeString)
}

// GetValidator returns the validator instance
func GetValidator() *validator.Validate {
	return validate
}

// ValidateStruct validates a struct using the registered validator
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// GetValidationErrors formats validation errors into a readable format
func GetValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			tag := e.Tag()
			
			switch tag {
			case "required":
				errors[field] = fmt.Sprintf("%s is required", field)
			case "email":
				errors[field] = "Invalid email format"
			case "min":
				errors[field] = fmt.Sprintf("%s must be at least %s characters", field, e.Param())
			case "max":
				errors[field] = fmt.Sprintf("%s must be at most %s characters", field, e.Param())
			case "strong_password":
				errors[field] = "Password must be at least 8 characters with uppercase, lowercase, number, and special character"
			case "no_xss":
				errors[field] = fmt.Sprintf("%s contains potentially dangerous content", field)
			case "safe_string":
				errors[field] = fmt.Sprintf("%s contains invalid characters", field)
			default:
				errors[field] = fmt.Sprintf("%s is invalid", field)
			}
		}
	}
	
	return errors
}

// Custom validators

// validateStrongPassword checks for strong password requirements
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	
	if len(password) < 8 {
		return false
	}
	
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
	
	return hasUpper && hasLower && hasNumber && hasSpecial
}

// validateNoXSS checks for potential XSS patterns
func validateNoXSS(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	
	// Simple XSS detection patterns
	xssPatterns := []string{
		"<script",
		"javascript:",
		"on[a-z]+=",
		"<iframe",
		"<object",
		"<embed",
		"<form",
	}
	
	lowerValue := strings.ToLower(value)
	for _, pattern := range xssPatterns {
		if strings.Contains(lowerValue, pattern) {
			return false
		}
	}
	
	return true
}

// validateSafeString checks for safe string content (no SQL injection patterns)
func validateSafeString(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	
	// Simple SQL injection detection patterns
	sqlPatterns := []string{
		"'",
		"\"",
		";",
		"--",
		"/*",
		"*/",
		"drop table",
		"delete from",
		"insert into",
		"update set",
		"select from",
		"union select",
	}
	
	lowerValue := strings.ToLower(value)
	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerValue, pattern) {
			return false
		}
	}
	
	return true
}

// SanitizeInput removes potentially dangerous characters from input
func SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	
	// Trim whitespace
	input = strings.TrimSpace(input)
	
	// Remove control characters except tab, newline, and carriage return
	sanitized := strings.Map(func(r rune) rune {
		if r < 32 && r != 9 && r != 10 && r != 13 {
			return -1
		}
		return r
	}, input)
	
	return sanitized
}