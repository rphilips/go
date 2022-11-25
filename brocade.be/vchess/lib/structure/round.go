package structure

import (
	"strconv"
	"time"
)

type Round struct {
	Round int
	Date  *time.Time
	Duels []*Duel
}

func (round Round) String() string {
	return "R" + strconv.Itoa(round.Round)
}
