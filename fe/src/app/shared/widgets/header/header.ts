import { Component } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';

@Component({
  imports: [RouterLink, RouterLinkActive],
  selector: 'ht47-header',
  styleUrl: './header.scss',
  templateUrl: './header.html',
})
export class Header {}
