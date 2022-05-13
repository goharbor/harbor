import { TestBed, inject } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { MessageHandlerService } from './message-handler.service';
import { MessageService } from '../components/global-message/message.service';
import { SessionService } from './session.service';
import { UserPermissionService } from './index';

describe('MessageHandlerService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [TranslateModule.forRoot()],
            providers: [
                MessageHandlerService,
                TranslateService,
                { provide: SessionService, useValue: null },
                { provide: UserPermissionService, useValue: null },
                { provide: MessageService, useValue: null },
            ],
        });
    });

    it('should be created', inject(
        [MessageHandlerService],
        (service: MessageHandlerService) => {
            expect(service).toBeTruthy();
        }
    ));
});
