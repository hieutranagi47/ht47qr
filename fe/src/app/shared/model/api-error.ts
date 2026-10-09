import { HttpErrorResponse } from '@angular/common/http';
import { APIErrorDetail, APIErrorResponse } from './qr-request';

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function isDetail(value: unknown): value is APIErrorDetail {
  return (
    isRecord(value) &&
    ['entity_type', 'entity_id', 'error_slug', 'message'].every(
      (key) => typeof value[key] === 'string',
    )
  );
}

// PNG requests also deliver JSON error bodies as Blobs.
export async function parseAPIError(
  error: unknown,
  fallbackMessage = 'Unable to generate the QR code. Please try again.',
): Promise<APIErrorResponse> {
  const fallback: APIErrorResponse = {
    message:
      error instanceof HttpErrorResponse && error.status === 0
        ? 'Unable to connect. Please check your connection and try again.'
        : fallbackMessage,
    slug:
      error instanceof HttpErrorResponse && error.status === 0
        ? 'network_error'
        : 'unexpected_error',
    details: [],
  };

  let body: unknown = error instanceof HttpErrorResponse ? error.error : error;
  try {
    if (body instanceof Blob) body = await body.text();
    if (typeof body === 'string') body = JSON.parse(body);
  } catch {
    return fallback;
  }

  if (
    !isRecord(body) ||
    typeof body['message'] !== 'string' ||
    !body['message'].trim() ||
    typeof body['slug'] !== 'string'
  )
    return fallback;

  return {
    message: body['message'],
    slug: body['slug'],
    details: Array.isArray(body['details']) ? body['details'].filter(isDetail) : [],
  };
}
