package repository

import (
	"database/sql"
	"strings"
)

// isDuplicateError MySQLの重複エラーかチェック
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	// MySQL Error 1062: Duplicate entry
	errStr := err.Error()
	return strings.Contains(errStr, "Error 1062") || 
		   strings.Contains(errStr, "Duplicate entry") ||
		   strings.Contains(errStr, "duplicate key value")
}

// isNotFoundError レコードが見つからないエラーかチェック
func isNotFoundError(err error) bool {
	return err == sql.ErrNoRows
}