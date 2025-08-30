package validators

import "regexp"

// ValidatePersianPhoneNumber checks if a phone number is a valid Persian phone number.
func ValidatePersianPhoneNumber(phoneNumber string) bool {
	// Matches 09 followed by 9 digits.
	re := regexp.MustCompile(`^09[0-9]{9}$`)
	return re.MatchString(phoneNumber)
}
