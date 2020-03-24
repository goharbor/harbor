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

package or

import (
	"github.com/goharbor/harbor/src/lib/selector"
	"sync"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/retention/policy/action"
	"github.com/goharbor/harbor/src/pkg/retention/policy/alg"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/pkg/errors"
)

// processor to handle the rules with OR mapping ways
type processor struct {
	// keep evaluator and its related selector if existing
	// attentions here, the selectors can be empty/nil, that means match all "**"
	evaluators map[*rule.Evaluator][]selector.Selector
	// action performer
	performers map[string]action.Performer
}

// New processor
func New(parameters []*alg.Parameter) alg.Processor {
	p := &processor{
		evaluators: make(map[*rule.Evaluator][]selector.Selector),
		performers: make(map[string]action.Performer),
	}

	if len(parameters) > 0 {
		for _, param := range parameters {
			if param.Evaluator != nil {
				if len(param.Selectors) > 0 {
					p.evaluators[&param.Evaluator] = param.Selectors
				}

				if param.Performer != nil {
					p.performers[param.Evaluator.Action()] = param.Performer
				}
			}
		}
	}

	return p
}

// Process the candidates with the rules
func (p *processor) Process(artifacts []*selector.Candidate) ([]*selector.Result, error) {
	if len(artifacts) == 0 {
		log.Debug("no artifacts to retention")
		return make([]*selector.Result, 0), nil
	}

	var (
		// collect errors by wrapping
		err error
		// collect processed candidates
		processedCandidates = make(map[string]cHash)
	)

	// for sync
	type chanItem struct {
		action    string
		processed []*selector.Candidate
	}

	resChan := make(chan *chanItem, 1)
	// handle error
	errChan := make(chan error, 1)
	// control chan
	done := make(chan bool, 1)

	// go routine for receiving results/error
	go func() {
		defer func() {
			// done
			done <- true
		}()

		for {
			select {
			case result := <-resChan:
				if result == nil {
					// chan is closed
					return
				}

				if _, ok := processedCandidates[result.action]; !ok {
					processedCandidates[result.action] = make(cHash)
				}

				listByAction := processedCandidates[result.action]
				for _, rp := range result.processed {
					// remove duplicated ones
					listByAction[rp.Hash()] = rp
				}
			case e := <-errChan:
				if err == nil {
					err = errors.Wrap(e, "artifact processing error")
				} else {
					err = errors.Wrap(e, err.Error())
				}
			}
		}
	}()

	wg := new(sync.WaitGroup)
	wg.Add(len(p.evaluators))

	for eva, selectors := range p.evaluators {
		var evaluator = *eva

		go func(evaluator rule.Evaluator, selectors []selector.Selector) {
			var (
				processed []*selector.Candidate
				err       error
			)

			defer func() {
				wg.Done()
			}()

			// init
			// pass array copy to the selector
			processed = append(processed, artifacts...)

			if len(selectors) > 0 {
				// selecting artifacts one by one
				// `&&` mappings
				for _, s := range selectors {
					if processed, err = s.Select(processed); err != nil {
						errChan <- err
						return
					}
				}
			}

			if processed, err = evaluator.Process(processed); err != nil {
				errChan <- err
				return
			}

			// Pass to the outside
			resChan <- &chanItem{
				action:    evaluator.Action(),
				processed: processed,
			}
		}(evaluator, selectors)
	}

	// waiting for all the rules are evaluated
	wg.Wait()
	// close result chan
	close(resChan)
	// check if the receiving loop exists
	<-done

	if err != nil {
		return nil, err
	}

	results := make([]*selector.Result, 0)
	// Perform actions
	for act, hash := range processedCandidates {
		var attachedErr error

		cl := hash.toList()

		if pf, ok := p.performers[act]; ok {
			if theRes, err := pf.Perform(cl); err != nil {
				attachedErr = err
			} else {
				results = append(results, theRes...)
			}
		} else {
			attachedErr = errors.Errorf("no performer added for action %s in OR processor", act)
		}

		if attachedErr != nil {
			for _, c := range cl {
				results = append(results, &selector.Result{
					Target: c,
					Error:  attachedErr,
				})
			}
		}
	}

	return results, nil
}

type cHash map[string]*selector.Candidate

func (ch cHash) toList() []*selector.Candidate {
	l := make([]*selector.Candidate, 0)

	for _, v := range ch {
		l = append(l, v)
	}

	return l
}
