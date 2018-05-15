package replication

const (
	//FilterItemKindProject : Kind of filter item is 'project'
	FilterItemKindProject = "project"
	//FilterItemKindRepository : Kind of filter item is 'repository'
	FilterItemKindRepository = "repository"
	//FilterItemKindTag : Kind of filter item is 'tag'
	FilterItemKindTag = "tag"
	//FilterItemKindLabel : Kind of filter item is 'label'
	FilterItemKindLabel = "label"

	//AdaptorKindHarbor : Kind of adaptor of Harbor
	AdaptorKindHarbor = "Harbor"

	//TriggerKindImmediate : Kind of trigger is 'Immediate'
	TriggerKindImmediate = "Immediate"
	//TriggerKindSchedule : Kind of trigger is 'Scheduled'
	TriggerKindSchedule = "Scheduled"
	//TriggerKindManual : Kind of trigger is 'Manual'
	TriggerKindManual = "Manual"

	//TriggerScheduleDaily : type of scheduling is 'Daily'
	TriggerScheduleDaily = "Daily"
	//TriggerScheduleWeekly : type of scheduling is 'Weekly'
	TriggerScheduleWeekly = "Weekly"
)
