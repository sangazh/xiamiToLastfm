package xiami

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadFile(t *testing.T) {
	data, err := ReadFile()
	assert.Nil(t, err)
	t.Log(data.Data[0])
}
