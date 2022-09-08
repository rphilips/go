package manuscript

import (
	pchapter "brocade.be/pbladng/lib/chapter"
	ptools "brocade.be/pbladng/lib/tools"
	ptopic "brocade.be/pbladng/lib/topic"
)

func TestNonempty(c *pchapter.Chapter) (err error) {
	if len(c.Topics) == 0 {
		err = ptools.Error("chapter-topics", c.Start, "chapter contains no topics")
		return
	}
	return
}

func TestMass(c *pchapter.Chapter, t *ptopic.Topic) (err error) {
	return
}
