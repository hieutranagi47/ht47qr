import { ComponentFixture, TestBed } from '@angular/core/testing';

import { QrTel } from './qr-tel';

describe('QrTel', () => {
  let component: QrTel;
  let fixture: ComponentFixture<QrTel>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [QrTel]
    })
    .compileComponents();

    fixture = TestBed.createComponent(QrTel);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
