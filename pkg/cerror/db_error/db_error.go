package db_error

import "strings"

func IsUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Duplicate entry")
}

func IsRecordNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "record not found")
}
