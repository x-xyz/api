package ptr

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type pointerSuite struct {
	suite.Suite
}

func (s *pointerSuite) TestPointer() {
	p1 := String(`abc123`)
	p2 := Int(123)
	p3 := Int32(4567)
	p4 := Int64(891011)
	p5 := Float32(306.247)
	p6 := Float64(689.777)
	p7 := Bool(true)

	s.Equal(*p1, `abc123`)
	s.Equal(*p2, int(123))
	s.Equal(*p3, int32(4567))
	s.Equal(*p4, int64(891011))
	s.Equal(*p5, float32(306.247))
	s.Equal(*p6, float64(689.777))
	s.Equal(*p7, true)
}

func TestReflectSuite(t *testing.T) {
	rs := new(pointerSuite)
	suite.Run(t, rs)
}
