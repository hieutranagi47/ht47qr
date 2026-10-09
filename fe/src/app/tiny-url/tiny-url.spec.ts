import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TinyUrl } from './tiny-url';

describe('TinyUrl', () => {
  let component: TinyUrl;
  let fixture: ComponentFixture<TinyUrl>;
  let http: HttpTestingController;

  function enterURL(value: string): void {
    const input: HTMLInputElement = fixture.nativeElement.querySelector('#long-url');
    input.value = value;
    input.dispatchEvent(new Event('input'));
    fixture.detectChanges();
  }

  function submit(): void {
    fixture.nativeElement
      .querySelector('form')
      .dispatchEvent(new Event('submit', { cancelable: true }));
    fixture.detectChanges();
  }

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TinyUrl],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    fixture = TestBed.createComponent(TinyUrl);
    component = fixture.componentInstance;
    http = TestBed.inject(HttpTestingController);
    await fixture.whenStable();
  });

  afterEach(() => http.verify());

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it.each([
    '',
    '   ',
    'example.com',
    'https://',
    'https:example.com',
    'https:///example.com',
    'https://example.com\\path',
    'javascript:alert(1)',
    'ftp://example.com',
    'https://exa mple.com',
  ])('rejects invalid URL %s', (url) => {
    enterURL(url);
    submit();
    expect(component.urlForm().invalid()).toBe(true);
    expect(fixture.nativeElement.querySelector('[type="submit"]').disabled).toBe(true);
    expect(fixture.nativeElement.querySelector('#url-errors').textContent.trim()).not.toBe('');
    http.expectNone('/shorten-url');
  });

  it('posts a trimmed URL once and shows an absolute link and expiry', () => {
    enterURL(' https://example.com/a/long/path?query=value ');
    submit();
    const request = http.expectOne('/shorten-url');
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual({
      long_url: 'https://example.com/a/long/path?query=value',
    });
    expect(request.request.headers.get('Idempotency-Key')).toBeTruthy();
    expect(fixture.nativeElement.querySelector('fieldset').disabled).toBe(true);
    submit();
    http.expectNone('/shorten-url');
    request.flush({
      short_url: '/r/Abcd1234',
      long_url: 'https://example.com/a/long/path?query=value',
      expires_at: '2026-10-10T12:00:00Z',
    });
    fixture.detectChanges();
    expect(component.isLoading()).toBe(false);
    const link: HTMLAnchorElement = fixture.nativeElement.querySelector('a');
    expect(link.href).toBe(new URL('/r/Abcd1234', document.baseURI).href);
    expect(fixture.nativeElement.querySelector('time').getAttribute('datetime')).toBe(
      '2026-10-10T12:00:00Z',
    );
    fixture.nativeElement.querySelector('button').click();
    fixture.detectChanges();
    expect(component.result()).toBeNull();
    expect(component.urlForm.long_url().value()).toBe('');
    expect(component.urlForm().dirty()).toBe(false);
    enterURL('https://example.com/a/long/path?query=value');
    submit();
    const next = http.expectOne('/shorten-url');
    expect(next.request.headers.get('Idempotency-Key')).not.toBe(
      request.request.headers.get('Idempotency-Key'),
    );
    next.flush({
      short_url: '/r/Efgh5678',
      long_url: 'https://example.com',
      expires_at: '2026-10-10T12:00:00Z',
    });
  });

  it('shows server errors and reuses the request key on retry, until the URL changes', async () => {
    enterURL('https://example.com');
    submit();
    const first = http.expectOne('/shorten-url');
    const response = {
      message: 'Could not create short URL',
      slug: 'creation_failed',
      details: [],
    };
    first.flush(response, { status: 500, statusText: 'Server Error' });
    await fixture.whenStable();
    fixture.detectChanges();
    expect(component.isLoading()).toBe(false);
    expect(fixture.nativeElement.querySelector('[role="alert"]').textContent).toContain(
      response.message,
    );
    submit();
    const retry = http.expectOne('/shorten-url');
    expect(retry.request.headers.get('Idempotency-Key')).toBe(
      first.request.headers.get('Idempotency-Key'),
    );
    retry.flush('Bad gateway', { status: 502, statusText: 'Bad Gateway' });
    await fixture.whenStable();
    expect(component.error()?.message).toBe('Unable to shorten the URL. Please try again.');
    enterURL('https://example.org');
    submit();
    const changed = http.expectOne('/shorten-url');
    expect(changed.request.headers.get('Idempotency-Key')).not.toBe(
      first.request.headers.get('Idempotency-Key'),
    );
    changed.flush({
      short_url: '/r/Abcd1234',
      long_url: 'https://example.org',
      expires_at: '2026-10-10T12:00:00Z',
    });
  });

  it('cancels an in-flight request when the component is destroyed', () => {
    enterURL('https://example.com');
    submit();
    const request = http.expectOne('/shorten-url');
    fixture.destroy();
    expect(request.cancelled).toBe(true);
  });
});
