import { TestBed } from '@angular/core/testing';

import { BioairService } from './bioair.service';

describe('BioairService', () => {
  let service: BioairService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(BioairService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
