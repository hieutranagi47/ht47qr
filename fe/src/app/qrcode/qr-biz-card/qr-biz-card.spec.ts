import { ComponentFixture, TestBed } from '@angular/core/testing';

import { QrBizCard } from './qr-biz-card';

describe('QrBizCard', () => {
  let component: QrBizCard;
  let fixture: ComponentFixture<QrBizCard>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [QrBizCard]
    })
    .compileComponents();

    fixture = TestBed.createComponent(QrBizCard);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
