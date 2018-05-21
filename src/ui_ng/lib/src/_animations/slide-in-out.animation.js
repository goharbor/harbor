// import the required animation functions from the angular animations module
import { trigger, state, animate, transition, style } from '@angular/animations';
export var SlideInOutAnimation = 
// trigger name for attaching this animation to an element using the [@triggerName] syntax
trigger('SlideInOutAnimation', [
    // end state styles for route container (host)
    state('in', style({
        // the view covers the whole screen with a semi tranparent background
        position: 'fix',
        right: 0,
        width: '340px',
        bottom: 0
        // backgroundColor: 'rgba(0, 0, 0, 0.8)'
    })),
    state('out', style({
        // the view covers the whole screen with a semi tranparent background
        position: 'fix',
        width: '30px',
        bottom: 0
        // backgroundColor: 'rgba(0, 0, 0, 0.8)'
    })),
    // route 'enter' transition
    transition('out => in', [
        // animation and styles at end of transition
        animate('.5s ease-in-out', style({
            // transition the right position to 0 which slides the content into view
            width: '340px',
        }))
    ]),
    // route 'leave' transition
    transition('in => out', [
        // animation and styles at end of transition
        animate('.5s ease-in-out', style({
            // transition the right position to -400% which slides the content out of view
            width: '30px',
        }))
    ])
]);
//# sourceMappingURL=slide-in-out.animation.js.map