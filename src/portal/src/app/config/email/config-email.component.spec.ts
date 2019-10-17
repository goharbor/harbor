import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { ConfirmMessageHandler } from '../config.msg.utils';
import { ConfigurationService } from '../config.service';
import { ConfigurationEmailComponent } from './config-email.component';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormsModule } from '@angular/forms';

describe('ConfigurationEmailComponent', () => {
    let component: ConfigurationEmailComponent;
    let fixture: ComponentFixture<ConfigurationEmailComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot(),
                FormsModule
            ],
            declarations: [ConfigurationEmailComponent],
            providers: [
                { provide: MessageHandlerService, useValue: null },
                TranslateService,
                { provide: ConfirmMessageHandler, useValue: null },
                { provide: ConfigurationService, useValue: null }
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ConfigurationEmailComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
