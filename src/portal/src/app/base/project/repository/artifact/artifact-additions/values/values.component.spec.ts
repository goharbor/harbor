import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA, SecurityContext } from '@angular/core';
import { MarkdownModule, MarkedOptions } from 'ngx-markdown';
import { ValuesComponent } from './values.component';
import { AdditionsService } from '../additions.service';
import { of } from 'rxjs';
import { AdditionLink } from '../../../../../../../../ng-swagger-gen/models/addition-link';
import { SharedTestingModule } from '../../../../../../shared/shared.module';

describe('ValuesComponent', () => {
    let component: ValuesComponent;
    let fixture: ComponentFixture<ValuesComponent>;

    const mockedValues = `
    adminserver.image.pullPolicy: IfNotPresent,
    adminserver.image.repository: vmware/harbor-adminserver,
    adminserver.image.tag: dev
    `;
    const fakedAdditionsService = {
        getDetailByLink() {
            return of(mockedValues);
        },
    };
    const mockedLink: AdditionLink = {
        absolute: false,
        href: '/test',
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [
                SharedTestingModule,
                MarkdownModule.forRoot({ sanitize: SecurityContext.HTML }),
            ],
            declarations: [ValuesComponent],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            providers: [
                { provide: AdditionsService, useValue: fakedAdditionsService },
                { provide: MarkedOptions, useValue: {} },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ValuesComponent);
        component = fixture.componentInstance;
        component.valueMode = true;
        component.valuesLink = mockedLink;
        component.ngOnInit();
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should get values  and render', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        const trs = fixture.nativeElement.getElementsByTagName('tr');
        expect(trs.length).toEqual(3);
    });
});
