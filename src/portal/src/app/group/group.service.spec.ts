import { TestBed, inject } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { GroupService } from './group.service';

describe('GroupService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [GroupService],
      imports: [
        HttpClientTestingModule
      ]
    });
  });

  it('should be created', inject([GroupService], (service: GroupService) => {
    expect(service).toBeTruthy();
  }));
});
