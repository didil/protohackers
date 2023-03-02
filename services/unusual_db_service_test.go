package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnusualDbService(t *testing.T) {
	svc := NewUnusualDbService()

	svc.Set("version", "koko")
	assert.Equal(t, "Ken's Key-Value Store 1.0", svc.Get("version"))

	svc.Set("my-key", "123")
	assert.Equal(t, "123", svc.Get("my-key"))

	svc.Set("my-key", "456")
	assert.Equal(t, "456", svc.Get("my-key"))

	svc.Set("my-other-key", "456=30")
	assert.Equal(t, "456=30", svc.Get("my-other-key"))
}
