// Copyright Project Harbor Authors. All rights reserved.

package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func generateScore() int64 {
	ticks := time.Now().Unix()
	rand := rand.New(rand.NewSource(ticks))
	return ticks + rand.Int63n(1000) // Double confirm to avoid potential duplications
}

// MakePeriodicPolicyUUID returns an UUID for the periodic policy.
func MakePeriodicPolicyUUID() (string, int64) {
	score := generateScore()
	return MakeIdentifier(), score
}

// KeyNamespacePrefix returns the based key based on the namespace.
func KeyNamespacePrefix(namespace string) string {
	ns := strings.TrimSpace(namespace)
	if !strings.HasSuffix(ns, ":") {
		return fmt.Sprintf("%s:", ns)
	}

	return ns
}

// KeyPeriod returns the key of period
func KeyPeriod(namespace string) string {
	return fmt.Sprintf("%s%s", KeyNamespacePrefix(namespace), "period")
}

// KeyPeriodicPolicy returns the key of periodic policies.
func KeyPeriodicPolicy(namespace string) string {
	return fmt.Sprintf("%s:%s", KeyPeriod(namespace), "policies")
}

// KeyPeriodicPolicyScore returns the key of policy key and score mapping.
func KeyPeriodicPolicyScore(namespace string) string {
	return fmt.Sprintf("%s:%s", KeyPeriod(namespace), "key_score")
}

// KeyPeriodicNotification returns the key of periodic pub/sub channel.
func KeyPeriodicNotification(namespace string) string {
	return fmt.Sprintf("%s:%s", KeyPeriodicPolicy(namespace), "notifications")
}

// KeyPeriodicLock returns the key of locker under period
func KeyPeriodicLock(namespace string) string {
	return fmt.Sprintf("%s:%s", KeyPeriod(namespace), "lock")
}

// KeyJobStats returns the key of job stats
func KeyJobStats(namespace string, jobID string) string {
	return fmt.Sprintf("%s%s:%s", KeyNamespacePrefix(namespace), "job_stats", jobID)
}

// KeyJobCtlCommands give the key for publishing ctl commands like 'stop' etc.
func KeyJobCtlCommands(namespace string, jobID string) string {
	return fmt.Sprintf("%s%s:%s", KeyNamespacePrefix(namespace), "ctl_commands", jobID)
}
