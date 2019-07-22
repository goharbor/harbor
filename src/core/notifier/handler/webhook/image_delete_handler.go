package webhook

import (
	"github.com/goharbor/harbor/src/webhook/model"
)

// DeleteImagePreprocessHandler preprocess image delete event data
type DeleteImagePreprocessHandler struct {
}

// Handle preprocess image delete event data and then publish hook event
func (h *DeleteImagePreprocessHandler) Handle(value interface{}) error {
	hookType := model.EventTypeDeleteImage

	if err := PreprocessAndSendImageHook(hookType, value); err != nil {
		return err
	}
	return nil
}

// IsStateful ...
func (h *DeleteImagePreprocessHandler) IsStateful() bool {
	return false
}
