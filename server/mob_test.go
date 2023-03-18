package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceWithBogusCoin(t *testing.T) {
	assert.Equal(t, "Hi alice, please send payment to 7YWHMfk9JZe0LM0g1ZauHuiSxhI", replaceWithBogusCoin("Hi alice, please send payment to 7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX"))
	assert.Equal(t, "Hi alice, please send payment to 7YWHMfk9JZe0LM0g1ZauHuiSxhI ok ?", replaceWithBogusCoin("Hi alice, please send payment to 7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX ok ?"))
	assert.Equal(t, "7YWHMfk9JZe0LM0g1ZauHuiSxhI ok ?", replaceWithBogusCoin("7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX ok ?"))
}
