/**
 * @ Author: Hieu Tran
 * @ Create Time: 2023-11-25 23:09:56
 * @ Modified by: Hieu Tran
 * @ Modify Time: 2023-11-25 23:34:06
 * @ Description: Set the real view height, include mobile view with auto address bar hidding.
 * @ Modified time: 2023-11-26 17:34:05
 */

import { Directive, OnDestroy, afterNextRender } from '@angular/core';
import { isBrowser } from '@app/shared/utils/running-env';

@Directive({
  standalone: true,
  selector: '[ht47MobileView]',
})
export class MobileViewDirective implements OnDestroy {
  protected screenChange: EventListener | undefined = undefined;

  constructor() {
    afterNextRender(() => {
      // Store the event change method into the directive scope to destroy it later.
      this.screenChange = this.resetMobileViewHeight.bind(this);

      // the method will be run at the first load
      this.resetMobileViewHeight();

      // trigger the resetMobileViewHeight when resize or orientationchange
      window.addEventListener('resize', this.screenChange, true);
      window.addEventListener('orientationchange', this.screenChange, true);
    });
  }

  ngOnDestroy(): void {
    if (isBrowser && this.screenChange) {
      window.removeEventListener('resize', this.screenChange, true);
      window.removeEventListener('orientationchange', this.screenChange, true);
    }
  }

  resetMobileViewHeight(): void {
    document.documentElement.style.setProperty(
      '--mobile-vh',
      `${window.innerHeight}px`
    );
  }
}
