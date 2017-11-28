package replication

const (
	//FilterItemKindProject : Kind of filter item is 'project'
	FilterItemKindProject = "project"
	//FilterItemKindRepository : Kind of filter item is 'repository'
	FilterItemKindRepository = "repository"
	//FilterItemKindTag : Kind of filter item is 'tag'
	FilterItemKindTag = "tag"

	//AdaptorKindHarbor : Kind of adaptor of Harbor
	AdaptorKindHarbor = "Harbor"

	//TriggerKindImmediate : Kind of trigger is 'Immediate'
	TriggerKindImmediate = "immediate"
	//TriggerKindSchedule : Kind of trigger is 'Schedule'
	TriggerKindSchedule = "schedule"
	//TriggerKindManual : Kind of trigger is 'Manual'
	TriggerKindManual = "manual"

	//TriggerScheduleDaily : type of scheduling is 'daily'
	TriggerScheduleDaily = "daily"
	//TriggerScheduleWeekly : type of scheduling is 'weekly'
	TriggerScheduleWeekly = "weekly"
)
