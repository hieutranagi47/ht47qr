import { ComponentFixture, TestBed } from '@angular/core/testing';

import { QrSms } from './qr-sms';

describe('QrSms', () => {
  let component: QrSms;
  let fixture: ComponentFixture<QrSms>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [QrSms]
    })
    .compileComponents();

    fixture = TestBed.createComponent(QrSms);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
