package lib

import (
	"fmt"
)

type Report struct {
	passed []string
	failed []string
}

// Passed case
func (r *Report) Passed(caseName string) {
	r.passed = append(r.passed, fmt.Sprintf("%s: [%s]", caseName, "PASSED"))
}

// Failed case
func (r *Report) Failed(caseName string, err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	r.failed = append(r.failed, fmt.Sprintf("%s: [%s] %s", caseName, "FAILED", errMsg))
}

// Print report
func (r *Report) Print() {
	passed := len(r.passed)
	failed := len(r.failed)
	total := passed + failed

	fmt.Println("=====================================")
	fmt.Printf("Overall: %d/%d passed , %d/%d failed\n", passed, total, failed, total)
	fmt.Println("=====================================")
	for _, res := range r.passed {
		fmt.Println(res)
	}

	for _, res := range r.failed {
		fmt.Println(res)
	}
}

// IsFail : Overall result
func (r *Report) IsFail() bool {
	return len(r.failed) > 0
}
