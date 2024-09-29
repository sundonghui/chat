package mode

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func Test_Mode_Debug(t *testing.T) {
	Set(Debug)
	assert.Equal(t, Get(), Debug)
	assert.True(t, IsDev())
	assert.Equal(t, gin.Mode(), gin.DebugMode)
}

func Test_Mode_Test(t *testing.T) {
	Set(Test)
	assert.Equal(t, Get(), Test)
	assert.True(t, IsDev())
	assert.Equal(t, gin.Mode(), gin.TestMode)
}

func Test_Mode_Release(t *testing.T) {
	Set(Release)
	assert.Equal(t, Get(), Release)
	assert.False(t, IsDev())
	assert.Equal(t, gin.Mode(), gin.ReleaseMode)
}

func Test_Mode_Unknown(t *testing.T) {
	assert.Panics(t, func() {
		Set("unknown")
	})
}
