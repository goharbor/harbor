import { ScrollSectionDirective } from './scroll-section.directive';
import { Injectable } from '@angular/core';

@Injectable({
    providedIn: 'root',
})
export class ScrollManagerService {
    private sections = new Map<string | number, ScrollSectionDirective>();

    scroll(id: string | number) {
        this.sections.get(id)!.scroll();
    }

    register(section: ScrollSectionDirective) {
        this.sections.set(section.id, section);
    }

    remove(section: ScrollSectionDirective) {
        this.sections.delete(section.id);
    }
}
