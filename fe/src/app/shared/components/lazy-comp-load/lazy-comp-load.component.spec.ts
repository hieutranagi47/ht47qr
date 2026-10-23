import { ComponentFixture, TestBed } from '@angular/core/testing';

import { LazyCompLoadComponent } from './lazy-comp-load.component';

describe('LazyCompLoadComponent', () => {
  let component: LazyCompLoadComponent;
  let fixture: ComponentFixture<LazyCompLoadComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [LazyCompLoadComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(LazyCompLoadComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
