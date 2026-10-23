
import { Component, OnInit, inject, ChangeDetectionStrategy } from '@angular/core';
import {
  NavigationCancel,
  NavigationEnd,
  RouteConfigLoadEnd,
  RouteConfigLoadStart,
  Router,
} from '@angular/router';
import { LzLoadVisibilityService } from './lz-load-visibility.service';

@Component({
    selector: 'ht47-lazy-comp-load',
    imports: [],
    templateUrl: './lazy-comp-load.component.html',
    changeDetection: ChangeDetectionStrategy.Eager,
    styleUrl: './lazy-comp-load.component.scss'
})
export class LazyCompLoadComponent implements OnInit {
  lzLoadingV = inject(LzLoadVisibilityService);
  private router = inject(Router);
  isShow = this.lzLoadingV.showLoading;

  ngOnInit(): void {
    this.router.events.subscribe((event: any) => {
      if (event instanceof RouteConfigLoadStart) {
        this.lzLoadingV.show();
      } else if (
        event instanceof RouteConfigLoadEnd ||
        event instanceof NavigationCancel ||
        event instanceof NavigationEnd
      ) {
        this.lzLoadingV.hide();
      }
    });
  }
}
