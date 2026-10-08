import { ComponentFixture, TestBed } from '@angular/core/testing';

import { QrText } from './qr-text';

describe('QrText', () => {
  let component: QrText;
  let fixture: ComponentFixture<QrText>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [QrText]
    })
    .compileComponents();

    fixture = TestBed.createComponent(QrText);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
