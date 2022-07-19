import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of } from 'rxjs';
import { TranslateService } from '@ngx-translate/core';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { ScannerComponent } from './scanner.component';
import { SharedTestingModule } from '../../../shared/shared.module';
import { ActivatedRoute } from '@angular/router';
import { Scanner } from '../../left-side-nav/interrogation-services/scanner/scanner';
import { ErrorHandler } from '../../../shared/units/error-handler';
import { ProjectService } from '../../../../../ng-swagger-gen/services/project.service';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { Registry } from '../../../../../ng-swagger-gen/models/registry';

describe('ScannerComponent', () => {
    const mockScanner1: Scanner = {
        uuid: 'abc',
        name: 'test1',
        description: 'just a sample',
        version: '1.0.0',
        url: 'http://168.0.0.1',
        health: 'healthy',
    };
    const mockScanner2: Scanner = {
        uuid: 'def',
        name: 'test2',
        description: 'just a sample',
        version: '2.0.0',
        url: 'http://168.0.0.2',
        health: 'healthy',
    };
    let component: ScannerComponent;
    let fixture: ComponentFixture<ScannerComponent>;
    let fakedProjectService = {
        getScannerOfProject() {
            return of(mockScanner1);
        },
        listScannerCandidatesOfProject() {
            return of([mockScanner1, mockScanner2]);
        },
        listScannerCandidatesOfProjectResponse() {
            const response: HttpResponse<Array<Registry>> = new HttpResponse<
                Array<Registry>
            >({
                headers: new HttpHeaders({
                    'x-total-count': [
                        mockScanner1,
                        mockScanner2,
                    ].length.toString(),
                }),
                body: [mockScanner1, mockScanner2],
            });
            return of(response);
        },
        setScannerOfProject() {
            return of(true);
        },
    };
    let fakedRoute = {
        snapshot: {
            parent: {
                parent: {
                    params: {
                        id: 1,
                    },
                },
            },
        },
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ScannerComponent],
            providers: [
                TranslateService,
                MessageHandlerService,
                ErrorHandler,
                { provide: ActivatedRoute, useValue: fakedRoute },
                { provide: ProjectService, useValue: fakedProjectService },
            ],
        }).compileComponents();
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(ScannerComponent);
        component = fixture.componentInstance;
        spyOn(component, 'getPermission').and.returnValue(undefined);
        fixture.detectChanges();
    });
    it('should creat', () => {
        expect(component).toBeTruthy();
    });
    it('should get scanner and render', () => {
        component.hasCreatePermission = true;
        let el: HTMLElement =
            fixture.nativeElement.querySelector('#scanner-name');
        expect(el.textContent.trim()).toEqual('test1');
    });
    it('select another scanner', () => {
        component.hasCreatePermission = true;
        component.getScanners();
        fixture.detectChanges();
        const editButton = fixture.nativeElement.querySelector('#edit-scanner');
        expect(editButton).toBeTruthy();
        editButton.click();
        fixture.detectChanges();
        component.selectedScanner = mockScanner2;
        fixture.detectChanges();
        const saveButton = fixture.nativeElement.querySelector('#save-scanner');
        saveButton.click();
        fixture.detectChanges();
        expect(component.opened).toBeFalsy();
    });
});
