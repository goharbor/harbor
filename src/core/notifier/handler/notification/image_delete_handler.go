package notification

import (
	"github.com/goharbor/harbor/src/notification/model"
)

// DeleteImagePreprocessHandler preprocess image delete event data
type DeleteImagePreprocessHandler struct {
}

// Handle preprocess image delete event data and then publish hook event
func (h *DeleteImagePreprocessHandler) Handle(value interface{}) error {
	eventType := model.EventTypeDeleteImage

	if err := preprocessAndSendImageHook(eventType, value); err != nil {
		return err
	}
	return nil
}

// IsStateful ...
func (h *DeleteImagePreprocessHandler) IsStateful() bool {
	return false
}
