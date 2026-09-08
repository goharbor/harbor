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

// import the required animation functions from the angular animations module
import {
    AnimationTriggerMetadata,
    trigger,
    state,
    animate,
    transition,
    style,
} from '@angular/animations';

export const SlideInOutAnimation: AnimationTriggerMetadata =
    // trigger name for attaching this animation to an element using the [@triggerName] syntax
    trigger('SlideInOutAnimation', [
        state(
            'in',
            style({
                right: 0,
            })
        ),
        state(
            'out',
            style({
                right: '-325px',
            })
        ),
        transition('in <=> out', [animate('0.5s ease-in-out')]),
    ]);
