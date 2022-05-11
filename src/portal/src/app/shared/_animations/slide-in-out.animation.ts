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
