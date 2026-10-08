import { Component, effect, EffectCleanupRegisterFn, input, output } from '@angular/core';

@Component({
  selector: 'ht47-qr-result',
  imports: [],
  templateUrl: './qr-result.html',
  styleUrl: './qr-result.scss',
})
export class QrResult {
  imgBlob = input<string>('');
  stopViewResult = output<void>();

  constructor() {
    effect((onCleanup: EffectCleanupRegisterFn) => {
      onCleanup(() => {
        if (this.imgBlob()) {
          URL.revokeObjectURL(this.imgBlob());
        }
      });
    });
  }

  stopViewResultHandling() {
    if (this.stopViewResult) {
      this.stopViewResult.emit();
    }
  }

  downloadBlob() {
    const a = document.createElement('a');
    document.body.append(a);
    a.style = 'display: none';

    a.href = this.imgBlob();
    a.download = `qr-${new Date().toUTCString()}.png`;
    a.click();
    URL.revokeObjectURL(this.imgBlob());
  }
}
