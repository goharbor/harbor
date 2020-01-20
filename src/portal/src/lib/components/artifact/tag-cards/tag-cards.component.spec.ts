import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { TagCardsComponent } from './tag-cards.component';

describe('TagCardsComponent', () => {
  let component: TagCardsComponent;
  let fixture: ComponentFixture<TagCardsComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ TagCardsComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(TagCardsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
