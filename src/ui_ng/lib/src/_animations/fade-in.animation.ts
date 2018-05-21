// import the required animation functions from the angular animations module
import {AnimationTriggerMetadata, trigger, animate, transition, style } from '@angular/animations';

export const FadeInAnimation: AnimationTriggerMetadata =
    // trigger name for attaching this animation to an element using the [@triggerName] syntax
    trigger('FadeInAnimation', [

        // route 'enter' transition
        transition(':enter', [

            // css styles at start of transition
            style({ opacity: 0 }),

            // animation and styles at end of transition
            animate('.3s', style({ opacity: 1 }))
        ]),
    ]);
