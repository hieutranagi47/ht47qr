import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: 'text',
    loadComponent() {
      return import('./qr-text/qr-text').then((m) => m.QrText);
    },
  },
  {
    path: 'sms',
    loadComponent() {
      return import('./qr-sms/qr-sms').then((m) => m.QrSms);
    },
  },
  {
    path: 'tel',
    loadComponent() {
      return import('./qr-tel/qr-tel').then((m) => m.QrTel);
    },
  },
  {
    path: 'contact',
    loadComponent() {
      return import('./qr-contact/qr-contact').then((m) => m.QrContact);
    },
  },
  {
    path: 'email',
    loadComponent() {
      return import('./qr-email/qr-email').then((m) => m.QrEmail);
    },
  },
  {
    path: 'email-to',
    loadComponent() {
      return import('./qr-email2/qr-email2').then((m) => m.QrEmail2);
    },
  },
  {
    path: 'geo',
    loadComponent() {
      return import('./qr-geo/qr-geo').then((m) => m.QrGeo);
    },
  },
  {
    path: 'i-cal',
    loadComponent() {
      return import('./qr-ical/qr-ical').then((m) => m.QrIcal);
    },
  },
  {
    path: 'biz-card',
    loadComponent() {
      return import('./qr-biz-card/qr-biz-card').then((m) => m.QrBizCard);
    },
  },
  {
    path: 'calendar',
    loadComponent() {
      return import('./qr-calendar/qr-calendar').then((m) => m.QrCalendar);
    },
  },
  {
    path: 'interview',
    loadComponent() {
      return import('./interview/interview').then((m) => m.Interview);
    },
  },
  {
    path: '',
    pathMatch: 'full',
    loadComponent() {
      return import('./home/home').then((m) => m.Home);
    },
  },
];
