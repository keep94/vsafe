package filters_test

import (
	"testing"

	"github.com/keep94/vsafe"
	"github.com/keep94/vsafe/filters"
	"github.com/stretchr/testify/assert"
)

func TestEntryFilterer(t *testing.T) {
	assert := assert.New(t)
	var entry vsafe.Entry
	filtererTrue := filters.EntryFilterer(
		func(ptr *vsafe.Entry) bool {
			return true
		})
	filtererFalse := filters.EntryFilterer(
		func(ptr *vsafe.Entry) bool {
			return false
		})
	assert.True(filtererTrue.Filter(&entry))
	assert.False(filtererFalse.Filter(&entry))
}
