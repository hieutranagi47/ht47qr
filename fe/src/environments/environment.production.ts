export const environment = {
  production: true,
  apiEndpoint: 'https://api.ht47.com/api/',
  apiVersion: 'v1',
  envName: 'prod',
  host: 'https://ht47-qrcode.vercel.app',
  cspConfig: {
    services: ['https://hieutranprofile.herokuapp.com', 'https://hieutranoath2.herokuapp.com'],
    galleries: ['https://hieutranprofile.herokuapp.com'],
    scriptsElm: [],
    frames: ['https://hieutranprofile.herokuapp.com'],
  },
  cspReport: 'https://hieutranprofile.herokuapp.com/csp-report',
  authConfig: {
    issuer: '',
    clientId: '',
    responseType: '',
    scope: '',
  },
};
