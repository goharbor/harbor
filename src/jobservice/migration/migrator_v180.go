// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package migration

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/jobservice/period"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/gomodule/redigo/redis"
)

const (
	jobNameGarbageCollection    = "IMAGE_GC"
	jobNameScanAll              = "IMAGE_SCAN_ALL"
	jobNameReplicationScheduler = "IMAGE_REPLICATE"
)

// PolicyMigrator migrate the cron job policy to new schema
type PolicyMigrator struct {
	// namespace of rdb
	namespace string

	// Pool for connecting to redis
	pool *redis.Pool
}

// PolicyMigratorFactory is a factory func to create PolicyMigrator
func PolicyMigratorFactory(pool *redis.Pool, namespace string) (RDBMigrator, error) {
	if pool == nil {
		return nil, errors.New("PolicyMigratorFactory: missing pool")
	}

	if utils.IsEmptyStr(namespace) {
		return nil, errors.New("PolicyMigratorFactory: missing namespace")
	}

	return &PolicyMigrator{
		namespace: namespace,
		pool:      pool,
	}, nil
}

// Metadata returns the base information of this migrator
func (pm *PolicyMigrator) Metadata() *MigratorMeta {
	return &MigratorMeta{
		FromVersion: "<1.8.0",
		ToVersion:   "1.8.1",
		ObjectRef:   "{namespace}:period:policies",
	}
}

// Migrate data
func (pm *PolicyMigrator) Migrate() error {
	conn := pm.pool.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Errorf("close redis connection error: %s", err)
		}
	}()

	allJobIDs, err := getAllJobStatsIDs(conn, pm.namespace)
	if err != nil {
		return errors.Wrap(err, "get job stats list error")
	}

	args := []interface{}{
		"id_placeholder",
		"id",
		"kind",
		"status",
		"status_hook",         // valid for 1.6 and 1.7,
		"multiple_executions", // valid for 1.7
		"numeric_policy_id",   // valid for 1.8
	}

	count := 0
	for _, fullID := range allJobIDs {
		args[0] = fullID
		values, err := redis.Values(conn.Do("HMGET", args...))
		if err != nil {
			logger.Errorf("Get stats fields of job %s failed with error: %s", fullID, err)
			continue
		}

		pID := toString(values[0])
		kind := toString(values[1])

		if !utils.IsEmptyStr(pID) && job.KindPeriodic == kind {
			logger.Debugf("Periodic job found: %s", pID)

			// Data requires migration
			// Missing 'numeric_policy_id' which is introduced in 1.8
			if values[5] == nil {
				logger.Infof("Migrate periodic job stats data is started: %s", pID)

				numbericPolicyID, err := getScoreByID(pID, conn, pm.namespace)
				if err != nil {
					logger.Errorf("Get numberic ID of periodic job policy failed with error: %s", err)
					continue
				}

				// Transaction
				err = conn.Send("MULTI")
				setArgs := []interface{}{
					fullID,
					"status",
					job.ScheduledStatus.String(), // make sure the status of periodic job is "Scheduled"
					"numeric_policy_id",
					numbericPolicyID,
				}
				// If status hook existing
				hookURL := toString(values[3])
				if !utils.IsEmptyStr(hookURL) {
					setArgs = append(setArgs, "web_hook_url", hookURL)
				}
				// Set fields
				err = conn.Send("HMSET", setArgs...)

				// Remove useless fields
				rmArgs := []interface{}{
					fullID,
					"status_hook",
					"multiple_executions",
				}
				err = conn.Send("HDEL", rmArgs...)

				// Update periodic policy model
				// conn is working, we need new conn
				// this inner connection will be closed by the calling method
				innerConn := pm.pool.Get()

				policy, er := getPeriodicPolicy(numbericPolicyID, innerConn, pm.namespace)
				if er == nil {
					policy.ID = pID
					if !utils.IsEmptyStr(hookURL) {
						// Copy web hook URL
						policy.WebHookURL = fmt.Sprintf("%s", hookURL)
					}

					if rawJSON, er := policy.Serialize(); er == nil {
						// Remove the old one first
						err = conn.Send("ZREMRANGEBYSCORE", rds.KeyPeriodicPolicy(pm.namespace), numbericPolicyID, numbericPolicyID)
						// Save back to the rdb
						err = conn.Send("ZADD", rds.KeyPeriodicPolicy(pm.namespace), numbericPolicyID, rawJSON)
					} else {
						logger.Errorf("Serialize policy %s failed with error: %s", pID, er)
					}
				} else {
					logger.Errorf("Get periodic policy %s failed with error: %s", pID, er)
				}

				// Check error before executing
				if err != nil {
					logger.Errorf("Build redis transaction failed with error: %s", err)
					continue
				}

				// Exec
				if _, err := conn.Do("EXEC"); err != nil {
					logger.Errorf("Migrate periodic job %s failed with error: %s", pID, err)
					continue
				}

				count++
				logger.Infof("Migrate periodic job stats data is completed: %s", pID)
			}
		}
	}

	logger.Infof("Migrate %d periodic policies", count)

	delScoreZset(conn, pm.namespace)

	return clearDuplicatedPolicies(conn, pm.namespace)
}

// getAllJobStatsIDs get all the IDs of the existing jobs
func getAllJobStatsIDs(conn redis.Conn, ns string) ([]string, error) {
	pattern := rds.KeyJobStats(ns, "*")
	args := []interface{}{
		0,
		"MATCH",
		pattern,
		"COUNT",
		100,
	}

	allFullIDs := make([]interface{}, 0)

	for {
		// Use SCAN to iterate the IDs
		values, err := redis.Values(conn.Do("SCAN", args...))
		if err != nil {
			return nil, err
		}

		// In case something wrong happened
		if len(values) != 2 {
			return nil, errors.Errorf("Invalid result returned for the SCAN command: %#v", values)
		}

		if fullIDs, ok := values[1].([]interface{}); ok {
			allFullIDs = append(allFullIDs, fullIDs...)
		}

		// Check the next cursor
		cur := toInt(values[0])
		if cur == -1 {
			// No valid next cursor got
			return nil, errors.Errorf("Failed to get the next SCAN cursor: %#v", values[0])
		}

		if cur != 0 {
			args[0] = cur
		} else {
			// end
			break
		}
	}

	IDs := make([]string, 0)
	for _, fullIDValue := range allFullIDs {
		if fullID, ok := fullIDValue.([]byte); ok {
			IDs = append(IDs, string(fullID))
		} else {
			logger.Debugf("Invalid job stats key: %#v", fullIDValue)
		}
	}

	return IDs, nil
}

// Get the score with the provided ID
func getScoreByID(id string, conn redis.Conn, ns string) (int64, error) {
	scoreKey := fmt.Sprintf("%s%s:%s", rds.KeyNamespacePrefix(ns), "period", "key_score")
	return redis.Int64(conn.Do("ZSCORE", scoreKey, id))
}

// Get periodic policy object by the numeric ID
func getPeriodicPolicy(numericID int64, conn redis.Conn, ns string) (*period.Policy, error) {
	// close this inner connection here
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Errorf("close redis connection error: %s", err)
		}
	}()

	bytes, err := redis.Values(conn.Do("ZRANGEBYSCORE", rds.KeyPeriodicPolicy(ns), numericID, numericID))
	if err != nil {
		return nil, err
	}

	p := &period.Policy{}
	if len(bytes) > 0 {
		if rawPolicy, ok := bytes[0].([]byte); ok {
			if err = p.DeSerialize(rawPolicy); err == nil {
				return p, nil
			}
		}
	}

	if err == nil {
		err = errors.Errorf("invalid data for periodic policy %d: %#v", numericID, bytes)
	}

	return nil, err
}

// Clear the duplicated policy entries for the job "IMAGE_GC" and "IMAGE_SCAN_ALL"
func clearDuplicatedPolicies(conn redis.Conn, ns string) error {
	hash := make(map[string]interface{})

	bytes, err := redis.Values(conn.Do("ZREVRANGE", rds.KeyPeriodicPolicy(ns), 0, -1, "WITHSCORES"))
	if err != nil {
		return err
	}

	count := 0
	for i, l := 0, len(bytes); i < l; i = i + 2 {
		rawPolicy := bytes[i].([]byte)
		p := &period.Policy{}

		if err := p.DeSerialize(rawPolicy); err != nil {
			logger.Errorf("DeSerialize policy: %s; error: %s\n", rawPolicy, err)
			continue
		}

		if p.JobName == jobNameScanAll ||
			p.JobName == jobNameGarbageCollection ||
			p.JobName == jobNameReplicationScheduler {
			score, _ := strconv.ParseInt(string(bytes[i+1].([]byte)), 10, 64)

			key := hashKey(p)
			if _, exists := hash[key]; exists {
				// Already existing, remove the duplicated one
				res, err := redis.Int(conn.Do("ZREMRANGEBYSCORE", rds.KeyPeriodicPolicy(ns), score, score))
				if err != nil || res == 0 {
					logger.Errorf("Failed to clear duplicated periodic policy: %s-%s:%v", p.JobName, p.ID, score)
				} else {
					logger.Infof("Remove duplicated periodic policy: %s-%s:%v", p.JobName, p.ID, score)
					count++
				}
			} else {
				hash[key] = score
			}
		}
	}

	logger.Infof("Clear %d duplicated periodic policies", count)

	return nil
}

// Remove the non-used key
func delScoreZset(conn redis.Conn, ns string) {
	key := fmt.Sprintf("%s%s", rds.KeyNamespacePrefix(ns), "period:key_score")
	reply, err := redis.Int(conn.Do("EXISTS", key))
	if err == nil && reply == 1 {
		reply, err = redis.Int(conn.Do("DEL", key))
		if err == nil && reply > 0 {
			logger.Infof("%s removed", key)
			return // success
		}
	}

	if err != nil {
		// Just logged
		logger.Errorf("Remove %s failed with error: %s", key, err)
	}
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}

	if bytes, ok := v.([]byte); ok {
		return string(bytes)
	}

	return ""
}

func toInt(v interface{}) int64 {
	if v == nil {
		return -1
	}

	if bytes, ok := v.([]byte); ok {
		if intV, err := strconv.ParseInt(string(bytes), 10, 64); err == nil {
			return intV
		}
	}

	return -1
}

func hashKey(p *period.Policy) string {
	key := p.JobName
	if p.JobParameters != nil && len(p.JobParameters) > 0 {
		if bytes, err := json.Marshal(p.JobParameters); err == nil {
			key = fmt.Sprintf("%s:%s", key, string(bytes))
		}
	}

	return base64.StdEncoding.EncodeToString([]byte(key))
}
