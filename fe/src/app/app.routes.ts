import { Routes } from '@angular/router';
import { QRCODE_ROUTE, TINY_URL_ROUTE } from './shared/constants/app.settings';

export const routes: Routes = [
  {
    path: `${QRCODE_ROUTE}`,
    loadChildren: () => {
      return import('./qrcode/qrcode.routes').then((r) => r.routes);
    },
  },
  {
    path: `${TINY_URL_ROUTE}`,
    loadComponent: () => import('./tiny-url/tiny-url').then((c) => c.TinyUrl),
  },
  {
    path: 'interview',
    loadComponent: () => {
      return import('./interview/interview').then((m) => m.Interview);
    },
  },
  {
    path: '',
    pathMatch: 'full',
    loadComponent: () => {
      return import('./home/home').then((m) => m.Home);
    },
  },
];
