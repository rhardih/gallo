package helpers

import (
	"testing"

	"gotest.tools/assert"
)

func TestDataURI(t *testing.T) {
	placeholder := Placeholder{12, 34, "#abcdef"}

	actualURI, err := placeholder.DataURI()
	expectedURI := "data:image/svg+xml;base64,CiAgPHN2ZyB3aWR0aD0iMTIiIGhlaWdodD0iMzQiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyIgdmVyc2lvbj0iMS4xIj4KICAgIDxyZWN0IHg9IjAiIHk9IjAiIHdpZHRoPSIxMDAlIiBoZWlnaHQ9IjEwMCUiIGZpbGw9IiNhYmNkZWYiPjwvcmVjdD4KICA8L3N2Zz4KICA="

	assert.NilError(t, err)
	assert.Equal(t, actualURI, expectedURI)
}
