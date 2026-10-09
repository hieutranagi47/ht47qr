import { provideHttpClient } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';

import { QrGeo } from './qr-geo';

describe('QrGeo', () => {
  let component: QrGeo;
  let fixture: ComponentFixture<QrGeo>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      providers: [provideHttpClient()],
      imports: [QrGeo],
    }).compileComponents();

    fixture = TestBed.createComponent(QrGeo);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
