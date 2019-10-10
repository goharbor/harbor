import { TestBed, inject } from '@angular/core/testing';

import { MemberService } from './member.service';

xdescribe('MemberService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [MemberService]
    });
  });

  it('should be created', inject([MemberService], (service: MemberService) => {
    expect(service).toBeTruthy();
  }));
});
