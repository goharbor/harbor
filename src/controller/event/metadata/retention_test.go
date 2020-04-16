package metadata

import (
	"testing"

	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/stretchr/testify/suite"
)

type retentionEventTestSuite struct {
	suite.Suite
}

func (r *retentionEventTestSuite) TestResolveOfDeleteRepositoryEventMetadata() {
	e := &event.Event{}
	metadata := &RetentionMetaData{
		Total:    0,
		Retained: 0,
		Deleted:  nil,
		Status:   "",
		TaskID:   0,
	}
	err := metadata.Resolve(e)
	r.Require().Nil(err)
	r.Equal(event2.TopicTagRetention, e.Topic)
	r.Require().NotNil(e.Data)
	_, ok := e.Data.(*event2.RetentionEvent)
	r.Require().True(ok)
}

func TestRetentionEventTestSuite(t *testing.T) {
	suite.Run(t, &retentionEventTestSuite{})
}
