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

package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
)

// TelegramJob implements the job interface, which send notification to telegram via bot API.
type TelegramJob struct {
	client *http.Client
	logger logger.Interface
}

// MaxFails returns that how many times this job can fail.
func (tg *TelegramJob) MaxFails() (result uint) {
	// Default max fails count is 3
	result = 3
	if maxFails, exist := os.LookupEnv(maxFails); exist {
		mf, err := strconv.ParseUint(maxFails, 10, 32)
		if err != nil {
			logger.Warningf("Fetch telegram job maxFails error: %s", err.Error())
			return result
		}
		result = uint(mf)
	}
	return result
}

// MaxCurrency is implementation of same method in Interface.
func (tg *TelegramJob) MaxCurrency() uint {
	return 1
}

// ShouldRetry ...
func (tg *TelegramJob) ShouldRetry() bool {
	return true
}

// Validate implements the interface in job/Interface
func (tg *TelegramJob) Validate(params job.Parameters) error {
	if params == nil {
		// Params are required
		return errors.New("missing parameter of telegram job")
	}

	text, ok := params["text"]
	if !ok {
		return errors.Errorf("missing job parameter 'text'")
	}
	_, ok = text.(string)
	if !ok {
		return errors.Errorf("malformed job parameter 'text', expecting string but got %s", reflect.TypeOf(text).String())
	}

	botToken, ok := params["bot_token"]
	if !ok {
		return errors.Errorf("missing job parameter 'bot_token'")
	}
	_, ok = botToken.(string)
	if !ok {
		return errors.Errorf("malformed job parameter 'bot_token', expecting string but got %s", reflect.TypeOf(botToken).String())
	}

	chatID, ok := params["chat_id"]
	if !ok {
		return errors.Errorf("missing job parameter 'chat_id'")
	}
	_, ok = chatID.(string)
	if !ok {
		return errors.Errorf("malformed job parameter 'chat_id', expecting string but got %s", reflect.TypeOf(chatID).String())
	}
	return nil
}

// Run implements the interface in job/Interface
func (tg *TelegramJob) Run(ctx job.Context, params job.Parameters) error {
	if err := tg.init(ctx, params); err != nil {
		return err
	}

	tg.logger.Info("start to run telegram job")

	err := tg.execute(params)
	if err != nil {
		tg.logger.Errorf("exit telegram job, error: %s", err)
	} else {
		tg.logger.Info("success to run telegram job")
	}
	return err
}

// init telegram job
func (tg *TelegramJob) init(ctx job.Context, params map[string]any) error {
	tg.logger = ctx.GetLogger()

	// default use secure transport
	tg.client = httpHelper.clients[secure]
	if v, ok := params["skip_cert_verify"]; ok {
		if skipCertVerify, ok := v.(bool); ok && skipCertVerify {
			// if skip cert verify is true, it means not verify remote cert, use insecure client
			tg.client = httpHelper.clients[insecure]
		}
	}
	return nil
}

// execute telegram job
func (tg *TelegramJob) execute(params map[string]any) error {
	text := params["text"].(string)
	botToken := params["bot_token"].(string)
	chatID := params["chat_id"].(string)

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	payload := map[string]string{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "failed to marshal telegram payload")
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payloadBytes))
	if err != nil {
		return errors.Wrap(err, "error to generate request")
	}
	req.Header.Set("Content-Type", "application/json")

	tg.logger.Infof("send request to Telegram API, payload: %s", string(payloadBytes))

	resp, err := tg.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "error to send request")
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			tg.logger.Errorf("error to read response body, error: %s", err)
		}

		return errors.Errorf("abnormal response code: %d, body: %s", resp.StatusCode, string(body))
	}
	return nil
}
