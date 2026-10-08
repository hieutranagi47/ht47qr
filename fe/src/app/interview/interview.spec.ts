import { ComponentFixture, TestBed } from '@angular/core/testing';

import { Interview } from './interview';

describe('Interview', () => {
  let component: Interview;
  let fixture: ComponentFixture<Interview>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [Interview]
    })
    .compileComponents();

    fixture = TestBed.createComponent(Interview);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
