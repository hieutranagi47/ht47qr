import { HttpClient, httpResource } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { environment as env } from '@env/environment';
import { catchError, from, mergeMap, throwError } from 'rxjs';
import { parseAPIError } from './shared/model/api-error';
import { QRCodePayload, QRCodeResponse } from './shared/model/qr-request';

@Injectable({
  providedIn: 'root',
})
export class AppStore {
  httpClient = inject(HttpClient);

  // Instance call the method
  generateQRCode<T>(payload: QRCodePayload<T>) {
    return httpResource<QRCodeResponse>(() => ({
      url: `${env.apiEndpoint}${env.apiVersion}/generate-qrcode`,
      method: 'POST',
      headers: { 'X-Special': 'true' },
      body: JSON.stringify(payload),
      params: {},
      reportProgress: true,
      transferCache: true,
      keepalive: true,
      mode: 'cors',
      redirect: 'error',
      priority: 'high',
      cache: 'force-cache',
      credentials: 'include',
      referrer: '',
      integrity: '',
    }));
  }

  generateQRCodeWithHttpClient<T>(payload: QRCodePayload<T>) {
    const form = new FormData();
    form.append('data', JSON.stringify(payload.data));
    form.append('foreground_color', payload.foreground_color);
    form.append('border_width', payload.border_width.toString());
    form.append('qr_width', payload.qr_width.toString());
    form.append('is_circle_shape', '' + payload.is_circle_shape);
    form.append('is_custom_shape', '' + payload.is_custom_shape);
    if (payload.logo_img) {
      form.append('logo_img', payload.logo_img);
    }
    if (payload.halftone_image) {
      form.append('halftone_img', payload.halftone_image);
    }
    return this.httpClient
      .post(`${env.apiEndpoint}${env.apiVersion}/generate-qrcode`, form, {
        headers: { 'X-Special': 'true' },
        responseType: 'blob',
        observe: 'body',
      })
      .pipe(
        catchError((error: unknown) =>
          from(parseAPIError(error)).pipe(mergeMap((response) => throwError(() => response))),
        ),
      );
  }
}
