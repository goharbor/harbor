import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { LabelFilterComponent } from './label-filter.component';
import { ClarityModule } from '@clr/angular';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormsModule } from '@angular/forms';

describe('LabelFilterComponent', () => {
    let component: LabelFilterComponent;
    let fixture: ComponentFixture<LabelFilterComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
              ],
              imports: [
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule
              ],
            declarations: [LabelFilterComponent],
            providers: [
                TranslateService
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(LabelFilterComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
