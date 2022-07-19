import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TopRepoComponent } from './top-repo.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { of } from 'rxjs';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { TopRepoService } from './top-repository.service';
import { SharedTestingModule } from '../../../shared/shared.module';

describe('TopRepoComponent', () => {
    let component: TopRepoComponent;
    let fixture: ComponentFixture<TopRepoComponent>;
    const mockMessageHandlerService = {
        showSuccess: () => {},
        handleError: () => {},
        isAppLevel: () => {},
    };
    const mockTopRepoService = {
        getTopRepos: () => of([]),
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [TopRepoComponent],
            providers: [
                { provide: TopRepoService, useValue: mockTopRepoService },
                {
                    provide: MessageHandlerService,
                    useValue: mockMessageHandlerService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(TopRepoComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
