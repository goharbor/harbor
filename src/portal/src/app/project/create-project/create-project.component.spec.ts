import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CreateProjectComponent } from './create-project.component';
import { InlineAlertComponent } from '../../shared/inline-alert/inline-alert.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { ProjectService } from "../../../lib/services";
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { ErrorHandler } from '../../../lib/utils/error-handler';

describe('CreateProjectComponent', () => {
    let component: CreateProjectComponent;
    let fixture: ComponentFixture<CreateProjectComponent>;
    const mockProjectService = {
        checkProjectExists: function() {
        },
        createProject: function () {
        }
    };
    const mockMessageHandlerService = {
        showSuccess: function() {
        }
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                FormsModule,
                ClarityModule,
                TranslateModule.forRoot()
            ],
            declarations: [CreateProjectComponent, InlineAlertComponent],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            providers: [
                TranslateService,
                {provide: ProjectService, useValue: mockProjectService},
                {provide: MessageHandlerService, useValue: mockMessageHandlerService},
                ErrorHandler
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(CreateProjectComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
