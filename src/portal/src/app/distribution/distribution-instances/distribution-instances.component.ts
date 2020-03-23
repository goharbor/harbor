import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { Component, OnInit, ViewChild, OnDestroy } from '@angular/core';
import {
  Subscription,
  Observable,
  forkJoin,
  throwError as observableThrowError
} from 'rxjs';
import { MsgChannelService } from '../msg-channel.service';
import { DistributionSetupModalComponent } from '../distribution-setup-modal/distribution-setup-modal.component';
import { OperationService } from '../../../lib/components/operation/operation.service';
import {
  ConfirmationState,
  ConfirmationTargets,
  ConfirmationButtons
} from '../../shared/shared.const';
import { ConfirmationDialogService } from '../../shared/confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../../shared/confirmation-dialog/confirmation-message';
import {
  operateChanges,
  OperateInfo,
  OperationState
} from '../../../lib/components/operation/operate';
import { TranslateService } from '@ngx-translate/core';
import { map, catchError, finalize } from 'rxjs/operators';
import { errorHandler } from '../../../lib/utils/shared/shared.utils';
import { DEFAULT_PAGE_SIZE } from '../../../lib/utils/utils';
import { Instance } from "../../../../ng-swagger-gen/models/instance";
import { PreheatService } from "../../../../ng-swagger-gen/services/preheat.service";

interface MultiOperateData {
  operation: string;
  instances: Instance[];
}

@Component({
  selector: 'dist-instances',
  templateUrl: './distribution-instances.component.html',
  styleUrls: ['./distribution-instances.component.scss']
})
export class DistributionInstancesComponent implements OnInit, OnDestroy {
  instances: Instance[] = [];
  selectedRow: Instance[] = [];

  pageSize: number = DEFAULT_PAGE_SIZE;
  currentPage: number = 1;
  totalCount: number = 0;
  queryString: string;

  chanSub: Subscription;

  private loading: boolean = true;
  private operationSubscription: Subscription;

  @ViewChild('setupModal', { static: false })
  setupModal: DistributionSetupModalComponent;

  constructor(
    private disService: PreheatService,
    private msgHandler: MessageHandlerService,
    private chanService: MsgChannelService,

    private translate: TranslateService,
    private operationDialogService: ConfirmationDialogService,
    private operationService: OperationService
  ) {
    // subscribe operation
    this.operationSubscription = operationDialogService.confirmationConfirm$.subscribe(
      confirmed => {
        if (
          confirmed &&
          confirmed.source === ConfirmationTargets.INSTANCE &&
          confirmed.state === ConfirmationState.CONFIRMED
        ) {
          this.operateInstance(confirmed.data);
        }
      }
    );
  }

  public get inProgress(): boolean {
    return this.loading;
  }

  ngOnInit() {
    this.loadData();
    this.chanSub = this.chanService.subscribe((msg: string) => {
      if (msg === 'created' || msg === 'updated' || 'delete') {
        this.loadData();
        return;
      }

      console.error('unknown msg', msg);
    });
  }

  ngOnDestroy() {
    if (this.operationSubscription) {
      this.operationSubscription.unsubscribe();
    }
    if (this.chanSub) {
      this.chanSub.unsubscribe();
    }
  }

  loadData() {
    this.selectedRow = [];
    const queryParam: PreheatService.ListInstancesParams = {
      page: this.currentPage,
      pageSize: this.pageSize
    };
    if (this.queryString) {
      queryParam.q = encodeURIComponent(`name=~${this.queryString}`);
    }
    this.loading = true;
    this.disService.ListInstancesResponse(queryParam)
      .pipe(finalize(() => this.loading = false))
      .subscribe(
      response => {
        this.totalCount = Number.parseInt(
          response.headers.get('x-total-count')
        );
        this.instances = response.body as Instance[];
      },
      err => this.msgHandler.error(err)
    );
  }

  refresh() {
    this.currentPage = 1;
    this.loadData();
  }

  doFilter($evt: any) {
    this.currentPage = 1;
    this.queryString = $evt;
    this.loadData();
  }

  addInstance() {
    this.setupModal.openSetupModal(false);
  }

  editInstance(instance: Instance) {
    this.setupModal.openSetupModal(true, instance);
  }

  // Operate the specified Instance
  operateInstances(operation: string, instances: Instance[]): void {
    let arr: string[] = [];
    let title: string;
    let summary: string;
    let buttons: ConfirmationButtons;

    switch (operation) {
      case 'delete':
        title = 'DISTRIBUTION.DELETION_TITLE';
        summary = 'DISTRIBUTION.DELETION_SUMMARY';
        buttons = ConfirmationButtons.DELETE_CANCEL;
        break;
      case 'enable':
        title = 'DISTRIBUTION.ENABLE_TITLE';
        summary = 'DISTRIBUTION.ENABLE_SUMMARY';
        buttons = ConfirmationButtons.ENABLE_CANCEL;
        break;
      case 'disable':
        title = 'DISTRIBUTION.DISABLE_TITLE';
        summary = 'DISTRIBUTION.DISABLE_SUMMARY';
        buttons = ConfirmationButtons.DISABLE_CANCEL;
        break;

      default:
        return;
    }

    if (instances && instances.length) {
      instances.forEach(instance => {
        arr.push(instance.name);
      });
    }
    // Confirm
    let msg: ConfirmationMessage = new ConfirmationMessage(
      title,
      summary,
      arr.join(','),
      { operation: operation, instances: instances },
      ConfirmationTargets.INSTANCE,
      buttons
    );
    this.operationDialogService.openComfirmDialog(msg);
  }

  operateInstance(data: MultiOperateData) {
    let observableLists: any[] = [];
    if (data.instances && data.instances.length) {
      switch (data.operation) {
        case 'delete':
          data.instances.forEach(instance => {
            observableLists.push(this.deleteInstance(instance));
          });
          break;

        case 'enable':
          data.instances.forEach(instance => {
            observableLists.push(this.enableInstance(instance));
          });
          break;

        case 'disable':
          data.instances.forEach(instance => {
            observableLists.push(this.disableInstance(instance));
          });
          break;
      }

      forkJoin(...observableLists).subscribe(item => {
        this.selectedRow = [];
        this.refresh();
      });
    }
  }
  deleteInstance(instance: Instance): Observable<any> {
    let operMessage = new OperateInfo();
    operMessage.name = 'OPERATION.DELETE_INSTANCE';
    operMessage.data.id = instance.id;
    operMessage.state = OperationState.progressing;
    operMessage.data.name = instance.name;
    this.operationService.publishInfo(operMessage);

    return this.disService.DeleteInstance({instanceId: instance.id}).pipe(
      map(() => {
        this.translate.get('DISTRIBUTION.DELETED_SUCCESS').subscribe(msg => {
          operateChanges(operMessage, OperationState.success);
          this.msgHandler.info(msg);
        });
      }),
      catchError(error => {
        const message = errorHandler(error);
        this.translate.get('DISTRIBUTION.DELETED_FAILED').subscribe(msg => {
          operateChanges(operMessage, OperationState.failure, msg);
          this.translate.get(message).subscribe(errMsg => {
            this.msgHandler.error(msg + ': ' + errMsg);
          });
        });
        return observableThrowError(message);
      })
    );
  }

  enableInstance(instance: Instance) {
    let operMessage = new OperateInfo();
    operMessage.name = 'OPERATION.ENABLE_INSTANCE';
    operMessage.data.id = instance.id;
    operMessage.state = OperationState.progressing;
    operMessage.data.name = instance.name;
    this.operationService.publishInfo(operMessage);

    instance.enabled = true;
    return this.disService
      .UpdateInstance({
        propertySet: {enabled: true},
        instanceId: instance.id
      })
      .pipe(
        map(() => {
          this.translate.get('DISTRIBUTION.ENABLE_SUCCESS').subscribe(msg => {
            operateChanges(operMessage, OperationState.success);
            this.msgHandler.info(msg);
          });
        }),
        catchError(error => {
          const message = errorHandler(error);
          this.translate.get('DISTRIBUTION.ENABLE_FAILED').subscribe(msg => {
            operateChanges(operMessage, OperationState.failure, msg);
            this.translate.get(message).subscribe(errMsg => {
              this.msgHandler.error(msg + ': ' + errMsg);
            });
          });
          return observableThrowError(message);
        })
      );
  }

  disableInstance(instance: Instance) {
    let operMessage = new OperateInfo();
    operMessage.name = 'OPERATION.DISABLE_INSTANCE';
    operMessage.data.id = instance.id;
    operMessage.state = OperationState.progressing;
    operMessage.data.name = instance.name;
    this.operationService.publishInfo(operMessage);

    instance.enabled = false;
    return this.disService
      .UpdateInstance({
        propertySet: {enabled: false},
        instanceId: instance.id
      })
      .pipe(
        map(() => {
          this.translate.get('DISTRIBUTION.DISABLE_SUCCESS').subscribe(msg => {
            operateChanges(operMessage, OperationState.success);
            this.msgHandler.info(msg);
          });
        }),
        catchError(error => {
          const message = errorHandler(error);
          this.translate.get('DISTRIBUTION.DISABLE_FAILED').subscribe(msg => {
            operateChanges(operMessage, OperationState.failure, msg);
            this.translate.get(message).subscribe(errMsg => {
              this.msgHandler.error(msg + ': ' + errMsg);
            });
          });
          return observableThrowError(message);
        })
      );
  }

  fmtTime(time: number) {
    let date = new Date();
    return date.setTime(time * 1000);
  }
}
