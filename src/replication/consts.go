package replication

const (
	//FilterItemKindProject : Kind of filter item is 'project'
	FilterItemKindProject = "project"
	//FilterItemKindRepository : Kind of filter item is 'repository'
	FilterItemKindRepository = "repository"
	//FilterItemKindTag : Kind of filter item is 'tag'
	FilterItemKindTag = "tag"

	//TODO: Refactor constants

	//TriggerKindManually : kind of trigger is 'manully'
	TriggerKindManually = "manually"
	//TriggerKindImmediately : kind of trigger is 'immediately'
	TriggerKindImmediately = "immediately"

	//AdaptorKindHarbor : Kind of adaptor of Harbor
	AdaptorKindHarbor = "Harbor"

	//TriggerKindImmediate : Kind of trigger is 'Immediate'
	TriggerKindImmediate = "Immediate"
	//TriggerKindSchedule : Kind of trigger is 'Schedule'
	TriggerKindSchedule = "Schedule"
	//TriggerKindManual : Kind of trigger is 'Manual'
	TriggerKindManual = "Manual"
	//TriggerScheduleDaily : type of scheduling is 'daily'
	TriggerScheduleDaily = "daily"
	//TriggerScheduleWeekly : type of scheduling is 'weekly'
	TriggerScheduleWeekly = "weekly"
)
