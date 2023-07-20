import { ScrollManagerService } from './scroll-manager.service';
import { Directive, ElementRef, Input, OnDestroy, OnInit } from '@angular/core';

@Directive({
    selector: '[appScrollSection]',
})
export class ScrollSectionDirective implements OnInit, OnDestroy {
    @Input('appScrollSection') id: string | number;

    constructor(
        private host: ElementRef<HTMLElement>,
        private manager: ScrollManagerService
    ) {}

    ngOnInit() {
        this.manager.register(this);
    }

    ngOnDestroy() {
        this.manager.remove(this);
    }

    scroll() {
        this.host.nativeElement.scrollIntoView({
            behavior: 'smooth',
        });
    }
}
