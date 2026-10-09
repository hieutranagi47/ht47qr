import { ChangeDetectionStrategy, Component, ElementRef, input, viewChild } from '@angular/core';

@Component({
  selector: 'ht47-modal',
  changeDetection: ChangeDetectionStrategy.OnPush,
  styleUrl: './modal.scss',
  templateUrl: './modal.html',
})
export class Modal {
  readonly ariaLabel = input('Dialog');
  private readonly dialogRef = viewChild.required<ElementRef<HTMLDialogElement>>('dialog');

  openModal(): void {
    const dialog = this.dialogRef().nativeElement;
    if (!dialog.open) {
      dialog.showModal();
    }
  }

  closeModal(): void {
    const dialog = this.dialogRef().nativeElement;
    if (dialog.open) {
      dialog.close();
    }
  }

  toggleModal(): void {
    if (this.dialogRef().nativeElement.open) {
      this.closeModal();
    } else {
      this.openModal();
    }
  }
}
