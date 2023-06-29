import { PullCommandComponent } from './pull-command.component';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SharedTestingModule } from '../../../../../../../../shared/shared.module';

describe('PullCommandComponent', () => {
    let component: PullCommandComponent;
    let fixture: ComponentFixture<PullCommandComponent>;
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [PullCommandComponent],
            imports: [SharedTestingModule],
        }).compileComponents();

        fixture = TestBed.createComponent(PullCommandComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
