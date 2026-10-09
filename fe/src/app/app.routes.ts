import { Routes } from '@angular/router';
import { QRCODE_ROUTE } from './shared/constants/app.settings';

export const routes: Routes = [
  {
    path: `${QRCODE_ROUTE}`,
    loadChildren: () => {
      return import('./qrcode/qrcode.routes').then((r) => r.routes);
    },
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
