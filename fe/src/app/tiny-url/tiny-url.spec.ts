import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TinyUrl } from './tiny-url';

describe('TinyUrl', () => {
  let component: TinyUrl;
  let fixture: ComponentFixture<TinyUrl>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TinyUrl],
    }).compileComponents();

    fixture = TestBed.createComponent(TinyUrl);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
