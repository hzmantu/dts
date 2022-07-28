package structs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var filepath = "../config/task.yaml"

func TestGetSqlOpenAddress(t *testing.T) {
	config := GetConfig(filepath)

	assert.IsType(t, config, &Config{})
}
