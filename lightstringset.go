package ffmpeg_cli_go

import (
	"maps"
	"slices"
)

type LightStringSet map[string]struct{}

func NewLightStringSet(items ...string) LightStringSet {
	s := LightStringSet{}
	for _, item := range items {
		s.Insert(item)
	}
	return s
}

func (s LightStringSet) Insert(item string) LightStringSet {
	s[item] = struct{}{}
	return s
}

func (s LightStringSet) Delete(item string) LightStringSet {
	delete(s, item)
	return s
}

func (s LightStringSet) Has(item string) bool {
	_, found := s[item]
	return found
}

func (s LightStringSet) List() []string {
	res := slices.Collect(maps.Keys(s))
	slices.Sort(res)
	return res
}
