import { Injectable, signal } from '@angular/core';

@Injectable({
  providedIn: 'root',
})
export class LzLoadVisibilityService {
  public readonly showLoading = signal(false);

  show(): void {
    this.showLoading.set(true);
  }

  hide(): void {
    this.showLoading.set(false);
  }
}
