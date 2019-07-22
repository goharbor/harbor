package webhook

import (
	"github.com/goharbor/harbor/src/webhook/model"
)

// PushImagePreprocessHandler preprocess image push event data
type PushImagePreprocessHandler struct {
}

// Handle preprocess image push event data and then publish hook event
func (h *PushImagePreprocessHandler) Handle(value interface{}) error {
	hookType := model.EventTypePushImage

	if err := PreprocessAndSendImageHook(hookType, value); err != nil {
		return err
	}
	return nil
}

// IsStateful ...
func (h *PushImagePreprocessHandler) IsStateful() bool {
	return false
}
