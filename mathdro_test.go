package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var me = newMathdroExporter()

func TestRecovered(t *testing.T) {
	result, err := me.GetMetrics()
	assert.Nil(t, err)
	assert.True(t, len(result) > 0, len(result))
}
