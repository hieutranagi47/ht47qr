import { Component } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { QRCODE_ROUTE } from '@app/shared/constants/app.settings';

@Component({
  imports: [RouterOutlet, RouterLink, RouterLinkActive],
  selector: 'ht47-qrcode',
  styleUrl: './qrcode.scss',
  templateUrl: './qrcode.html',
})
export class Qrcode {
  readonly qrCodeRoute = QRCODE_ROUTE;
}
