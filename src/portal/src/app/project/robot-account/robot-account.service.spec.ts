import { TestBed, inject } from '@angular/core/testing';
import { RobotService } from './robot-account.service';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { RobotApiRepository } from "./robot.api.repository";

describe('RobotService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        HttpClientTestingModule
      ],
      providers: [RobotService, RobotApiRepository]
    });
  });

  it('should be created', inject([RobotService], (service: RobotService) => {
    expect(service).toBeTruthy();
  }));
});
