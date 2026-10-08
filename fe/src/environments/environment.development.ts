export const environment = {
  production: false,
  apiEndpoint: 'http://localhost:6789/api/',
  apiVersion: 'v1',
  envName: 'local',
  host: 'http://localhost:4200',
  cspConfig: {
    services: ['http://localhost:6789'],
    galleries: ['http://localhost:6789'],
    scriptsElm: [],
    frames: ['http://localhost:6789'],
  },
  cspReport: 'http://localhost:6789/csp-report',
  authConfig: {
    issuer: '',
    clientId: '',
    responseType: '',
    scope: '',
  },
};
