import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { Component, Input, OnInit, OnDestroy, ElementRef } from '@angular/core';
import { Router } from '@angular/router';
import { Subscription } from "rxjs";
import { RouterTestingModule } from '@angular/router/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ClarityModule } from "@clr/angular";
import { Message } from './message';
import { MessageService } from './message.service';
import { MessageComponent } from './message.component';

describe('MessageComponent', () => {
    let component: MessageComponent;
    let fixture: ComponentFixture<MessageComponent>;
    let fakeElementRef = null;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                ClarityModule,
                RouterTestingModule,
                TranslateModule.forRoot()
            ],
            declarations: [MessageComponent],
            providers: [
                MessageService,
                TranslateService,
                {provide: ElementRef, useValue: fakeElementRef}
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(MessageComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
