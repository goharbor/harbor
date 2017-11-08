package replication

const (
	//FilterItemKindProject : Kind of filter item is 'project'
	FilterItemKindProject = "project"
	//FilterItemKindRepository : Kind of filter item is 'repository'
	FilterItemKindRepository = "repository"
	//FilterItemKindTag : Kind of filter item is 'tag'
	FilterItemKindTag = "tag"

	//TriggerKindManually : kind of trigger is 'manully'
	TriggerKindManually = "manually"
	//TriggerKindSchedule : kind of trigger is 'schedule'
	TriggerKindSchedule = "schedule"
	//TriggerKindImmediately : kind of trigger is 'immediately'
	TriggerKindImmediately = "immediately"

	//AdaptorKindHarbor : Kind of adaptor of Harbor
	AdaptorKindHarbor = "Harbor"
)
