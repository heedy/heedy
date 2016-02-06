/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/

package users

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

//
// func IsValidName(n string) bool {
// 	if strings.Contains(n, "/") ||
// 		strings.Contains(n, "\\") ||
// 		strings.Contains(n, " ") ||
// 		strings.Contains(n, "?") ||
// 		strings.Contains(n, "\t") ||
// 		strings.Contains(n, "\n") ||
// 		strings.Contains(n, "\r") ||
// 		strings.Contains(n, "#") ||
// 		len(n) == 0 {
// 		return false
// 	}
//
// 	return true
// }

type FakeSqlResult struct {
	LastInsertIdVal   int64
	LastInsertIdError error
	RowsAffectedVal   int64
	RowsAffectedError error
}

func (fsr FakeSqlResult) LastInsertId() (int64, error) {
	return fsr.LastInsertIdVal, fsr.LastInsertIdError
}
func (fsr FakeSqlResult) RowsAffected() (int64, error) {
	return fsr.RowsAffectedVal, fsr.RowsAffectedError
}

func TestGetDeleteError(t *testing.T) {
	var fakeResult FakeSqlResult

	// With no rows changed we return a "nothing to delete" error
	err := getDeleteError(fakeResult, nil)
	assert.Equal(t, ErrNothingToDelete, err)

	// Passing in an error returns the same error
	tmpError := errors.New("testing error")
	assert.Equal(t, tmpError, getDeleteError(fakeResult, tmpError))

	// This should be set if the operation isn't supported for the given db type
	fakeResult.RowsAffectedError = tmpError
	assert.Equal(t, tmpError, getDeleteError(fakeResult, nil))

	// Now test what we were expecting
	fakeResult.RowsAffectedError = nil
	fakeResult.RowsAffectedVal = 1

	assert.Equal(t, nil, getDeleteError(fakeResult, nil))
}
