package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type stackTestSuite struct {
	suite.Suite
}

func (c *stackTestSuite) SetupTest() {}

func (c *stackTestSuite) TestFrame() {
	stack := newStack()
	frames := stack.frames()
	c.Equal(len(frames), 4)
	fmt.Println(frames.format())
}

func (c *stackTestSuite) TestFormat() {
	stack := newStack()
	frames := stack.frames()
	c.Contains(frames[len(frames)-1].Function, "testing.tRunner")
}

func TestStackTestSuite(t *testing.T) {
	suite.Run(t, &stackTestSuite{})
}
