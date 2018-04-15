package mimemagic

import "bytes"
//import "fmt"

type section struct {
	Name string
	matchers []matcher
}
type matcher struct {
	Level, Offset, Range int
	Value []byte
	Mask  []byte
}

func Match(guess string, dat []byte) string {
	if s := smap[guess]; s!=nil && matchSection(s,dat) { return s.Name }
	for i:=0; i<len(sarr); i++ {
		if matchSection(&sarr[i], dat) { return sarr[i].Name }
	}
	return ""
}

func matchSection(s *section, b []byte) bool {
	lvl   := 0
	state := false
	for i:=0; i<len(s.matchers); i++ {
		m := &s.matchers[i]

		if m.Level <= lvl && state { return true }
		if m.Level > lvl && !state { continue }

//		fmt.Printf("%-30s %#v state=%v\n",s.Name,m,state)

		state = false
		lvl = m.Level
		if len(b) < m.Offset + len(m.Value) { continue }
		max := m.Offset + len(m.Value) + m.Range
		if max > len(b) { max = len(b) }
		if bytes.Index(b[m.Offset:max], m.Value) >= 0 {
			state = true
		}
	}
	return state
}
