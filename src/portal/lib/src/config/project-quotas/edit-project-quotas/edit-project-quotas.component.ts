import {
  Component,
  EventEmitter,
  Output,
  ViewChild,
  OnInit,
} from '@angular/core';
import { NgForm } from '@angular/forms';
import { ActivatedRoute } from "@angular/router";



import { TranslateService } from '@ngx-translate/core';

import { InlineAlertComponent } from '../../../inline-alert/inline-alert.component';

import { QuotaUnits } from "../../../shared/shared.const";

import { clone, getSuitableUnit, getByte, SeparationNumberCharacter } from '../../../utils';
import { EditQuotaQuotaInterface, QuotaHardLimitInterface } from '../../../service';

@Component({
  selector: 'edit-project-quotas',
  templateUrl: './edit-project-quotas.component.html',
  styleUrls: ['./edit-project-quotas.component.scss']
})
export class EditProjectQuotasComponent implements OnInit {
  openEditQuota: boolean;
  defaultTextsObj: { editQuota: string; setQuota: string; countQuota: string; storageQuota: string; isSystemDefaultQuota: boolean } = {
    editQuota: '',
    setQuota: '',
    countQuota: '',
    storageQuota: '',
    isSystemDefaultQuota: false,
  };
  quotaHardLimitValue: QuotaHardLimitInterface = {
    storageLimit: ''
    , storageUnit: ''
    , countLimit: ''
  };
  quotaUnits = QuotaUnits;
  staticBackdrop = true;
  closable = false;
  quotaForm: NgForm;
  @ViewChild(InlineAlertComponent)
  inlineAlert: InlineAlertComponent;

  @ViewChild('quotaForm')
  currentForm: NgForm;
  @Output() confirmAction = new EventEmitter();
  constructor(
    private translateService: TranslateService,
    private route: ActivatedRoute) { }

  ngOnInit() {
  }

  onSubmit(): void {
    const emitData = {
      formValue: this.currentForm.value,
      isSystemDefaultQuota: this.defaultTextsObj.isSystemDefaultQuota,
      id: this.quotaHardLimitValue.id
    }
    this.confirmAction.emit(emitData);
  }
  onCancel() {
    this.openEditQuota = false;
  }

  openEditQuotaModal(defaultTextsObj: EditQuotaQuotaInterface): void {
    this.currentForm.form.reset();
    setTimeout(() => {
      this.defaultTextsObj = defaultTextsObj;
      if (this.defaultTextsObj.isSystemDefaultQuota) {
        this.quotaHardLimitValue = {
          storageLimit: defaultTextsObj.quotaHardLimitValue.storageLimit
          , storageUnit: defaultTextsObj.quotaHardLimitValue.storageUnit
          , countLimit: defaultTextsObj.quotaHardLimitValue.countLimit
        };
      } else {

        this.quotaHardLimitValue = {
          storageLimit: SeparationNumberCharacter(defaultTextsObj.quotaHardLimitValue.spec.storage, QuotaUnits[3].UNIT).numberStr
          , storageUnit: SeparationNumberCharacter(defaultTextsObj.quotaHardLimitValue.spec.storage, QuotaUnits[3].UNIT).character
          , countLimit: defaultTextsObj.quotaHardLimitValue.status.hard.count
          , id: defaultTextsObj.quotaHardLimitValue.id
          , countUsed: defaultTextsObj.quotaHardLimitValue.status.used.count
          , storageUsed: defaultTextsObj.quotaHardLimitValue.status.used.storage
        };
      }

      this.openEditQuota = true;
    }, 100)

  }
  get isValid() {
    return this.currentForm.valid && this.currentForm.dirty;
  }
  getSuitableUnit(value) {
    const QuotaUnitsCopy = clone(QuotaUnits);
    return getSuitableUnit(value, QuotaUnitsCopy)
  }
  getByte(countString: string, unit: string) {
    if (+countString === +countString) {
      return getByte(+countString, unit);
    }
    return 0
  }
  getDangerStyle(limit: string, used: string, unit?: string) {
    if (unit) {
      return limit !== '-1' ? +used / getByte(+limit, unit) > 0.9 : false;
    }
    return limit !== '-1' ? +used / +limit > 0.9 : false;
  }
}
