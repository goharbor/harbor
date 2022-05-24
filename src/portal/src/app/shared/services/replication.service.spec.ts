import { TestBed, inject } from '@angular/core/testing';
import {
    ReplicationService,
    ReplicationDefaultService,
} from './replication.service';
import { SharedTestingModule } from '../shared.module';
import { HttpClientTestingModule } from '@angular/common/http/testing';

describe('ReplicationService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule, HttpClientTestingModule],
            providers: [
                ReplicationDefaultService,
                {
                    provide: ReplicationService,
                    useClass: ReplicationDefaultService,
                },
            ],
        });
    });

    it('should be initialized', inject(
        [ReplicationDefaultService],
        (service: ReplicationService) => {
            expect(service).toBeTruthy();
        }
    ));
});
