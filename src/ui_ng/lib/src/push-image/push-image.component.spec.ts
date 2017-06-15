import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpModule } from '@angular/http';
import { DebugElement } from '@angular/core';
import { Observable } from 'rxjs/Observable';

import { PushImageButtonComponent } from './push-image.component';
import { CopyInputComponent } from './copy-input.component';
import { InlineAlertComponent } from '../inline-alert/inline-alert.component';

import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { SharedModule } from '../shared/shared.module';
import { click } from '../utils';

describe('PushImageButtonComponent (inline template)', () => {
  let component: PushImageButtonComponent;
  let fixture: ComponentFixture<PushImageButtonComponent>;
  let serviceConfig: IServiceConfig;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      declarations: [InlineAlertComponent, CopyInputComponent, PushImageButtonComponent],
      providers: [
        { provide: SERVICE_CONFIG, useValue: {} }
      ]
    });

  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(PushImageButtonComponent);
    component = fixture.componentInstance;
    component.projectName = "testing";
    component.registryUrl = "https://testing.harbor.com"
    serviceConfig = TestBed.get(SERVICE_CONFIG);

    fixture.detectChanges();
  });

  it('should be created', () => {
    expect(component).toBeTruthy();
  });

  it('should open the drop-down panel', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      let el: HTMLElement = fixture.nativeElement.querySelector('button');
      expect(el).not.toBeNull();
      el.click();

      fixture.detectChanges();
      let copyInputs: HTMLInputElement[] = fixture.nativeElement.querySelectorAll('.command-input');
      expect(copyInputs.length).toEqual(2);

      expect(copyInputs[0].value.trim()).toEqual(`docker tag SOURCE_IMAGE[:TAG] ${component.registryUrl}/${component.projectName}/IMAGE[:TAG]`);
      expect(copyInputs[1].value.trim()).toEqual(`docker push ${component.registryUrl}/${component.projectName}/IMAGE[:TAG]`);
    });
  }));

});
