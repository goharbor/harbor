import { ScrollManagerService } from './scroll-manager.service';
import { Directive, HostListener, Input } from '@angular/core';

@Directive({
    selector: '[appScrollAnchor]',
})
export class ScrollAnchorDirective {
    @Input('appScrollAnchor') id: string | number;

    constructor(private manager: ScrollManagerService) {}

    @HostListener('click')
    scroll() {
        this.manager.scroll(this.id);
    }
}
