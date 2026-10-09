import { Component, viewChild } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { vi } from 'vitest';
import { Modal } from './modal';

@Component({
  imports: [Modal],
  template: `
    <ht47-modal #modal ariaLabel="Choose a color">
      <button trigger-btn type="button"><span>Choose color</span></button>
      <p>Modal content</p>
      <button class="close-button" type="button" (click)="modal.closeModal()">Apply</button>
    </ht47-modal>
  `,
})
class ModalHost {
  readonly modal = viewChild.required(Modal);
}

describe('Modal', () => {
  let fixture: ComponentFixture<ModalHost>;
  let dialog: HTMLDialogElement;
  let showModal: ReturnType<typeof vi.fn>;
  let close: ReturnType<typeof vi.fn>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ModalHost],
    }).compileComponents();

    fixture = TestBed.createComponent(ModalHost);
    await fixture.whenStable();
    dialog = fixture.nativeElement.querySelector('dialog');
    // jsdom does not implement native modal focus or top-layer behavior.
    showModal = vi.fn(() => dialog.setAttribute('open', ''));
    close = vi.fn(() => dialog.removeAttribute('open'));
    Object.defineProperties(dialog, {
      showModal: { configurable: true, value: showModal },
      close: { configurable: true, value: close },
    });
  });

  it('projects the trigger outside the dialog and content inside it', () => {
    expect(dialog.querySelector('[trigger-btn]')).toBeNull();
    expect(fixture.nativeElement.querySelector('[trigger-btn]')).toBeTruthy();
    expect(dialog.querySelector('p')?.textContent).toBe('Modal content');
    expect(dialog.getAttribute('aria-label')).toBe('Choose a color');
    expect(dialog.open).toBe(false);
  });

  it('opens when a nested element inside the projected trigger is clicked', () => {
    fixture.nativeElement.querySelector('[trigger-btn] span').click();

    expect(showModal).toHaveBeenCalledOnce();
    expect(dialog.open).toBe(true);
  });

  it('closes from a projected content button using the modal reference', () => {
    fixture.nativeElement.querySelector('[trigger-btn]').click();
    dialog.querySelector<HTMLButtonElement>('.close-button')!.click();

    expect(close).toHaveBeenCalledOnce();
    expect(dialog.open).toBe(false);
    expect(showModal).toHaveBeenCalledOnce();
  });

  it('does not open when regular content is clicked', () => {
    dialog.querySelector('p')!.click();

    expect(showModal).not.toHaveBeenCalled();
  });

  it('ignores repeated open and close calls', () => {
    const modal = fixture.componentInstance.modal();
    modal.closeModal();
    expect(close).not.toHaveBeenCalled();
    modal.openModal();
    modal.openModal();
    modal.closeModal();
    modal.closeModal();

    expect(showModal).toHaveBeenCalledOnce();
    expect(close).toHaveBeenCalledOnce();
  });

  it('toggles the dialog open and closed', () => {
    const modal = fixture.componentInstance.modal();
    modal.toggleModal();
    expect(dialog.open).toBe(true);
    modal.toggleModal();
    expect(dialog.open).toBe(false);
  });
});
