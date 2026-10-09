export const environment = {
  production: false,
  apiEndpoint: 'https://localhost:8334/qrcode/api/',
  apiVersion: 'v1',
  envName: 'local',
  host: 'http://localhost:4200',
  cspConfig: {
    services: ['https://localhost:4200'],
    galleries: ['https://localhost:4200'],
    scriptsElm: [],
    frames: ['https://localhost:4200'],
  },
  cspReport: 'https://localhost:4200/csp-report',
  authConfig: {
    issuer: '',
    clientId: '',
    responseType: '',
    scope: '',
  },
};
