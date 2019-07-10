package dao

import "testing"

func TestJoinNumberConditions(t *testing.T) {
	type args struct {
		ids []int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "normal test", args: args{[]int{1, 2, 3}}, want: "1,2,3"},
		{name: "dummy test", args: args{[]int{}}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JoinNumberConditions(tt.args.ids); got != tt.want {
				t.Errorf("JoinNumberConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}
