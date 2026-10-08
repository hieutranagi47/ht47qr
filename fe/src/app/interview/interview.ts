import { Component, OnInit } from '@angular/core';

interface Item {
  id: number;
  value: number;
  isActive: boolean;
}

const items: Item[] = [
  { id: 1, value: 1, isActive: true },
  { id: 2, value: 2, isActive: false },
  { id: 3, value: 3, isActive: true },
  { id: 4, value: 4, isActive: true },
  { id: 5, value: 5, isActive: false },
  { id: 6, value: 6, isActive: true },
  { id: 7, value: 7, isActive: false },
  { id: 8, value: 8, isActive: true },
  { id: 9, value: 9, isActive: true },
  { id: 10, value: 10, isActive: false },
];

@Component({
  selector: 'ht47-interview',
  imports: [],
  templateUrl: './interview.html',
  styleUrl: './interview.scss',
})
export class Interview implements OnInit {
  evenSum = 0;
  ngOnInit(): void {
    this.evenSum = items
      .filter((item) => item.isActive && item.value % 2 === 0)
      .reduce((sum, item) => sum + item.value, 0);
  }
}

/***
# Tech Interview Exercise: Sum of Even and Active Numbers

## Overview
This project demonstrates two different approaches to compute the sum of the even and active numbers from a list of objects using Angular. The result will be displayed in a centered manner on a black background, following the style of the provided mockup.

## Problem Statement
Given a list of items with `id`, `value`, and `isActive` properties, compute the sum of the even and active `value` fields in **two different ways** and display the result centered on the page. The page will have a black background with white text, styled according to a mockup.

### Interface
```typescript
interface Item {
  id: number;
  value: number;
  isActive: boolean;
}
```
 */
