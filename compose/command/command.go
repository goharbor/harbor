package command

type Command int32

const (
	CREATE_APP = 1 << iota
	STOP_APP
	SCALE_APP
	RESTART_APP
)
