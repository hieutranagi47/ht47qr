export type QRCodeTextPayload = {
  type: 'text';
  text: string;
};

export type QRCodeWiFiPayload = {
  type: 'wifi';
  wifi_name: string;
  password: string;
};

export type QRCodeSmsPayload = {
  type: 'sms';
  phone_number: string;
  message: string;
};

export type QRCodeEMailPayload = {
  type: 'mail';
  email: string;
  subject: string;
  body: string;
};

export type QRCodeEmail2Payload = {
  type: 'mailto';
  email: string;
  subject: string;
  body: string;
};

export type QRCodeGeoPayload = {
  type: 'geo';
  latitude: string;
  longitude: string;
  label: string;
};

export type QRCodeTelPayload = {
  type: 'tel';
  phone_number: string;
};

export type QRCodeContactPayload = {
  type: 'contact';
  name: string;
  phone: string;
  email: string;
};

export type QRCodeVCardPayload = {
  type: 'vcard';
  name: string;
  phone: string;
  organization: string;
  title: string;
  address: string;
  website: string;
  email: string;
};

export type QRCodeCalendarPayload = {
  type: 'calendar';
  summary: string;
  location: string;
  description: string;
  start_time: string;
  end_time: string;
};

export type QRCodeICalPayload = {
  type: 'ical';
  summary: string;
  start_time: string;
  end_time: string;
};

export type QRCodeDataPayload =
  | QRCodeTextPayload
  | QRCodeWiFiPayload
  | QRCodeSmsPayload
  | QRCodeEMailPayload
  | QRCodeEmail2Payload
  | QRCodeGeoPayload
  | QRCodeTelPayload
  | QRCodeContactPayload
  | QRCodeVCardPayload
  | QRCodeCalendarPayload
  | QRCodeICalPayload;

export type QRCodePayload<T> = {
  data: T;
  logo_img: File | null;
  halftone_image: File | null;
  qr_width: number;
  foreground_color: string;
  border_width: number;
  is_circle_shape: boolean;
  is_custom_shape: boolean;
};

export type APIErrorDetail = {
  entity_type: string;
  entity_id: string;
  error_slug: string;
  message: string;
};

export type APIErrorResponse = {
  message: string;
  slug: string;
  details: APIErrorDetail[];
};

export type QRCodeResponse = Blob;
