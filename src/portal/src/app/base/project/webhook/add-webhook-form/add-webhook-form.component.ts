import {
    Component,
    EventEmitter,
    Input,
    OnDestroy,
    OnInit,
    Output,
    ViewChild,
} from '@angular/core';
import { NgForm } from '@angular/forms';
import {
    debounceTime,
    distinctUntilChanged,
    filter,
    finalize,
    switchMap,
} from 'rxjs/operators';
import { ProjectWebhookService } from '../webhook.service';
import { compareValue } from '../../../../shared/units/utils';
import { InlineAlertComponent } from '../../../../shared/components/inline-alert/inline-alert.component';
import { WebhookService } from '../../../../../../ng-swagger-gen/services/webhook.service';
import { WebhookPolicy } from '../../../../../../ng-swagger-gen/models/webhook-policy';
import { Subject, Subscription } from 'rxjs';

@Component({
    selector: 'add-webhook-form',
    templateUrl: './add-webhook-form.component.html',
    styleUrls: ['./add-webhook-form.component.scss'],
})
export class AddWebhookFormComponent implements OnInit, OnDestroy {
    closable: boolean = true;
    checking: boolean = false;
    submitting: boolean = false;
    @Input() projectId: number;
    webhook: WebhookPolicy = {
        enabled: true,
        event_types: [],
        targets: [
            {
                type: 'http',
                address: '',
                skip_cert_verify: true,
            },
        ],
    };
    originValue: WebhookPolicy;
    isModify: boolean;
    @Input() isOpen: boolean;
    // eslint-disable-next-line @angular-eslint/no-output-native
    @Output() close = new EventEmitter<boolean>();
    @ViewChild('webhookForm', { static: true }) currentForm: NgForm;
    @ViewChild(InlineAlertComponent) inlineAlert: InlineAlertComponent;
    @Input()
    metadata: any;
    @Output() notify = new EventEmitter<WebhookPolicy>();
    checkNameOnGoing: boolean = false;
    isNameExisting: boolean = false;
    private _nameSubject = new Subject<string>();
    _nameSubscription: Subscription;
    constructor(
        private webhookService: WebhookService,
        private projectWebhookService: ProjectWebhookService
    ) {}

    ngOnInit() {
        this.subscribeName();
    }
    ngOnDestroy() {
        if (this._nameSubscription) {
            this._nameSubscription.unsubscribe();
            this._nameSubscription = null;
        }
    }
    reset() {
        this.isNameExisting = false;
        this._nameSubject.next('');
    }
    subscribeName() {
        if (!this._nameSubscription) {
            this._nameSubscription = this._nameSubject
                .pipe(
                    debounceTime(500),
                    distinctUntilChanged(),
                    filter(name => {
                        if (
                            this.isModify &&
                            this.originValue &&
                            this.originValue.name === name
                        ) {
                            return false;
                        }
                        return name.length > 0;
                    }),
                    switchMap(name => {
                        this.isNameExisting = false;
                        this.checkNameOnGoing = true;
                        return this.webhookService
                            .ListWebhookPoliciesOfProject({
                                projectNameOrId: this.projectId.toString(),
                                q: encodeURIComponent(`name=${name}`),
                            })
                            .pipe(
                                finalize(() => (this.checkNameOnGoing = false))
                            );
                    })
                )
                .subscribe(res => {
                    if (res && res.length > 0) {
                        this.isNameExisting = true;
                    }
                });
        }
    }
    inputName() {
        this._nameSubject.next(this.webhook.name);
    }
    onCancel() {
        this.reset();
        this.close.emit(false);
        this.currentForm.reset();
        this.inlineAlert.close();
    }

    add() {
        this.submitting = true;
        this.webhookService
            .CreateWebhookPolicyOfProject({
                projectNameOrId: this.projectId.toString(),
                policy: this.webhook,
            })
            .pipe(finalize(() => (this.submitting = false)))
            .subscribe(
                response => {
                    this.reset();
                    this.notify.emit();
                    this.inlineAlert.close();
                },
                error => {
                    this.inlineAlert.showInlineError(error);
                }
            );
    }

    save() {
        this.submitting = true;
        this.webhookService
            .UpdateWebhookPolicyOfProject({
                projectNameOrId: this.projectId.toString(),
                webhookPolicyId: this.webhook.id,
                policy: this.webhook,
            })
            .pipe(finalize(() => (this.submitting = false)))
            .subscribe(
                response => {
                    this.reset();
                    this.inlineAlert.close();
                    this.notify.emit();
                },
                error => {
                    this.inlineAlert.showInlineError(error);
                }
            );
    }

    setCertValue($event: any): void {
        this.webhook.targets[0].skip_cert_verify = !$event;
    }

    public get isValid(): boolean {
        return (
            this.currentForm &&
            this.currentForm.valid &&
            !this.submitting &&
            !this.checking &&
            this.hasEventType()
        );
    }
    hasChange(): boolean {
        return !compareValue(this.originValue, this.webhook);
    }

    setEventType(eventType) {
        if (this.webhook.event_types.indexOf(eventType) === -1) {
            this.webhook.event_types.push(eventType);
        } else {
            this.webhook.event_types.splice(
                this.webhook.event_types.findIndex(item => item === eventType),
                1
            );
        }
    }
    getEventType(eventType): boolean {
        return eventType && this.webhook.event_types.indexOf(eventType) !== -1;
    }
    hasEventType(): boolean {
        return (
            this.metadata &&
            this.metadata.event_type &&
            this.metadata.event_type.length > 0 &&
            this.webhook.event_types &&
            this.webhook.event_types.length > 0
        );
    }
    eventTypeToText(eventType: string): string {
        return this.projectWebhookService.eventTypeToText(eventType);
    }
}
