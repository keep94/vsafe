package filters_test

import (
	"testing"

	"github.com/keep94/vsafe"
	"github.com/keep94/vsafe/filters"
	"github.com/stretchr/testify/assert"
)

func TestEntryMapper(t *testing.T) {
	assert := assert.New(t)
	var entry vsafe.Entry
	mapperTrue := filters.EntryMapper(
		func(src, dest *vsafe.Entry) bool {
			return true
		})
	mapperFalse := filters.EntryMapper(
		func(src, dest *vsafe.Entry) bool {
			return false
		})
	assert.Nil(mapperFalse.Map(&entry))
	assert.NotNil(mapperTrue.Map(&entry))
	assert.Same(mapperTrue.Map(&entry), mapperTrue.Map(&entry))
	assert.NotSame(mapperTrue.Map(&entry), mapperTrue.Clone().Map(&entry))
}
