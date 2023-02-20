import { HarborRouteReuseStrategy } from './harbor-route-reuse-strategy';
import { ActivatedRouteSnapshot } from '@angular/router';

describe('HarborRouteReuseStrategy', () => {
    let harborRouteReuseStrategy: HarborRouteReuseStrategy;
    beforeEach(() => {
        harborRouteReuseStrategy = new HarborRouteReuseStrategy();
    });
    it('should be created', () => {
        expect(harborRouteReuseStrategy).toBeTruthy();
    });
    it('shouldReuseRoute', () => {
        const future: ActivatedRouteSnapshot = new ActivatedRouteSnapshot();
        const curr: ActivatedRouteSnapshot = new ActivatedRouteSnapshot();
        expect(
            harborRouteReuseStrategy.shouldReuseRoute(future, curr)
        ).toBeTruthy();
    });
});
