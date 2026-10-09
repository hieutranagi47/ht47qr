import { DatePipe } from '@angular/common';
import {
  ChangeDetectionStrategy,
  Component,
  DestroyRef,
  inject,
  OnInit,
  signal,
} from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { form, FormField, required, validate } from '@angular/forms/signals';
import { finalize } from 'rxjs';
import { APIErrorResponse } from '@app/shared/model/qr-request';
import { ShortenURLResponse, TinyUrlApi } from './tiny-url-api';
import { SeoMetaService } from '@app/core/seo/meta.service';

@Component({
  imports: [DatePipe, FormField],
  selector: 'ht47-tiny-url',
  styleUrl: './tiny-url.scss',
  templateUrl: './tiny-url.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class TinyUrl implements OnInit {
  private readonly api = inject(TinyUrlApi);
  private readonly meta = inject(SeoMetaService);
  private readonly destroyRef = inject(DestroyRef);
  private request: { longURL: string; key: string } | null = null;
  readonly isLoading = signal(false);
  readonly error = signal<APIErrorResponse | null>(null);
  readonly result = signal<ShortenURLResponse | null>(null);
  private readonly payload = signal({ long_url: '' });
  readonly urlForm = form(this.payload, (path) => {
    required(path.long_url, { message: 'URL is required.' });
    validate(path.long_url, ({ value }) => {
      const raw = value().trim();
      if (!value()) return null;
      try {
        const url = new URL(raw);
        if (/^https?:\/\/[^/\\]/.test(raw) && url.hostname && !/[\s\\]/.test(raw)) return null;
      } catch {
        // Report malformed URLs through the same field validation message.
      }
      return { kind: 'url', message: 'Enter a valid URL starting with http:// or https://.' };
    });
  });

  ngOnInit(): void {
    this.meta.updateMeta('Shorten URL', 'make your url shortener to share with your partners');
  }

  onSubmit(event: Event): void {
    event.preventDefault();
    this.urlForm.long_url().markAsTouched();
    if (this.isLoading() || this.urlForm().invalid()) return;

    const longURL = this.urlForm.long_url().value().trim();
    if (this.request?.longURL !== longURL) {
      this.request = { longURL, key: crypto.randomUUID() };
    }
    this.isLoading.set(true);
    this.error.set(null);
    this.api
      .shortenURL(longURL, this.request.key)
      .pipe(
        takeUntilDestroyed(this.destroyRef),
        finalize(() => this.isLoading.set(false)),
      )
      .subscribe({
        next: (response) => this.result.set(response),
        error: (error: APIErrorResponse) => this.error.set(error),
      });
  }

  createANewOne(): void {
    this.urlForm().reset({ long_url: '' });
    this.result.set(null);
    this.error.set(null);
    this.request = null;
  }
}
