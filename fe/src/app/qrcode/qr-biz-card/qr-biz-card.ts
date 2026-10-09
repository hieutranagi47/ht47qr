import { NgClass } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import {
  email,
  Field,
  form,
  FormField,
  max,
  maxLength,
  min,
  minLength,
  pattern,
  required,
} from '@angular/forms/signals';
import { AppStore } from '@app/app-store';
import { Modal } from '@app/shared/components/modal/modal';
import { QrResult } from '@app/shared/components/qr-result/qr-result';
import { ClickIODirective } from '@app/shared/directives/mouse-click/click-io.directive';
import { QRCodePayload, QRCodeVCardPayload } from '@app/shared/model/qr-request';
import { ColorEvent } from 'ngx-color';
import { ColorSketchModule } from 'ngx-color/sketch';

const initFormData: QRCodePayload<QRCodeVCardPayload> = {
  data: {
    type: 'vcard',
    phone: '',
    name: '',
    email: '',
    organization: '',
    title: '',
    address: '',
    website: '',
  },
  logo_img: null,
  halftone_image: null,
  qr_width: 10,
  foreground_color: '#000000',
  border_width: 10,
  is_circle_shape: false,
  is_custom_shape: false,
};

@Component({
  selector: 'ht47-qr-biz-card',
  imports: [ColorSketchModule, ClickIODirective, NgClass, FormField, NgClass, QrResult, Modal],
  templateUrl: './qr-biz-card.html',
  styleUrl: './qr-biz-card.scss',
})
export class QrBizCard {
  readonly appStore = inject(AppStore);
  readonly isLoading = signal(false);
  readonly error = signal<string | null>(null);
  readonly imgBlob = signal<string>('');
  protected readonly qrCodePayload = signal<QRCodePayload<QRCodeVCardPayload>>(initFormData);
  protected readonly qrCodeTextForm = form(this.qrCodePayload, (path) => {
    required(path.data.phone, { message: 'The Phone number is required' });
    pattern(path.data.phone, /^[+]{1}(?:[0-9\-\\(\\)\\/.]\s?){6,15}[0-9]{1}$/, {
      message: 'Wrong phone number format, ie: +84912346789',
    });
    required(path.data.name, { message: 'Name is required' });
    minLength(path.data.name, 3, { message: 'Name must be from 3 characters' });
    maxLength(path.data.name, 100, { message: 'Name must be less than 100 characters' });
    required(path.data.title, { message: 'Title is required' });
    minLength(path.data.title, 3, { message: 'Title must be from 3 characters' });
    maxLength(path.data.title, 100, { message: 'Title must be less than 100 characters' });
    required(path.data.email, { message: 'Email is required' });
    email(path.data.email, { message: 'Please input a valid email' });
    required(path.data.organization, { message: 'Organization is required' });
    pattern(path.data.website, /^(https?:\/\/)?([\da-z\.-]+)\.([a-z\.]{2,6})([\/\w \.-]*)*\/?$/gi, {
      message: 'Incorrect website link',
    });
    minLength(path.data.organization, 10, { message: 'Organization must be from 10 characters' });
    maxLength(path.data.organization, 100, {
      message: 'Organization must be less than 100 characters',
    });
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

  onSubmit($event: Event): void {
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

  forceGroundColorChange($event: ColorEvent): void {
    this.qrCodeTextForm.foreground_color().controlValue.set($event.color.hex);
  }

  createANewOne(): void {
    this.qrCodeTextForm().controlValue.set(initFormData);
    this.qrCodeTextForm().reset();
    this.imgBlob.set('');
  }
}
