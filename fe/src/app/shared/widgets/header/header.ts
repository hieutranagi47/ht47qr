import { Component } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { QRCODE_ROUTE } from '@app/shared/constants/app.settings';

@Component({
  imports: [RouterLink, RouterLinkActive],
  selector: 'ht47-header',
  styleUrl: './header.scss',
  templateUrl: './header.html',
})
export class Header {
  readonly qrCodeRoute = QRCODE_ROUTE;
}
