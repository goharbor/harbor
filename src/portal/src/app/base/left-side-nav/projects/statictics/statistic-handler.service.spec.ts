import { TestBed, inject } from '@angular/core/testing';

import { StatisticHandler } from './statistic-handler.service';

describe('StatisticHandlerService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            providers: [StatisticHandler],
        });
    });

    it('should be created', inject(
        [StatisticHandler],
        (service: StatisticHandler) => {
            expect(service).toBeTruthy();
        }
    ));
});
