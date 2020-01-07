import { ComponentFixture, fakeAsync, TestBed, tick } from '@angular/core/testing';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { TranslateService } from "@ngx-translate/core";
import { OperationComponent } from './operation.component';
import { OperationService } from './operation.service';
import { HarborLibraryModule } from '../../harbor-library.module';
import { IServiceConfig, SERVICE_CONFIG } from '../../entities/service.config';
import { OperateInfo } from './operate';

describe('OperationComponent', () => {

    let component: OperationComponent;
    let fixture: ComponentFixture<OperationComponent>;
    let config: IServiceConfig = {
        configurationEndpoint: '/api/configurations/testing'
    };

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [
                HarborLibraryModule,
                BrowserAnimationsModule,
            ],
            providers: [
                OperationService,
                TranslateService,
                {provide: SERVICE_CONFIG, useValue: config}
            ]
        });
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(OperationComponent);
        component = fixture.componentInstance;
        component.animationState = 'out';
        fixture.detectChanges();
    });
    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should automatically close', fakeAsync(async () => {
        component.animationState = "in";
        fixture.detectChanges();
        // wait animation finishing
        tick(1000);
        await fixture.whenStable();
        const container: HTMLDivElement = fixture.nativeElement.querySelector(".operDiv");
        container.dispatchEvent(new Event('mouseleave'));
        fixture.detectChanges();
        // wait animation finishing
        tick(10000);
        await fixture.whenStable();
        const right: string = getComputedStyle(fixture.nativeElement.querySelector(".operDiv")).right;
        expect(right).toEqual("-325px");
    }));
    it("should show '50+' after pushing 60 new operateInfos", fakeAsync(() => {
        const operationService: OperationService = TestBed.get(OperationService);
        for (let i = 0; i < 60; i++) {
            let operateInfo = new OperateInfo();
            if (i > 19) {
                operateInfo.state = "progressing";
            }
            if ( i > 39 ) {
                operateInfo.state = "failure";
            }
            tick(50000);
            operationService.publishInfo(operateInfo);
        }
        fixture.detectChanges();
        const toolBar: HTMLAnchorElement = fixture.nativeElement.querySelector(".toolBar");
        expect(toolBar.textContent).toContain('50+');
    }));
    it('check toggleTitle function', () => {
        const errorSpan: HTMLSpanElement = document.createElement('span');
        errorSpan.style.display = "none";
        component.toggleTitle(errorSpan);
        expect(errorSpan.style.display).toEqual('block');
        component.toggleTitle(errorSpan);
        expect(errorSpan.style.display).toEqual('none');
    });
    it('check calculateTime function', () => {
        expect(component.calculateTime(1000)).toEqual('less than 1 minute');
        expect(component.calculateTime(61000)).toEqual('1 minute(s) ago');
        expect(component.calculateTime(3601000)).toEqual('1 hour(s) ago');
        expect(component.calculateTime(24 * 3601000)).toEqual('1 day(s) ago');
    });
});
