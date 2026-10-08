import { isBrowser } from './running-env';

export const isLegacyBrowser = (): boolean => {
  return (
    !('IntersectionObserver' in window) ||
    !('IntersectionObserverEntry' in window) ||
    !('intersectionRatio' in window?.IntersectionObserverEntry?.prototype)
  );
};

export const isModemBrowser = (): boolean => {
  return (
    window.IntersectionObserverEntry &&
    !!('intersectionRatio' in window.IntersectionObserverEntry.prototype)
  );
};

export const goToTop = (): void => {
  if (isBrowser) {
    setTimeout(() => {
      window.scrollTo({
        top: 0,
        left: 0,
        behavior: 'smooth',
      });
    }, 0);
  }
};

export const scrollToTopElm = (elmId: string): void => {
  const elm = document.querySelector(`#${elmId}`);
  if (elm) {
    elm.scrollTo({
      top: 0,
      left: 0,
      behavior: 'smooth',
    });
  }
};

export const downloadFile = (path: string, filename: string): void => {
  if (!isBrowser) {
    return;
  }
  // Create a new anchor link
  const anchor = document.createElement('a');
  anchor.href = path;
  anchor.download = filename;

  // Append the anchor to the DOM
  document.body.appendChild(anchor);

  // Trigger the `click` event
  anchor.click();

  // Remove the anchor element from the DOM
  document.body.removeChild(anchor);
};
