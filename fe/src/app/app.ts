import { Component, signal } from '@angular/core';
import { Header } from './shared/widgets/header/header';
import { Footer } from './shared/widgets/footer/footer';
import { RouterOutlet } from '@angular/router';
import { MobileViewDirective } from './shared/directives/view-height/mobile-view.directive';

@Component({
  imports: [Header, Footer, RouterOutlet, MobileViewDirective],
  selector: 'app-root',
  styleUrl: './app.scss',
  templateUrl: './app.html',
})
export class App {
  protected readonly title = signal('QR-Code');
}
