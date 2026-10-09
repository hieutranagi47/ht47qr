import { HttpErrorResponse } from '@angular/common/http';
import { parseAPIError } from './api-error';

const response = {
  message: 'Unable to generate the QR code.',
  slug: 'qr_code_generation_failed',
  details: [],
};

describe('parseAPIError', () => {
  it.each([response, JSON.stringify(response), new Blob([JSON.stringify(response)])])(
    'reads the server contract from object, string and Blob bodies',
    async (body) => {
      expect(await parseAPIError(new HttpErrorResponse({ error: body, status: 500 }))).toEqual(
        response,
      );
    },
  );
  it.each(['<html>Bad gateway</html>', new Blob(['not JSON']), null, { message: 'wrong shape' }])(
    'provides a safe fallback for malformed bodies',
    async (body) => {
      const result = await parseAPIError(new HttpErrorResponse({ error: body, status: 502 }));
      expect(result.slug).toBe('unexpected_error');
      expect(result.details).toEqual([]);
      expect(result.message).toContain('Please try again');
    },
  );
  it('explains network failures', async () => {
    expect((await parseAPIError(new HttpErrorResponse({ status: 0 }))).slug).toBe('network_error');
  });
  it('drops malformed details', async () => {
    expect(
      (await parseAPIError({ ...response, details: [null, { message: 'invalid' }] })).details,
    ).toEqual([]);
  });
});
