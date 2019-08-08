import {
  Component,
  EventEmitter,
  Output,
  ViewChild,
  OnInit,
} from '@angular/core';
import { NgForm, Validators } from '@angular/forms';

import { InlineAlertComponent } from '../../../inline-alert/inline-alert.component';

import { QuotaUnits, QuotaUnlimited, QUOTA_DANGER_COEFFICIENT, QUOTA_WARNING_COEFFICIENT } from "../../../shared/shared.const";

import { clone, getSuitableUnit, getByte, GetIntegerAndUnit, validateCountLimit, validateLimit } from '../../../utils';
import { EditQuotaQuotaInterface, QuotaHardLimitInterface } from '../../../service';
import { distinctUntilChanged } from 'rxjs/operators';

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
    storageLimit: -1
    , storageUnit: ''
    , countLimit: -1
  };
  quotaUnits = QuotaUnits;
  staticBackdrop = true;
  closable = false;
  quotaForm: NgForm;
  @ViewChild(InlineAlertComponent, {static: false})
  inlineAlert: InlineAlertComponent;

  @ViewChild('quotaForm', {static: false})
  currentForm: NgForm;
  @Output() confirmAction = new EventEmitter();
  quotaDangerCoefficient: number = QUOTA_DANGER_COEFFICIENT;
  quotaWarningCoefficient: number = QUOTA_WARNING_COEFFICIENT;
  constructor() { }

  ngOnInit() {
  }

  onSubmit(): void {
    const emitData = {
      formValue: this.currentForm.value,
      isSystemDefaultQuota: this.defaultTextsObj.isSystemDefaultQuota,
      id: this.quotaHardLimitValue.id
    };
    this.confirmAction.emit(emitData);
  }
  onCancel() {
    this.openEditQuota = false;
  }

  openEditQuotaModal(defaultTextsObj: EditQuotaQuotaInterface): void {
    this.defaultTextsObj = defaultTextsObj;
    if (this.defaultTextsObj.isSystemDefaultQuota) {
      this.quotaHardLimitValue = {
        storageLimit: defaultTextsObj.quotaHardLimitValue.storageLimit === QuotaUnlimited ?
          QuotaUnlimited : GetIntegerAndUnit(defaultTextsObj.quotaHardLimitValue.storageLimit
            , clone(QuotaUnits), 0, clone(QuotaUnits)).partNumberHard
        , storageUnit: defaultTextsObj.quotaHardLimitValue.storageLimit === QuotaUnlimited ?
          QuotaUnits[3].UNIT : GetIntegerAndUnit(defaultTextsObj.quotaHardLimitValue.storageLimit
            , clone(QuotaUnits), 0, clone(QuotaUnits)).partCharacterHard
        , countLimit: defaultTextsObj.quotaHardLimitValue.countLimit
      };
    } else {
      this.quotaHardLimitValue = {
        storageLimit: defaultTextsObj.quotaHardLimitValue.hard.storage === QuotaUnlimited ?
          QuotaUnlimited : GetIntegerAndUnit(defaultTextsObj.quotaHardLimitValue.hard.storage
            , clone(QuotaUnits), defaultTextsObj.quotaHardLimitValue.used.storage, clone(QuotaUnits)).partNumberHard
        , storageUnit: defaultTextsObj.quotaHardLimitValue.hard.storage === QuotaUnlimited ?
          QuotaUnits[3].UNIT : GetIntegerAndUnit(defaultTextsObj.quotaHardLimitValue.hard.storage
            , clone(QuotaUnits), defaultTextsObj.quotaHardLimitValue.used.storage, clone(QuotaUnits)).partCharacterHard
        , countLimit: defaultTextsObj.quotaHardLimitValue.hard.count
        , id: defaultTextsObj.quotaHardLimitValue.id
        , countUsed: defaultTextsObj.quotaHardLimitValue.used.count
        , storageUsed: defaultTextsObj.quotaHardLimitValue.used.storage
      };
    }
    let defaultForm = {
      count: this.quotaHardLimitValue.countLimit
      , storage: this.quotaHardLimitValue.storageLimit
      , storageUnit: this.quotaHardLimitValue.storageUnit
    };
    this.currentForm.resetForm(defaultForm);
    this.openEditQuota = true;

    this.currentForm.form.controls['storage'].setValidators(
      [
        Validators.required,
        Validators.pattern('(^-1$)|(^([1-9]+)([0-9]+)*$)'),
        validateLimit(this.currentForm.form.controls['storageUnit'])
      ]);
    this.currentForm.form.controls['count'].setValidators(
      [
        Validators.required,
        Validators.pattern('(^-1$)|(^([1-9]+)([0-9]+)*$)'),
        validateCountLimit()
      ]);
    this.currentForm.form.valueChanges
      .pipe(distinctUntilChanged((a, b) => JSON.stringify(a) === JSON.stringify(b)))
      .subscribe((data) => {
        ['storage', 'storageUnit', 'count'].forEach(fieldName => {
          if (this.currentForm.form.get(fieldName) && this.currentForm.form.get(fieldName).value !== null) {
            this.currentForm.form.get(fieldName).updateValueAndValidity();
          }
        });
      });
  }

  get isValid() {
    return this.currentForm.valid && this.currentForm.dirty;
  }
  getSuitableUnit(value) {
    const QuotaUnitsCopy = clone(QuotaUnits);
    return getSuitableUnit(value, QuotaUnitsCopy);
  }
  getIntegerAndUnit(valueHard, valueUsed) {
    return GetIntegerAndUnit(valueHard
      , clone(QuotaUnits), valueUsed, clone(QuotaUnits));
  }
  getByte(count: number, unit: string) {
    if (+count === +count) {
      return getByte(+count, unit);
    }
    return 0;
  }
  isDangerColor(limit: number | string, used: number | string, unit?: string) {
    if (unit) {
      return limit !== QuotaUnlimited ? +used / getByte(+limit, unit) >= this.quotaDangerCoefficient : false;
    }
    return limit !== QuotaUnlimited ? +used / +limit >= this.quotaDangerCoefficient : false;
  }
  isWarningColor(limit: number | string, used: number | string, unit?: string) {
    if (unit) {
      return limit !== QuotaUnlimited ?
      +used / getByte(+limit, unit) >= this.quotaWarningCoefficient && +used / getByte(+limit, unit) <= this.quotaDangerCoefficient : false;
    }
    return limit !== QuotaUnlimited ?
    +used / +limit >= this.quotaWarningCoefficient && +used / +limit <= this.quotaDangerCoefficient : false;
  }
}
