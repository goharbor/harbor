import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { TopRepoComponent } from './top-repo.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { of } from 'rxjs';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { TopRepoService } from './top-repository.service';
describe('TopRepoComponent', () => {
    let component: TopRepoComponent;
    let fixture: ComponentFixture<TopRepoComponent>;
    const mockMessageHandlerService = {
        showSuccess: () => { },
        handleError: () => { },
        isAppLevel: () => { },
    };
    const mockTopRepoService = {
        getTopRepos: () => of([])
    };
    beforeEach(async(() => {
        TestBed.configureTestingModule({
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            imports: [
                BrowserAnimationsModule,
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule,
                RouterTestingModule,
                NoopAnimationsModule,
                HttpClientTestingModule
            ],
            declarations: [TopRepoComponent],
            providers: [
                TranslateService,
                { provide: TopRepoService, useValue: mockTopRepoService },
                { provide: MessageHandlerService, useValue: mockMessageHandlerService },

            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(TopRepoComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
