import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ReplicationTasksPageComponent } from './replication-tasks-page.component';

describe('ReplicationTasksPageComponent', () => {
  let component: ReplicationTasksPageComponent;
  let fixture: ComponentFixture<ReplicationTasksPageComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ReplicationTasksPageComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ReplicationTasksPageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
