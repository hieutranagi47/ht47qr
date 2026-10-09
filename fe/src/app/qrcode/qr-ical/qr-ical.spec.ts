import { ComponentFixture, TestBed } from '@angular/core/testing';

import { QrIcal } from './qr-ical';

describe('QrIcal', () => {
  let component: QrIcal;
  let fixture: ComponentFixture<QrIcal>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [QrIcal]
    })
    .compileComponents();

    fixture = TestBed.createComponent(QrIcal);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
