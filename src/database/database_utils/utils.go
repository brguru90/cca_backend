package database_utils

import "strings"

func GetDBErrorString(ins_err error) (err_str string, known_err bool) {
	errStr := ins_err.Error()
	resp_err := "Unrecognized error"
	switch {
	case strings.Contains(strings.ToLower(errStr), "index must have unique name"):
	case strings.Contains(strings.ToLower(errStr), "duplicate"):
		resp_err = "Already exists"
	default:
		return resp_err, false
	}
	return resp_err, true
}
