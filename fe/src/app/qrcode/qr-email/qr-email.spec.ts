import { ComponentFixture, TestBed } from '@angular/core/testing';

import { QrEmail } from './qr-email';

describe('QrEmail', () => {
  let component: QrEmail;
  let fixture: ComponentFixture<QrEmail>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [QrEmail]
    })
    .compileComponents();

    fixture = TestBed.createComponent(QrEmail);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
