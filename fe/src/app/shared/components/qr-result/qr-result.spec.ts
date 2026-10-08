import { ComponentFixture, TestBed } from '@angular/core/testing';

import { QrResult } from './qr-result';

describe('QrResult', () => {
  let component: QrResult;
  let fixture: ComponentFixture<QrResult>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [QrResult]
    })
    .compileComponents();

    fixture = TestBed.createComponent(QrResult);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
