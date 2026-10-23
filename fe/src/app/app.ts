import { Component, inject, OnInit, signal } from '@angular/core';
import { Header } from './shared/widgets/header/header';
import { Footer } from './shared/widgets/footer/footer';
import { RouterOutlet } from '@angular/router';
import { MobileViewDirective } from './shared/directives/view-height/mobile-view.directive';
import { SeoMetaService } from './core/seo/meta.service';
import { LazyCompLoadComponent } from './shared/components/lazy-comp-load/lazy-comp-load.component';

@Component({
  imports: [Header, Footer, RouterOutlet, MobileViewDirective, LazyCompLoadComponent],
  selector: 'app-root',
  styleUrl: './app.scss',
  templateUrl: './app.html',
})
export class App implements OnInit {
  private readonly meta = inject(SeoMetaService);
  protected readonly title = signal('QR-Code');

  ngOnInit(): void {
    this.meta.updateMeta('QR Code Generator', 'Generating your own QR code with your style');
  }
}
