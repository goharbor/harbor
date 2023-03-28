package job

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStatus_After(t *testing.T) {
	type args struct {
		another Status
	}
	tests := []struct {
		name string
		s    Status
		args args
		want bool
	}{
		{
			name: "PenndingStatus is not after RunningStatus",
			s:    PendingStatus,
			args: args{
				another: PendingStatus,
			},
			want: false,
		},
		{
			name: "RunningStatus is after PendingStatus",
			s:    RunningStatus,
			args: args{
				another: PendingStatus,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.s.After(tt.args.another), "After(%v)", tt.args.another)
		})
	}
}

func TestStatus_Before(t *testing.T) {
	type args struct {
		another Status
	}
	tests := []struct {
		name string
		s    Status
		args args
		want bool
	}{
		{
			name: "RunningStatus is not before PendingStatus",
			s:    RunningStatus,
			args: args{
				another: RunningStatus,
			},
			want: false,
		},
		{
			name: "PenndingStatus is before RunningStatus",
			s:    PendingStatus,
			args: args{
				another: RunningStatus,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.s.Before(tt.args.another), "Before(%v)", tt.args.another)
		})
	}
}

func TestStatus_Code(t *testing.T) {
	tests := []struct {
		name string
		s    Status
		want int
	}{
		{
			name: "PenndingStatus is 0",
			s:    PendingStatus,
			want: 0,
		},
		{
			name: "ScheduledStatus is 1",
			s:    ScheduledStatus,
			want: 1,
		},
		{
			name: "RunningStatus is 2",
			s:    RunningStatus,
			want: 2,
		},
		{
			name: "FinishedStatus is 3",
			s:    StoppedStatus,
			want: 3,
		},
		{
			name: "FinishedStatus is 3",
			s:    ErrorStatus,
			want: 3,
		},
		{
			name: "FinishedStatus is 3",
			s:    SuccessStatus,
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.s.Code(), "Code()")
		})
	}
}

func TestStatus_Compare(t *testing.T) {
	type args struct {
		another Status
	}
	tests := []struct {
		name string
		s    Status
		args args
		want int
	}{
		{
			name: "PenndingStatus is less than runningStatus",
			s:    PendingStatus,
			args: args{
				another: PendingStatus,
			},
			want: 0,
		},
		{
			name: "ScheduledStatus is less than RunningStatus",
			s:    ScheduledStatus,
			args: args{
				another: PendingStatus,
			},
			want: 1,
		},
		{
			name: "RunningStatus is less than FinishedStatus",
			s:    PendingStatus,
			args: args{
				another: RunningStatus,
			},
			want: -2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.s.Compare(tt.args.another), "Compare(%v)", tt.args.another)
		})
	}
}

func TestStatus_Equal(t *testing.T) {
	type args struct {
		another Status
	}
	tests := []struct {
		name string
		s    Status
		args args
		want bool
	}{
		{
			name: "PendingStatus is equal to PendingStatus",
			s:    PendingStatus,
			args: args{
				another: PendingStatus,
			},
			want: true,
		},
		{
			name: "ScheduledStatus is not equal to PendingStatus",
			s:    ScheduledStatus,
			args: args{
				another: PendingStatus,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.s.Equal(tt.args.another), "Equal(%v)", tt.args.another)
		})
	}
}

func TestStatus_Final(t *testing.T) {
	tests := []struct {
		name string
		s    Status
		want bool
	}{
		{
			name: "PendingStatus is not final",
			s:    PendingStatus,
			want: false,
		},
		{
			name: "ScheduledStatus is not final",
			s:    ScheduledStatus,
			want: false,
		},
		{
			name: "RunningStatus is not final",
			s:    RunningStatus,
			want: false,
		},
		{
			name: "StoppedStatus is final",
			s:    StoppedStatus,
			want: true,
		},
		{
			name: "ErrorStatus is final",
			s:    ErrorStatus,
			want: true,
		},
		{
			name: "SuccessStatus is final",
			s:    SuccessStatus,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.s.Final(), "Final()")
		})
	}
}

func TestStatus_String(t *testing.T) {
	tests := []struct {
		name string
		s    Status
		want string
	}{
		{
			name: "PendingStatus to string is pending",
			s:    PendingStatus,
			want: "Pending",
		},
		{
			name: "ScheduledStatus to string is scheduled",
			s:    ScheduledStatus,
			want: "Scheduled",
		},
		{
			name: "RunningStatus to string is running",
			s:    RunningStatus,
			want: "Running",
		},
		{
			name: "StoppedStatus to string is stopped",
			s:    StoppedStatus,
			want: "Stopped",
		},
		{
			name: "ErrorStatus to string is error",
			s:    ErrorStatus,
			want: "Error",
		},
		{
			name: "SuccessStatus to string is success",
			s:    SuccessStatus,
			want: "Success",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.s.String(), "String()")
		})
	}
}

func TestStatus_Validate(t *testing.T) {
	tests := []struct {
		name    string
		s       Status
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Error status checksum failure",
			s:    Status("error"),
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return false
			},
		},
		{
			name: "SuccessStatus check success",
			s:    SuccessStatus,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.s.Validate(), fmt.Sprintf("Validate()"))
		})
	}
}
