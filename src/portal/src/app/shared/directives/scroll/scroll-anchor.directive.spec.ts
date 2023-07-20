import { ScrollManagerService } from './scroll-manager.service';
import { ScrollAnchorDirective } from './scroll-anchor.directive';

describe('ScrollAnchorDirective', () => {
    it('should create an instance', () => {
        const directive = new ScrollAnchorDirective(new ScrollManagerService());
        expect(directive).toBeTruthy();
    });
});
