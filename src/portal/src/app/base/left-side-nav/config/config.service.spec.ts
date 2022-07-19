import { TestBed, inject } from '@angular/core/testing';
import { ConfigureService } from 'ng-swagger-gen/services/configure.service';
import { of } from 'rxjs';
import { SharedTestingModule } from '../../../shared/shared.module';
import { Configuration } from './config';
import { ConfigService } from './config.service';

describe('ConfigService', () => {
    const fakedConfigureService = {
        getConfigurations(): any {
            return of(null);
        },
    };
    let getConfigSpy: jasmine.Spy;
    beforeEach(() => {
        getConfigSpy = spyOn(
            fakedConfigureService,
            'getConfigurations'
        ).and.returnValue(of(new Configuration()));
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            providers: [
                ConfigService,
                { provide: ConfigureService, useValue: fakedConfigureService },
            ],
        });
    });

    it('should be created', inject(
        [ConfigService],
        (service: ConfigService) => {
            expect(service).toBeTruthy();
        }
    ));

    it('should init config', inject(
        [ConfigService],
        (service: ConfigService) => {
            expect(getConfigSpy.calls.count()).toEqual(0);
            service.initConfig();
            expect(getConfigSpy.calls.count()).toEqual(1);
            // only init once
            service.initConfig();
            expect(getConfigSpy.calls.count()).toEqual(1);
            expect(service).toBeTruthy();
        }
    ));
});
