import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';
import { AppStore } from './app-store';
import { QRCodePayload, QRCodeTextPayload } from './shared/model/qr-request';

const payload: QRCodePayload<QRCodeTextPayload> = {
  data: { type: 'text', text: 'hello world' },
  foreground_color: '#000000',
  qr_width: 10,
  border_width: 10,
  is_circle_shape: false,
  is_custom_shape: false,
  logo_img: null,
  halftone_image: null,
};
const response = {
  message: 'Please correct the form and try again.',
  slug: 'form_validation',
  details: [
    {
      entity_type: 'form_field',
      entity_id: 'invalid_input_logo_img',
      error_slug: 'logo_img',
      message: 'Only valid PNG and JPEG images are allowed.',
    },
  ],
};

describe('AppStore QR generation', () => {
  let service: AppStore;
  let http: HttpTestingController;
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(AppStore);
    http = TestBed.inject(HttpTestingController);
  });
  afterEach(() => http.verify());

  it('decodes JSON Blob errors and preserves validation details', async () => {
    const result = firstValueFrom(service.generateQRCodeWithHttpClient(payload));
    const assertion = expect(result).rejects.toEqual(response);
    const request = http.expectOne((request) => request.url.endsWith('/generate-qrcode'));
    expect(request.request.responseType).toBe('blob');
    expect(request.request.body.get('data')).toBe(JSON.stringify(payload.data));
    request.flush(new Blob([JSON.stringify(response)], { type: 'application/json' }), {
      status: 400,
      statusText: 'Bad Request',
    });
    await assertion;
  });

  it('returns the image Blob on success', async () => {
    const result = firstValueFrom(service.generateQRCodeWithHttpClient(payload));
    const blob = new Blob(['image'], { type: 'image/png' });
    http.expectOne((request) => request.url.endsWith('/generate-qrcode')).flush(blob);
    expect(await result).toBe(blob);
  });
});
