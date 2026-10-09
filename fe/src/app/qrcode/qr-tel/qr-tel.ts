import { finalize } from 'rxjs';
import { APIErrorResponse } from '@app/shared/model/qr-request';
import { Component, computed, inject, signal } from '@angular/core';
import { AppStore } from '@app/app-store';
import { form, FormField, required, min, max, pattern } from '@angular/forms/signals';
import { QRCodePayload, QRCodeSmsPayload, QRCodeTelPayload } from '@app/shared/model/qr-request';
import { ColorSketchModule } from 'ngx-color/sketch';
import { ColorEvent } from 'ngx-color';
import { ClickIODirective } from '@app/shared/directives/mouse-click/click-io.directive';
import { NgClass } from '@angular/common';
import { QrResult } from '@app/shared/components/qr-result/qr-result';
import { Modal } from '@app/shared/components/modal/modal';

const initFormData: QRCodePayload<QRCodeTelPayload> = {
  data: { type: 'tel', phone_number: '' },
  logo_img: null,
  halftone_image: null,
  qr_width: 10,
  foreground_color: '#000000',
  border_width: 10,
  is_circle_shape: false,
  is_custom_shape: false,
};

@Component({
  selector: 'ht47-qr-tel',
  imports: [ColorSketchModule, ClickIODirective, NgClass, FormField, NgClass, QrResult, Modal],
  templateUrl: './qr-tel.html',
  styleUrl: './qr-tel.scss',
})
export class QrTel {
  readonly appStore = inject(AppStore);
  readonly isLoading = signal(false);
  readonly showColorPickerModal = signal(false);
  readonly error = signal<APIErrorResponse | null>(null);
  readonly imgBlob = signal<string>('');
  protected readonly qrCodePayload = signal<QRCodePayload<QRCodeTelPayload>>(initFormData);
  protected readonly qrCodeTextForm = form(this.qrCodePayload, (path) => {
    required(path.data.phone_number, { message: 'Phone number is required.' });
    pattern(path.data.phone_number, /^[+]{1}(?:[0-9\-\\(\\)\\/.]\s?){6,15}[0-9]{1}$/, {
      message: 'Enter a valid phone number, for example +84912346789.',
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

  toggleColorPickerModal() {
    this.showColorPickerModal.update((value) => !value);
  }

  closeColorPickerModal() {
    this.showColorPickerModal.set(false);
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
