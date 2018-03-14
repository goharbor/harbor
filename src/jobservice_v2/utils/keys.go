package utils

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func generateScore() int64 {
	ticks := time.Now().Unix()
	rand := rand.New(rand.NewSource(ticks))
	return ticks + rand.Int63n(1000) //Double confirm to avoid potential duplications
}

//MakePeriodicPolicyUUID returns an UUID for the periodic policy.
func MakePeriodicPolicyUUID() (string, int64) {
	score := generateScore()
	return MakePeriodicPolicyUUIDWithScore(score), score
}

//MakePeriodicPolicyUUIDWithScore returns the UUID based on the specified score for the periodic policy.
func MakePeriodicPolicyUUIDWithScore(score int64) string {
	rawUUID := fmt.Sprintf("%s:%s:%d", "periodic", "policy", score)
	return base64.StdEncoding.EncodeToString([]byte(rawUUID))
}

//ExtractScoreFromUUID extracts the score from the UUID.
func ExtractScoreFromUUID(UUID string) int64 {
	if IsEmptyStr(UUID) {
		return 0
	}

	rawData, err := base64.StdEncoding.DecodeString(UUID)
	if err != nil {
		return 0
	}

	data := string(rawData)
	fragments := strings.Split(data, ":")
	if len(fragments) != 3 {
		return 0
	}

	score, err := strconv.ParseInt(fragments[2], 10, 64)
	if err != nil {
		return 0
	}

	return score
}

//KeyNamespacePrefix returns the based key based on the namespace.
func KeyNamespacePrefix(namespace string) string {
	ns := strings.TrimSpace(namespace)
	if !strings.HasSuffix(ns, ":") {
		return fmt.Sprintf("%s:", ns)
	}

	return ns
}

//KeyPeriod returns the key of period
func KeyPeriod(namespace string) string {
	return fmt.Sprintf("%s%s", KeyNamespacePrefix(namespace), "period")
}

//KeyPeriodicPolicy return the key of periodic policies.
func KeyPeriodicPolicy(namespace string) string {
	return fmt.Sprintf("%s:%s", KeyPeriod(namespace), "policies")
}

//KeyPeriodicNotification returns the key of periodic pub/sub channel.
func KeyPeriodicNotification(namespace string) string {
	return fmt.Sprintf("%s:%s", KeyPeriodicPolicy(namespace), "notifications")
}

//KeyPeriodicLock returns the key of locker under period
func KeyPeriodicLock(namespace string) string {
	return fmt.Sprintf("%s:%s", KeyPeriod(namespace), "lock")
}
