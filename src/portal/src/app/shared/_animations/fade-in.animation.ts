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
    animate,
    transition,
    style,
} from '@angular/animations';

export const FadeInAnimation: AnimationTriggerMetadata =
    // trigger name for attaching this animation to an element using the [@triggerName] syntax
    trigger('FadeInAnimation', [
        // router-guard 'enter' transition
        transition(':enter', [
            // css styles at start of transition
            style({ opacity: 0 }),

            // animation and styles at end of transition
            animate('.3s', style({ opacity: 1 })),
        ]),
    ]);
