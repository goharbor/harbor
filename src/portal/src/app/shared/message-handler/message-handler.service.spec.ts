import { TestBed, inject } from '@angular/core/testing';

import { MessageHandlerService } from './message-handler.service';

xdescribe('MessageHandlerService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [MessageHandlerService]
    });
  });

  it('should be created', inject([MessageHandlerService], (service: MessageHandlerService) => {
    expect(service).toBeTruthy();
  }));
});
