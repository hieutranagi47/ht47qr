import { Component, computed, inject, signal } from '@angular/core';
import { AppStore } from '@app/app-store';
import {
  form,
  Field,
  required,
  min,
  max,
  pattern,
  minLength,
  maxLength,
  email,
  FormField,
} from '@angular/forms/signals';
import { QRCodeGeoPayload, QRCodePayload } from '@app/shared/model/qr-request';
import { ColorSketchModule } from 'ngx-color/sketch';
import { ColorEvent } from 'ngx-color';
import { ClickIODirective } from '@app/shared/directives/mouse-click/click-io.directive';
import { NgClass } from '@angular/common';
import { QrResult } from '@app/shared/components/qr-result/qr-result';
import { Modal } from '@app/shared/components/modal/modal';

const initFormData: QRCodePayload<QRCodeGeoPayload> = {
  data: { type: 'geo', latitude: '', longitude: '', label: '' },
  logo_img: null,
  halftone_image: null,
  qr_width: 10,
  foreground_color: '#000000',
  border_width: 10,
  is_circle_shape: false,
  is_custom_shape: false,
};

@Component({
  selector: 'ht47-qr-geo',
  imports: [ColorSketchModule, ClickIODirective, NgClass, FormField, NgClass, QrResult, Modal],
  templateUrl: './qr-geo.html',
  styleUrl: './qr-geo.scss',
})
export class QrGeo {
  readonly appStore = inject(AppStore);
  readonly isLoading = signal(false);
  readonly error = signal<string | null>(null);
  readonly imgBlob = signal<string>('');
  protected readonly qrCodePayload = signal<QRCodePayload<QRCodeGeoPayload>>(initFormData);
  protected readonly qrCodeTextForm = form(this.qrCodePayload, (path) => {
    required(path.data.latitude, { message: 'Latitude number is required' });
    pattern(path.data.latitude, /^-?(180(\.0{1,10})?|((1[0-7]\d)|([1-9]?\d))(\.\d{1,10})?)$/, {
      message: 'Wrong latitude format, ie: 10.801379',
    });
    required(path.data.longitude, { message: 'Name is required' });
    pattern(path.data.longitude, /^-?(180(\.0{1,10})?|((1[0-7]\d)|([1-9]?\d))(\.\d{1,10})?)$/, {
      message: 'Wrong longitude format, ie: 106.711273',
    });
    minLength(path.data.label, 3, { message: 'From 3 characters' });
    maxLength(path.data.label, 250, { message: 'Must less than 250 characters' });
    min(path.qr_width, 5, { message: 'Size of the QR Code must be 5 -> 21' });
    max(path.qr_width, 21, { message: 'Size of the QR Code must be 5 -> 21' });
    min(path.border_width, 0, { message: 'Border width must be 0 -> 20' });
    max(path.border_width, 20, { message: 'Border width must be 0 -> 20' });
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
    this.isLoading.set(true);
    this.error.set(null);

    this.appStore.generateQRCodeWithHttpClient(this.qrCodeTextForm().value()).subscribe({
      next: (response) => {
        if (response instanceof Blob) {
          this.imgBlob.set(URL.createObjectURL(response));
        }
      },
      error: (err) => {
        this.error.set(err?.error?.error_message || 'An unexpected error occurred.');
      },
      complete: () => {
        this.isLoading.set(false);
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
  }
}
