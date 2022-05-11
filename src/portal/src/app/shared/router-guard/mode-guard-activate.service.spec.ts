import { TestBed, inject } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { ModeGuard } from './mode-guard-activate.service';
import { AppConfigService } from '../../services/app-config.service';

describe('ModeGuardActivateServiceGuard', () => {
    const fakeAppConfigService = null;
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [RouterTestingModule],
            providers: [
                ModeGuard,
                { provide: AppConfigService, useValue: fakeAppConfigService },
            ],
        }).compileComponents();
    });

    it('should ...', inject([ModeGuard], (guard: ModeGuard) => {
        expect(guard).toBeTruthy();
    }));
});
