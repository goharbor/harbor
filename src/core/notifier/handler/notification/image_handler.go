package notification

// ImagePreprocessHandler preprocess image event data
type ImagePreprocessHandler struct {
}

// Handle preprocess image event data and then publish hook event
func (h *ImagePreprocessHandler) Handle(value interface{}) error {
	if err := preprocessAndSendImageHook(value); err != nil {
		return err
	}
	return nil
}

// IsStateful ...
func (h *ImagePreprocessHandler) IsStateful() bool {
	return false
}
