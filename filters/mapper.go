package filters

import (
	"github.com/keep94/consume"
	"github.com/keep94/vsafe"
)

type entryMapper struct {
	M    func(src, dest *vsafe.Entry) bool
	temp vsafe.Entry
}

func EntryMapper(m func(src, dest *vsafe.Entry) bool) consume.Mapper {
	return &entryMapper{M: m}
}

func (e *entryMapper) Map(ptr interface{}) interface{} {
	if e.M(ptr.(*vsafe.Entry), &e.temp) {
		return &e.temp
	}
	return nil
}

func (e *entryMapper) Clone() consume.Mapper {
	result := *e
	return &result
}
