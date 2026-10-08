export const environment = {
  production: false,
  apiEndpoint: 'https://localhost:1443/api/',
  apiVersion: 'v1',
  envName: 'local',
  host: 'http://localhost:4200',
  cspConfig: {
    services: ['https://localhost:1443'],
    galleries: ['https://localhost:1443'],
    scriptsElm: [],
    frames: ['https://localhost:1443'],
  },
  cspReport: 'https://localhost:1443/csp-report',
  authConfig: {
    issuer: '',
    clientId: '',
    responseType: '',
    scope: '',
  },
};
