/*
 * Copyright Project Harbor Authors
 *
 * This product is licensed to you under the Apache License, Version 2.0 (the "License").
 * You may not use this product except in compliance with the License.
 *
 * This product may include a number of subcomponents with separate copyright notices
 * and license terms. Your use of these subcomponents is subject to the terms and
 * conditions of the subcomponent's license, as noted in the LICENSE file.
 */

import { ComponentFixture, TestBed } from '@angular/core/testing';
import { GridViewComponent } from './grid-view.component';
import { SharedTestingModule } from '../../../../shared/shared.module';

describe('GridViewComponent', () => {
    let component: GridViewComponent;
    let fixture: ComponentFixture<GridViewComponent>;
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [GridViewComponent],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(GridViewComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
