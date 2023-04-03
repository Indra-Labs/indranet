package engine

import (
	"fmt"
	"testing"
)

var Expected = []string{
	`
	Segments{
		Segment{ DStart: 0, DEnd: 192, PEnd: 256, SLen: 191, Last: 191},
		Segment{ DStart: 256, DEnd: 448, PEnd: 512, SLen: 191, Last: 191},
		Segment{ DStart: 512, DEnd: 704, PEnd: 768, SLen: 191, Last: 191},
		Segment{ DStart: 768, DEnd: 960, PEnd: 1024, SLen: 191, Last: 191},
		Segment{ DStart: 1024, DEnd: 1216, PEnd: 1280, SLen: 191, Last: 191},
		Segment{ DStart: 1280, DEnd: 1472, PEnd: 1536, SLen: 191, Last: 191},
		Segment{ DStart: 1536, DEnd: 1728, PEnd: 1792, SLen: 191, Last: 191},
		Segment{ DStart: 1792, DEnd: 1822, PEnd: 1832, SLen: 191, Last: 12},
	}
`,
	`
	Segments{
		Segment{ DStart: 0, DEnd: 131, PEnd: 131, SLen: 4031, Last: 258},
	}
`,
	`
	Segments{
		Segment{ DStart: 0, DEnd: 128, PEnd: 256, SLen: 4031, Last: 4031},
		Segment{ DStart: 256, DEnd: 259, PEnd: 262, SLen: 4031, Last: 258},
	}
`,
	`
	Segments{
		Segment{ DStart: 0, DEnd: 66, PEnd: 66, SLen: 4031, Last: 129},
	}
`,
}

func TestNewSegments(t *testing.T) {
	msgSize := 2<<17 + 111
	segSize := 256
	s := NewPacketSegments(msgSize, segSize, PacketBaseLen, 64)
	o := fmt.Sprint(s)
	if o != Expected[0] {
		t.Errorf(
			"Failed to correctly generate.\ngot:\n'%s'\nexpected:\n'%s'",
			o, Expected[0])
	}
	msgSize = 2 << 18
	segSize = 4096
	s = NewPacketSegments(msgSize, segSize, PacketBaseLen, 0)
	o = fmt.Sprint(s)
	if o != Expected[1] {
		t.Errorf(
			"Failed to correctly generate.\ngot:\n%s\nexpected:\n%s",
			o, Expected[1])
	}
	s = NewPacketSegments(msgSize, segSize, PacketBaseLen, 128)
	o = fmt.Sprint(s)
	if o != Expected[2] {
		t.Errorf(
			"Failed to correctly generate.\ngot:\n%s\nexpected:\n%s",
			o, Expected[2])
	}
	msgSize = 2 << 17
	segSize = 4096
	s = NewPacketSegments(msgSize, segSize, PacketBaseLen, 0)
	o = fmt.Sprint(s)
	if o != Expected[3] {
		t.Errorf(
			"Failed to correctly generate.\ngot:\n%s\nexpected:\n%s",
			o, Expected[3])
	}
}
