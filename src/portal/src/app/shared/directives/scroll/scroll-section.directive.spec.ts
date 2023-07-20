import { ScrollManagerService } from './scroll-manager.service';
import { ScrollSectionDirective } from './scroll-section.directive';

describe('ScrollSectionDirective', () => {
    it('should create an instance', () => {
        const directive = new ScrollSectionDirective(
            { nativeElement: <HTMLDivElement>document.createElement('div') },
            new ScrollManagerService()
        );
        expect(directive).toBeTruthy();
    });
});
