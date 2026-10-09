import { provideHttpClient } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';

import { QrContact } from './qr-contact';

describe('QrContact', () => {
  let component: QrContact;
  let fixture: ComponentFixture<QrContact>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      providers: [provideHttpClient()],
      imports: [QrContact],
    }).compileComponents();

    fixture = TestBed.createComponent(QrContact);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
