package webhook

import (
	"github.com/goharbor/harbor/src/webhook/model"
)

// PullImagePreprocessHandler preprocess image pull event data
type PullImagePreprocessHandler struct {
}

// Handle preprocess image pull event data and then publish hook event
func (h *PullImagePreprocessHandler) Handle(value interface{}) error {
	hookType := model.EventTypePullImage

	if err := PreprocessAndSendImageHook(hookType, value); err != nil {
		return err
	}
	return nil
}

// IsStateful ...
func (h *PullImagePreprocessHandler) IsStateful() bool {
	return false
}
