import { TestBed } from '@angular/core/testing';

import { LzLoadVisibilityService } from './lz-load-visibility.service';

describe('LzLoadVisibilityService', () => {
  let service: LzLoadVisibilityService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(LzLoadVisibilityService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
