import { ComponentFixture, TestBed } from '@angular/core/testing';

import { QrEmail2 } from './qr-email2';

describe('QrEmail2', () => {
  let component: QrEmail2;
  let fixture: ComponentFixture<QrEmail2>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [QrEmail2]
    })
    .compileComponents();

    fixture = TestBed.createComponent(QrEmail2);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
