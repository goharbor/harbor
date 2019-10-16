import { TestBed, inject } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { MemberService } from './member.service';

describe('MemberService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        HttpClientTestingModule
      ],
      providers: [MemberService]
    });
  });

  it('should be created', inject([MemberService], (service: MemberService) => {
    expect(service).toBeTruthy();
  }));
});
