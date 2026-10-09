import { Directive, ElementRef, inject } from '@angular/core';

@Directive({
  selector: '[ht47Modal]',
})
export class Modal {
  private readonly el = inject(ElementRef);
}
