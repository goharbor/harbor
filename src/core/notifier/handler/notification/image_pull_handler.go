package notification

import (
	"github.com/goharbor/harbor/src/pkg/notification/model"
)

// PullImagePreprocessHandler preprocess image pull event data
type PullImagePreprocessHandler struct {
}

// Handle preprocess image pull event data and then publish hook event
func (h *PullImagePreprocessHandler) Handle(value interface{}) error {
	eventType := model.EventTypePullImage

	if err := preprocessAndSendImageHook(eventType, value); err != nil {
		return err
	}
	return nil
}

// IsStateful ...
func (h *PullImagePreprocessHandler) IsStateful() bool {
	return false
}
