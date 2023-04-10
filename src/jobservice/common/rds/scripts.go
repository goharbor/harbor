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

package rds

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

const (
	requeueKeysPerJob = 4
)

// luaFuncStCodeText is common lua script function
var luaFuncStCodeText = `
-- for easily compare status
local stMap = { ['Pending'] = 0, ['Scheduled'] = 1, ['Running'] = 2, ['Success'] = 3, ['Stopped'] = 3, ['Error'] = 3 }

local function stCode(status)
  -- return 0 as default status
  return stMap[status] or 0
end
`

// luaFuncCompareText is common lua script function
var luaFuncCompareText = `
local function compare(status, revision)
  local sCode = stCode(status)
  local aCode = stCode(ARGV[1])
  local aRev = tonumber(ARGV[2]) or 0
  local aCheckInT = tonumber(ARGV[3]) or 0
  if revision < aRev or
    ( revision == aRev and sCode <= aCode ) or
    ( revision == aRev and aCheckInT ~= 0 )
  then
     return 'ok'
  end
  return 'no'
end
`

// Script used to set the status of the job
//
// KEY[1]: key of job stats
// KEY[2]: key of inprogress track
// ARGV[1]: status text
// ARGV[2]: stats revision
// ARGV[3]: update timestamp
// ARGV[4]: job ID
var setStatusScriptText = fmt.Sprintf(`
%s

local res, st, code, rev, aCode, aRev

res = redis.call('hmget', KEYS[1], 'status', 'revision')
if res then
  st = res[1]
  code = stCode(st)
  aCode = stCode(ARGV[1])
  rev = tonumber(res[2]) or 0
  aRev = tonumber(ARGV[2]) or 0

  -- set same status repeatedly is allowed
  if rev < aRev or ( rev == aRev and (code < aCode or st == ARGV[1])) then
    redis.call('hmset', KEYS[1], 'status', ARGV[1], 'update_time', ARGV[3])
    -- update inprogress track if necessary
    if aCode == 3 then
      -- final status
      local c = redis.call('hincrby', KEYS[2], ARGV[4], -1)
      -- lock count is 0, del it
      if c <= 0 then
        redis.call('hdel', KEYS[2], ARGV[4])
      end

      if ARGV[1] == 'Success' or ARGV[1] == 'Stopped' then
        -- expire the job stats with shorter interval (1 day)
        redis.call('expire', KEYS[1], 86400)
      elseif ARGV[1] == 'Error' then
        -- expire the job stats with normal interval (7 days) incase it may be retried again
        redis.call('expire', KEYS[1], 604800)
      else
        -- remove the expire time if existing
        redis.call('persist', KEYS[1])
      end
    end

    return 'ok'
  end
end

return st
`, luaFuncStCodeText)

// SetStatusScript is lua script for setting job status atomically
var SetStatusScript = redis.NewScript(2, setStatusScriptText)

// Used to set the hook ACK
//
// KEY[1]: key of job stats
// KEY[2]: key of inprogress track
// ARGV[1]: job status
// ARGV[2]: revision of job stats
// ARGV[3]: check in timestamp
// ARGV[4]: job ID
var hookAckScriptText = fmt.Sprintf(`
%s

%s

local function canSetAck(jk, nrev)
  local res = redis.call('hmget', jk, 'revision', 'ack')
  if res then
    local rev = tonumber(res[1]) or 0
    local ackv = res[2]

    if ackv then
      -- ack existing
      local ack = cjson.decode(ackv)
      local cmp = compare(ack['status'], ack['revision'])
      if cmp == 'ok' then
        return 'ok'
      end
    else
      -- no ack yet
      if rev <= nrev then
        return 'ok'
      end
    end
  end

  return nil
end

if canSetAck(KEYS[1], tonumber(ARGV[2])) ~= 'ok' then
  return 'none'
end

-- can-set-ack check is ok
-- set new ack
local newAck = {['status'] = ARGV[1], ['revision'] = tonumber(ARGV[2]), ['check_in_at'] = tonumber(ARGV[3])}
local ackJson = cjson.encode(newAck)

redis.call('hset', KEYS[1], 'ack', ackJson)

-- update the inprogress track
if stCode(ARGV[1]) == 3 then
  -- final status
  local c = redis.call('hincrby', KEYS[2], ARGV[4], -1)
  -- lock count is 0, del it
  if c <= 0 then
     redis.call('hdel', KEYS[2], ARGV[4])
  end
end

return 'ok'
`, luaFuncStCodeText, luaFuncCompareText)

// HookAckScript is defined to set the hook event ACK in the job stats map
var HookAckScript = redis.NewScript(2, hookAckScriptText)

// Used to reset job status
//
// KEYS[1]: key of job stats
// KEYS[2]: key of inprogress job track
// ARGV[1]: job ID
// ARGV[2]: start status
// ARGV[3]: revision of job stats
var statusResetScriptText = `
local now = tonumber(ARGV[3]) or 0

-- reset status and revision
redis.call('hmset', KEYS[1], 'status', ARGV[2], 'revision', now, 'update_time', now)
redis.call('hdel', KEYS[1], 'ack', 'check_in', 'check_in_at')

-- reset inprogress track
redis.call('hset', KEYS[2], ARGV[1], 2)
`

// StatusResetScript is lua script to reset the job stats
var StatusResetScript = redis.NewScript(2, statusResetScriptText)

// Copy from upstream worker framework
// Used by the reaper to re-enqueue jobs that were in progress
//
// KEYS[1] = the 1st job's in progress queue
// KEYS[2] = the 1st job's job queue
// KEYS[3] = the 2nd job's in progress queue
// KEYS[4] = the 2nd job's job queue
// ...
// KEYS[N] = the last job's in progress queue
// KEYS[N+1] = the last job's job queue
// ARGV[1] = workerPoolID for job queue
var redisLuaReenqueueJob = fmt.Sprintf(`
local function releaseLock(lockKey, lockInfoKey, workerPoolID)
  redis.call('decr', lockKey)
  redis.call('hincrby', lockInfoKey, workerPoolID, -1)
end

local keylen = #KEYS
local res, jobQueue, inProgQueue, workerPoolID, lockKey, lockInfoKey
workerPoolID = ARGV[1]

for i=1,keylen,%d do
  inProgQueue = KEYS[i]
  jobQueue = KEYS[i+1]
  lockKey = KEYS[i+2]
  lockInfoKey = KEYS[i+3]
  res = redis.call('rpoplpush', inProgQueue, jobQueue)
  if res then
    releaseLock(lockKey, lockInfoKey, workerPoolID)
    return {res, inProgQueue, jobQueue}
  end
end
return nil`, requeueKeysPerJob)

// RedisLuaReenqueueScript returns redis script of redisLuaReenqueueJob
func RedisLuaReenqueueScript(jobTypesCount int) *redis.Script {
	return redis.NewScript(jobTypesCount*requeueKeysPerJob, redisLuaReenqueueJob)
}

var saveJobScript = `
local function is_blank(str)
  return str == nil or str == '' or string.match(str, "^%s*$") ~= nil
end

local key1 = KEYS[1] -- rds.KeyJobStats(bt.namespace, stats.Info.JobID)
local key2 = KEYS[2] -- rds.KeyJobTrackInProgress(bt.namespace)
local key3 = KEYS[3] -- rds.KeyUpstreamJobAndExecutions(bt.namespace, stats.Info.UpstreamJobID)

local jobID = ARGV[1]
local jobName = ARGV[2]
local jobKind = ARGV[3]
local isUnique = tostring(ARGV[4])
local status = ARGV[5]
local refLink = ARGV[6]
local enqueueTime = tonumber(ARGV[7])
local runAt = tonumber(ARGV[8])
local cronSpec = ARGV[9]
local webHookURL = ARGV[10]
local numericPID = tonumber(ARGV[11])
local checkInAt = tonumber(ARGV[12])
local dieAt = tonumber(ARGV[13])
local upstreamJobID = ARGV[14]
local parameters = ARGV[15]
local currentTime = ARGV[16]
local revision = ARGV[17]

local args = {}
table.insert(args, "id")
table.insert(args, jobID)
table.insert(args, "name")
table.insert(args, jobName)
table.insert(args, "kind")
table.insert(args, jobKind)
table.insert(args, "unique")
table.insert(args, isUnique)
table.insert(args, "status")
table.insert(args, status)
table.insert(args, "ref_link")
table.insert(args, refLink)
table.insert(args, "enqueue_time")
table.insert(args, enqueueTime)
table.insert(args, "run_at")
table.insert(args, runAt)
table.insert(args, "cron_spec")
table.insert(args, cronSpec)
table.insert(args, "web_hook_url")
table.insert(args, webHookURL)
table.insert(args, "numeric_policy_id")
table.insert(args, numericPID)

if checkInAt > 0 then
  table.insert(args, "check_in")
  table.insert(args, "[REDUNDANT]")
  table.insert(args, "check_in_at")
  table.insert(args, checkInAt)
end

if dieAt > 0 then
  table.insert(args, "die_at")
  table.insert(args, dieAt)
end

if not is_blank(upstreamJobID) then
  table.insert(args, "upstream_job_id")
  table.insert(args, upstreamJobID)
end

if not is_blank(parameters) then
  table.insert(args, "parameters")
  table.insert(args, parameters)
end

table.insert(args, "update_time")
table.insert(args, currentTime)

table.insert(args, "revision")
table.insert(args, revision)

-- add other argv
for i = 18, #ARGV, 1 do
  table.insert(args, ARGV[i])
end

local result1 = redis.call("HMSET", key1, unpack(args))
if not result1 then
  error("HMSET error")
end

local result2 = redis.call("HSET", key2, jobID, 2)
if not result2 then
  error("HSET error")
end
if not is_blank(upstreamJobID) then
  local result3 = redis.call("ZADD", key3, "NX", runAt, jobID)
  if not result3 then
    error("ZADD error")
  end
end
`

func SaveScript() *redis.Script {
	return redis.NewScript(3, saveJobScript)
}
