import { TestBed, inject } from '@angular/core/testing';
import { EXECUTION_STATUS, P2pProviderService } from './p2p-provider.service';

describe('P2pProviderService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            providers: [P2pProviderService],
        });
    });

    it('should be created', inject(
        [P2pProviderService],
        (service: P2pProviderService) => {
            expect(service).toBeTruthy();
        }
    ));
    it('function getDuration should work', inject(
        [P2pProviderService],
        (service: P2pProviderService) => {
            const date = new Date();
            const noDuration: string = service.getDuration(
                new Date(date).toUTCString(),
                new Date(date.getTime()).toUTCString()
            );
            expect(noDuration).toEqual('-');
            const durationMs: string = service.getDuration(
                new Date(date).toUTCString(),
                new Date(date.getTime() + 10).toUTCString()
            );
            expect(durationMs).toEqual('-');
            const durationSec: string = service.getDuration(
                new Date(date).toUTCString(),
                new Date(date.getTime() + 1010).toUTCString()
            );
            expect(durationSec).toEqual('1s');
            const durationMin: string = service.getDuration(
                new Date(date).toUTCString(),
                new Date(date.getTime() + 61010).toUTCString()
            );
            expect(durationMin).toEqual('1m1s');
            const durationMinOnly: string = service.getDuration(
                new Date(date).toUTCString(),
                new Date(date.getTime() + 60000).toUTCString()
            );
            expect(durationMinOnly).toEqual('1m');
        }
    ));
    it('function willChangStatus should work', inject(
        [P2pProviderService],
        (service: P2pProviderService) => {
            expect(
                service.willChangStatus(EXECUTION_STATUS.PENDING)
            ).toBeTruthy();
            expect(
                service.willChangStatus(EXECUTION_STATUS.RUNNING)
            ).toBeTruthy();
            expect(
                service.willChangStatus(EXECUTION_STATUS.SCHEDULED)
            ).toBeTruthy();
        }
    ));
});
