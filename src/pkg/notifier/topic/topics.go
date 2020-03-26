package topic

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/notifier"
	"github.com/goharbor/harbor/src/pkg/notifier/handler/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

// Subscribe topics
func init() {
	handlersMap := map[string][]notifier.NotificationHandler{
		model.WebhookTopic: {&notification.HTTPHandler{}},
		model.SlackTopic:   {&notification.SlackHandler{}},
	}

	for t, handlers := range handlersMap {
		for _, handler := range handlers {
			if err := notifier.Subscribe(t, handler); err != nil {
				log.Errorf("failed to subscribe topic %s: %v", t, err)
				continue
			}
			log.Debugf("topic %s is subscribed", t)
		}
	}
}
