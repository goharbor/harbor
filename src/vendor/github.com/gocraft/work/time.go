package work

import "time"

var nowMock int64

func nowEpochSeconds() int64 {
	if nowMock != 0 {
		return nowMock
	}
	return time.Now().Unix()
}

func setNowEpochSecondsMock(t int64) {
	nowMock = t
}

func resetNowEpochSecondsMock() {
	nowMock = 0
}

// convert epoch seconds to a time
func epochSecondsToTime(t int64) time.Time {
	return time.Time{}
}
