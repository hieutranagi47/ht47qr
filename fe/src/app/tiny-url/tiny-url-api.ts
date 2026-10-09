import { DOCUMENT } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { catchError, from, map, mergeMap, throwError } from 'rxjs';
import { parseAPIError } from '@app/shared/model/api-error';

export interface ShortenURLResponse {
  short_url: string;
  long_url: string;
  expires_at: string;
}

@Injectable({ providedIn: 'root' })
export class TinyUrlApi {
  private readonly http = inject(HttpClient);
  private readonly document = inject(DOCUMENT);

  shortenURL(longURL: string, idempotencyKey: string) {
    // These routes are registered at the Go server root, outside the QR API prefix.
    return this.http
      .post<ShortenURLResponse>(
        '/shorten-url',
        { long_url: longURL },
        {
          headers: { 'Idempotency-Key': idempotencyKey },
        },
      )
      .pipe(
        map((response) => ({
          ...response,
          short_url: new URL(response.short_url, this.document.baseURI).href,
        })),
        catchError((error: unknown) =>
          from(parseAPIError(error, 'Unable to shorten the URL. Please try again.')).pipe(
            mergeMap((response) => throwError(() => response)),
          ),
        ),
      );
  }
}
