package work

import (
	"math/rand"
)

type prioritySampler struct {
	sum     uint
	samples []sampleItem
}

type sampleItem struct {
	priority uint

	// payload:
	redisJobs               string
	redisJobsInProg         string
	redisJobsPaused         string
	redisJobsLock           string
	redisJobsLockInfo       string
	redisJobsMaxConcurrency string
}

func (s *prioritySampler) add(priority uint, redisJobs, redisJobsInProg, redisJobsPaused, redisJobsLock, redisJobsLockInfo, redisJobsMaxConcurrency string) {
	sample := sampleItem{
		priority:                priority,
		redisJobs:               redisJobs,
		redisJobsInProg:         redisJobsInProg,
		redisJobsPaused:         redisJobsPaused,
		redisJobsLock:           redisJobsLock,
		redisJobsLockInfo:       redisJobsLockInfo,
		redisJobsMaxConcurrency: redisJobsMaxConcurrency,
	}
	s.samples = append(s.samples, sample)
	s.sum += priority
}

// sample re-sorts s.samples, modifying it in-place. Higher weighted things will tend to go towards the beginning.
// NOTE: as written currently makes 0 allocations.
// NOTE2: this is an O(n^2 algorithm) that is:
//     5492ns for 50 jobs (50 is a large number of unique jobs in my experience)
//     54966ns for 200 jobs
//     ~1ms for 1000 jobs
//     ~4ms for 2000 jobs
func (s *prioritySampler) sample() []sampleItem {
	lenSamples := len(s.samples)
	remaining := lenSamples
	sumRemaining := s.sum
	lastValidIdx := 0

	// Algorithm is as follows:
	// Loop until we sort everything. We're going to sort it in-place, probabilistically moving the highest weights to the front of the slice.
	//   Pick a random number
	//   Move backwards through the slice on each iteration,
	//     and see where the random number fits in the continuum.
	//     If we find where it fits, sort the item to the next slot towards the front of the slice.
	for remaining > 1 {
		// rn from [0 to sumRemaining)
		rn := uint(rand.Uint32()) % sumRemaining

		prevSum := uint(0)
		for i := lenSamples - 1; i >= lastValidIdx; i-- {
			sample := s.samples[i]
			if rn < (sample.priority + prevSum) {
				// move the sample to the beginning
				s.samples[i], s.samples[lastValidIdx] = s.samples[lastValidIdx], s.samples[i]

				sumRemaining -= sample.priority
				break
			} else {
				prevSum += sample.priority
			}
		}

		lastValidIdx++
		remaining--
	}

	return s.samples
}
