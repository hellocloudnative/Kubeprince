package service

import (
	"lvsucc/internal/ipvs"
)

//BuildLvscare is
func BuildLvscare() Lvser {
	l := &lvscare{}
	handle := ipvs.New()
	l.handle = handle

	return l
}
