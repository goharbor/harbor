package notification

import (
	"github.com/goharbor/harbor/src/pkg/notification/model"
)

// PushImagePreprocessHandler preprocess image push event data
type PushImagePreprocessHandler struct {
}

// Handle preprocess image push event data and then publish hook event
func (h *PushImagePreprocessHandler) Handle(value interface{}) error {
	eventType := model.EventTypePushImage

	if err := preprocessAndSendImageHook(eventType, value); err != nil {
		return err
	}
	return nil
}

// IsStateful ...
func (h *PushImagePreprocessHandler) IsStateful() bool {
	return false
}
