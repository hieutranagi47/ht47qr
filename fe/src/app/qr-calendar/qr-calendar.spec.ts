import { ComponentFixture, TestBed } from '@angular/core/testing';

import { QrCalendar } from './qr-calendar';

describe('QrCalendar', () => {
  let component: QrCalendar;
  let fixture: ComponentFixture<QrCalendar>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [QrCalendar]
    })
    .compileComponents();

    fixture = TestBed.createComponent(QrCalendar);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
