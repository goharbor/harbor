import { Directive, OnChanges, Input, SimpleChanges } from '@angular/core';
import { NG_ASYNC_VALIDATORS, Validator, Validators, ValidatorFn, AbstractControl } from '@angular/forms';

import { ProjectService} from '../project/project.service';

import { MemberService } from '../project/member/member.service';
import { Member } from '../project/member/member';

@Directive({
  selector: '[targetExists]',
  providers: [
    ProjectService, MemberService,
    { provide: NG_ASYNC_VALIDATORS, useExisting: TargetExistsValidatorDirective, multi: true},
  ]
})
export class TargetExistsValidatorDirective implements Validator, OnChanges {
  @Input() targetExists: string;
  @Input() projectId: number;

  private valFn = Validators.nullValidator;

  constructor(
    private projectService: ProjectService,
    private memberService: MemberService) {}

  ngOnChanges(changes: SimpleChanges): void {
    const change = changes['targetExists'];
    if (change) {
      const target: string = change.currentValue;
      this.valFn = this.targetExistsValidator(target);
    } else {
      this.valFn = Validators.nullValidator;
    }
  }
  validate(control: AbstractControl): {[key: string]: any} {
    return this.valFn(control);
  }   

  targetExistsValidator(target: string):  ValidatorFn {
    return (control: AbstractControl): {[key: string]: any} => {
      console.log('Target:' + target + ', validate value:' + control.value);
      switch(target) {
      case 'PROJECT_NAME':
        return new Promise(resolve=>{
                this.projectService
                    .checkProjectExists(control.value)
                    .subscribe(res=>resolve({'targetExists': true}),error=>resolve(null));
              });
      case 'MEMBER_NAME':
        return new Promise(resolve=>{
                this.memberService
                    .listMembers(this.projectId, control.value)
                    .subscribe((members: Member[])=>{
                     return members.filter(m=>{
                        if(m.username === control.value) {
                          return true;
                        }
                        return null;
                      }).length > 0 ?
                        resolve({'targetExists': true}) : resolve(null);                   
                    },error=>resolve(null));
              });
      }
    }
  }
}
