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
  email,
  FormField,
} from '@angular/forms/signals';
import { QRCodeEMailPayload, QRCodePayload } from '@app/shared/model/qr-request';
import { ColorSketchModule } from 'ngx-color/sketch';
import { ColorEvent } from 'ngx-color';
import { ClickIODirective } from '@app/shared/directives/mouse-click/click-io.directive';
import { NgClass } from '@angular/common';
import { QrResult } from '@app/shared/components/qr-result/qr-result';

const initFormData: QRCodePayload<QRCodeEMailPayload> = {
  data: { type: 'mail', email: '', subject: '', body: '' },
  logo_img: null,
  halftone_image: null,
  qr_width: 10,
  foreground_color: '#000000',
  border_width: 10,
  is_circle_shape: false,
  is_custom_shape: false,
};

@Component({
  selector: 'ht47-qr-email2',
  imports: [ColorSketchModule, ClickIODirective, NgClass, FormField, NgClass, QrResult],
  templateUrl: './qr-email2.html',
  styleUrl: './qr-email2.scss',
})
export class QrEmail2 {
  readonly appStore = inject(AppStore);
  readonly isLoading = signal(false);
  readonly showColorPickerModal = signal(false);
  readonly error = signal<string | null>(null);
  readonly imgBlob = signal<string>('');
  protected readonly qrCodePayload = signal<QRCodePayload<QRCodeEMailPayload>>(initFormData);
  protected readonly qrCodeTextForm = form(this.qrCodePayload, (path) => {
    required(path.data.email, { message: 'Email is required' });
    email(path.data.email, { message: 'Please input a valid email' });
    required(path.data.subject, { message: 'Message is required' });
    minLength(path.data.subject, 10, {
      message: 'Message length must be greater than 10 characters',
    });
    maxLength(path.data.subject, 255, { message: 'Text length must be less than 10 characters' });
    required(path.data.body, { message: 'Message is required' });
    minLength(path.data.body, 10, {
      message: 'Message length must be greater than 10 characters',
    });
    maxLength(path.data.body, 255, { message: 'Text length must be less than 10 characters' });
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
  }
}
