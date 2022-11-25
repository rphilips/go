package structure

type Duel struct {
	Home   *Team
	Remote *Team
	Match  []*Match
	Score  string
}

func (duel Duel) String() string {
	return duel.Home.Name + " vs. " + duel.Remote.Name

}
