import { Type, WritableSignal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { Subject } from 'rxjs';
import { AppStore } from '@app/app-store';
import { APIErrorResponse, QRCodePayload } from '@app/shared/model/qr-request';
import { QrText } from './qr-text/qr-text';
import { QrTel } from './qr-tel/qr-tel';
import { QrSms } from './qr-sms/qr-sms';
import { QrEmail } from './qr-email/qr-email';
import { QrEmail2 } from './qr-email2/qr-email2';
import { QrGeo } from './qr-geo/qr-geo';
import { QrContact } from './qr-contact/qr-contact';
import { QrBizCard } from './qr-biz-card/qr-biz-card';

interface QRForm {
  qrCodePayload: WritableSignal<QRCodePayload<Record<string, string>>>;
  isLoading: WritableSignal<boolean>;
  error: WritableSignal<APIErrorResponse | null>;
  onSubmit(event: Event): void;
}

const values: Record<string, string> = {
  text: 'Hello from the QR form',
  phone_number: '+84912346789',
  phone: '+84912346789',
  name: 'Jane Doe',
  email: 'jane@example.com',
  subject: 'Hello from the QR form',
  body: 'Hello from the QR form',
  message: 'Hello from the QR form',
  latitude: '10',
  longitude: '106',
  label: 'Home',
  organization: 'Example company',
  title: 'Developer',
  address: 'Example address',
  website: 'https://example.com',
};
const validationError: APIErrorResponse = {
  message: 'Please correct the form and try again.',
  slug: 'form_validation',
  details: [
    {
      entity_type: 'form_field',
      entity_id: 'invalid_input_logo_img',
      error_slug: 'logo_img',
      message: 'Logo: Only valid PNG and JPEG images are allowed.',
    },
  ],
};

for (const componentType of [
  QrText,
  QrTel,
  QrSms,
  QrEmail,
  QrEmail2,
  QrGeo,
  QrContact,
  QrBizCard,
]) {
  describe(`${componentType.name} request errors`, () => {
    it('displays validation details, stops loading, and allows retry', async () => {
      let request = new Subject<Blob>();
      const generate = vi.fn(() => request.asObservable());
      await TestBed.configureTestingModule({
        imports: [componentType],
        providers: [{ provide: AppStore, useValue: { generateQRCodeWithHttpClient: generate } }],
      }).compileComponents();
      const fixture = TestBed.createComponent(componentType as Type<unknown>);
      const component = fixture.componentInstance as QRForm;
      component.qrCodePayload.update((payload) => ({
        ...payload,
        data: Object.fromEntries(
          Object.entries(payload.data).map(([key, value]) => [key, values[key] ?? value]),
        ),
      }));
      fixture.detectChanges();
      component.onSubmit(new Event('submit'));
      expect(generate).toHaveBeenCalledTimes(1);
      expect(component.isLoading()).toBe(true);
      fixture.detectChanges();
      const submitButton: HTMLButtonElement = fixture.nativeElement.querySelector('button[type="submit"]');
      const form: HTMLFormElement = fixture.nativeElement.querySelector('form');
      const controls = Array.from(form.querySelectorAll('input, textarea, button'));
      expect(submitButton.disabled).toBe(true);
      expect(submitButton.textContent).toBe('Generating...');
      expect(form.getAttribute('aria-busy')).toBe('true');
      expect(controls.length).toBeGreaterThan(0);
      expect(controls.every((control) => control.matches(':disabled'))).toBe(true);
      component.onSubmit(new Event('submit'));
      expect(generate).toHaveBeenCalledTimes(1);
      request.error(validationError);
      fixture.detectChanges();
      expect(component.isLoading()).toBe(false);
      expect(submitButton.textContent).toBe('Generate QR Code');
      expect(form.getAttribute('aria-busy')).toBe('false');
      expect(form.querySelectorAll('input:disabled, textarea:disabled').length).toBe(0);
      expect(fixture.nativeElement.querySelector('[role="alert"]').textContent).toContain(
        validationError.details[0].message,
      );
      request = new Subject<Blob>();
      component.onSubmit(new Event('submit'));
      expect(component.error()).toBeNull();
      expect(generate).toHaveBeenCalledTimes(2);
      request.complete();
      fixture.detectChanges();
      expect(component.isLoading()).toBe(false);
      expect(submitButton.textContent).toBe('Generate QR Code');
      expect(form.querySelectorAll('input:disabled, textarea:disabled').length).toBe(0);
    });
  });
}
