import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ChartDetailDependencyComponent } from './chart-detail-dependency.component';

describe('ChartDetailDependencyComponent', () => {
    let component: ChartDetailDependencyComponent;
    let fixture: ComponentFixture<ChartDetailDependencyComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot()
            ],
            declarations: [ChartDetailDependencyComponent],
            providers: [
                TranslateService
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ChartDetailDependencyComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
