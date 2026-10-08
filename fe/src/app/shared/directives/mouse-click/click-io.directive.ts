import { DOCUMENT } from '@angular/common';
import {
  Directive,
  ElementRef,
  HostListener,
  OnDestroy,
  OnInit,
  Renderer2,
  inject,
  input,
  output,
} from '@angular/core';

@Directive({
  standalone: true,
  selector: '[ht47Clickio]',
})
export class ClickIODirective implements OnInit, OnDestroy {
  toggleClass = input<string>('toggle-me');
  clickOut = output<MouseEvent>();

  private document: Document = inject(DOCUMENT);
  private elmRef = inject(ElementRef);
  private renderer = inject(Renderer2);
  onWindow: boolean = false;
  onElement: boolean = false;

  windowMouseDownHandler = (ev: MouseEvent) => {
    this.onWindow = true;
  };

  windowMouseUpHandler = (ev: MouseEvent) => {
    if (this.onWindow && this.onElement) {
      this.renderer.addClass(this.elmRef.nativeElement, this.toggleClass());
    } else {
      if (this.clickOut) {
        this.clickOut.emit(ev);
      }
      if (this.elmRef.nativeElement.className.indexOf(this.toggleClass) !== -1) {
        this.renderer.removeClass(this.elmRef.nativeElement, this.toggleClass());
      }
      this.document.removeEventListener('mousedown', this.windowMouseDownHandler);
      this.document.removeEventListener('mouseup', this.windowMouseUpHandler);
    }
    this.onWindow = false;
    this.onElement = false;
  };

  /** ELEMENT MOUSE DOWN */
  @HostListener('mousedown', ['$event'])
  elementMouseDown(e: MouseEvent) {
    this.onElement = true;
    /**Because we cannot terminate HostListener,
     * so I have to use pure javascript to control window event listener
     * */
    this.document.addEventListener('mousedown', this.windowMouseDownHandler);
    this.document.addEventListener('mouseup', this.windowMouseUpHandler);
  }

  ngOnInit(): void {
    this.onWindow = false;
    this.onElement = false;
    this.toggleClass = this.toggleClass || 'clicked-in';
  }

  ngOnDestroy(): void {
    this.document.removeEventListener('mousedown', this.windowMouseDownHandler);
    this.document.removeEventListener('mouseup', this.windowMouseUpHandler);
  }
}
