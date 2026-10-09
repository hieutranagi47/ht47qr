import { finalize } from 'rxjs';
import { APIErrorResponse } from '@app/shared/model/qr-request';
import { Component, computed, inject, signal } from '@angular/core';
import { AppStore } from '@app/app-store';
import {
  form,
  Field,
  required,
  minLength,
  maxLength,
  min,
  max,
  pattern,
  FormField,
} from '@angular/forms/signals';
import { QRCodePayload, QRCodeSmsPayload } from '@app/shared/model/qr-request';
import { ColorSketchModule } from 'ngx-color/sketch';
import { ColorEvent } from 'ngx-color';
import { ClickIODirective } from '@app/shared/directives/mouse-click/click-io.directive';
import { NgClass } from '@angular/common';
import { QrResult } from '@app/shared/components/qr-result/qr-result';
import { Modal } from '@app/shared/components/modal/modal';

const initFormData: QRCodePayload<QRCodeSmsPayload> = {
  data: { type: 'sms', phone_number: '', message: '' },
  logo_img: null,
  halftone_image: null,
  qr_width: 10,
  foreground_color: '#000000',
  border_width: 10,
  is_circle_shape: false,
  is_custom_shape: false,
};

@Component({
  selector: 'ht47-qr-sms',
  imports: [ColorSketchModule, ClickIODirective, NgClass, FormField, NgClass, QrResult, Modal],
  templateUrl: './qr-sms.html',
  styleUrl: './qr-sms.scss',
})
export class QrSms {
  readonly appStore = inject(AppStore);
  readonly isLoading = signal(false);
  readonly error = signal<APIErrorResponse | null>(null);
  readonly imgBlob = signal<string>('');
  protected readonly qrCodePayload = signal<QRCodePayload<QRCodeSmsPayload>>(initFormData);
  protected readonly qrCodeTextForm = form(this.qrCodePayload, (path) => {
    required(path.data.phone_number, { message: 'Phone number is required.' });
    pattern(path.data.phone_number, /^[+]{1}(?:[0-9\-\\(\\)\\/.]\s?){6,15}[0-9]{1}$/, {
      message: 'Enter a valid phone number, for example +84912346789.',
    });
    required(path.data.message, { message: 'Message is required.' });
    minLength(path.data.message, 10, {
      message: 'Message must contain at least 10 characters.',
    });
    maxLength(path.data.message, 255, {
      message: 'Message must contain no more than 255 characters.',
    });
    min(path.qr_width, 6, { message: 'QR width must be between 6 and 21.' });
    max(path.qr_width, 21, { message: 'QR width must be between 6 and 21.' });
    min(path.border_width, 0, { message: 'Border width must be between 0 and 20.' });
    max(path.border_width, 20, { message: 'Border width must be between 0 and 20.' });
  });
  previewLogoImg = computed(() => {
    const logo = this.qrCodeTextForm.logo_img().value();
    if (logo) {
      return URL.createObjectURL(logo);
    }
    return '';
  });
  previewHalftoneImg = computed(() => {
    const halftone = this.qrCodeTextForm.halftone_image().value();
    if (halftone) {
      return URL.createObjectURL(halftone);
    }
    return '';
  });

  onSubmit($event: Event) {
    $event.preventDefault();
    if (this.isLoading() || this.qrCodeTextForm().invalid()) return;
    this.isLoading.set(true);
    this.error.set(null);

    this.appStore
      .generateQRCodeWithHttpClient(this.qrCodeTextForm().value())
      .pipe(finalize(() => this.isLoading.set(false)))
      .subscribe({
        next: (response) => {
          if (response instanceof Blob) {
            this.imgBlob.set(URL.createObjectURL(response));
          }
        },
        error: (err: APIErrorResponse) => {
          this.error.set(err);
        },
      });
  }

  addLogo($event: Event): void {
    const input = $event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      this.qrCodeTextForm.logo_img().controlValue.set(input.files[0]);
    }
  }

  clearLogoImage(): void {
    this.qrCodeTextForm.logo_img().controlValue.set(null);
  }

  addHalftoneLogo($event: Event): void {
    const input = $event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      this.qrCodeTextForm.halftone_image().controlValue.set(input.files[0]);
    }
  }

  clearHalftoneImage(): void {
    this.qrCodeTextForm.halftone_image().controlValue.set(null);
  }

  forceGroundColorChange($event: ColorEvent) {
    this.qrCodeTextForm.foreground_color().controlValue.set($event.color.hex);
  }

  createANewOne(): void {
    this.qrCodeTextForm().controlValue.set(initFormData);
    this.qrCodeTextForm().reset();
    this.imgBlob.set('');
    this.error.set(null);
  }
}
