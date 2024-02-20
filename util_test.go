package easytpl

import (
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func TestGetExtendsTplName(t *testing.T) {
	td := TplDelims{Left: "{{", Right: "}}"}

	s, ok := getExtendsTplName([]byte(`{{ extends "parent.tpl" }}`), td)
	assert.True(t, ok)
	assert.Eq(t, "parent.tpl", s)

	s, ok = getExtendsTplName([]byte(`{{ extends 'parent.tpl' }}`), td)
	assert.True(t, ok)
	assert.Eq(t, "parent.tpl", s)

	s, ok = getExtendsTplName([]byte(`{{ extends parent.tpl }}`), td)
	assert.True(t, ok)
	assert.Eq(t, "parent.tpl", s)
}
