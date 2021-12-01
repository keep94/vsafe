// Package filters contains filterers and mappers for the consume package
package filters

import (
	"github.com/keep94/consume"
	"github.com/keep94/vsafe"
)

type entryFilterer func(ptr *vsafe.Entry) bool

func EntryFilterer(f func(ptr *vsafe.Entry) bool) consume.Filterer {
	return entryFilterer(f)
}

func (e entryFilterer) Filter(ptr interface{}) bool {
	return e(ptr.(*vsafe.Entry))
}
