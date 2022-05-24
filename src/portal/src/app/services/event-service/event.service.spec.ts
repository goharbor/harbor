import { TestBed } from '@angular/core/testing';
import { EventService } from './event.service';
import { Subscription } from 'rxjs';

describe('EventServiceService', () => {
    let service: EventService;

    beforeEach(() => {
        TestBed.configureTestingModule({});
        service = TestBed.inject(EventService);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });

    it('able to subscribe', () => {
        let result: string;
        const sub1 = service.subscribe('testEvent', data => {
            result = data;
        });
        expect(sub1).toBeTruthy();
        expect(sub1 instanceof Subscription).toEqual(true);
        service.publish('testEvent', 'resultString');
        sub1.unsubscribe();
        expect(result).toEqual('resultString');
    });
});
