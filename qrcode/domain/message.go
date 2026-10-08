package domain

import "fmt"

type MessageType string

const (
	MessageTypeWifi     MessageType = "wifi"
	MessageTypeSms      MessageType = "sms"
	MessageTypeMail     MessageType = "mail"
	MessageTypeMailTo   MessageType = "mailto"
	MessageTypeGeo      MessageType = "geo"
	MessageTypeTel      MessageType = "tel"
	MessageTypeContact  MessageType = "contact"
	MessageTypeVCard    MessageType = "vcard"
	MessageTypeCalendar MessageType = "calendar"
	MessageTypeICal     MessageType = "ical"
	MessageTypeText     MessageType = "text"
)

type MessageRequest struct {
	Type         MessageType
	WifiName     string
	Password     string
	PhoneNumber  string
	Message      string
	Email        string
	Subject      string
	Body         string
	Latitude     string
	Longitude    string
	Label        string
	Name         string
	Phone        string
	Organization string
	Title        string
	Address      string
	Website      string
	Summary      string
	Location     string
	Description  string
	StartTime    string
	EndTime      string
	Text         string
}

func WifiData(wifiName string, password string) string {
	return fmt.Sprintf("WIFI:T:WPA;S:%s;P:%s;;", wifiName, password)
}

func SmsData(phoneNumber string, message string) string {
	return fmt.Sprintf("SMSTO:%s:%s", phoneNumber, message)
}

func MailData(email string, subject string, body string) string {
	return fmt.Sprintf("MATMSG:TO:%s;SUB:%s;BODY:%s;;", email, subject, body)
}

func MailDataTo(email string, subject string, body string) string {
	return fmt.Sprintf("mailto:%s?subject=%s&body=%s;;", email, subject, body)
}

func GeoData(latitude string, longitude string, label string) string {
	str := fmt.Sprintf("GEO:%s,%s", latitude, longitude)

	if label != "" {
		str += fmt.Sprintf("?q=%s", label)
	}

	return str
}

func TelData(phoneNumber string) string {
	return fmt.Sprintf("TEL:%s", phoneNumber)
}

func ContactData(name string, phone string, email string) string {
	return fmt.Sprintf("MECARD:N:%s;TEL:%s;EMAIL:%s;;", name, phone, email)
}

func VCardData(name string, phone string, email string, organization string, title string, address string, website string) string {
	return fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nN:%s\nORG:%s\nTITLE:%s\nTEL:%s\nEMAIL:%s\nADR:%s\nURL:%s\nEND:VCARD", name, organization, title, phone, email, address, website)
}

func CalendarData(summary string, location string, description string, startTime string, endTime string) string {
	return fmt.Sprintf("BEGIN:VCALENDAR\nVERSION:2.0\nBEGIN:VEVENT\nSUMMARY:%s\nLOCATION:%s\nDESCRIPTION:%s\nDTSTART:%s\nDTEND:%s\nEND:VEVENT\nEND:VCALENDAR", summary, location, description, startTime, endTime)
}

func ICalData(summary string, startDate string, endDate string) string {
	return fmt.Sprintf("BEGIN:EVENT\nSUMMARY:%s\nDTSTART:%s\nDTEND:%s\nEND:EVENT", summary, startDate, endDate)
}

// Payload validates the fields required by the selected message type.
// Unknown types retain the legacy text fallback.
func (m MessageRequest) Payload() (string, error) {
	required := func(values ...string) bool {
		for _, v := range values {
			if v == "" {
				return false
			}
		}
		return true
	}
	switch m.Type {
	case MessageTypeWifi:
		if !required(m.WifiName, m.Password) {
			return "", fmt.Errorf("missing required fields for %s type", m.Type)
		}
		return WifiData(m.WifiName, m.Password), nil
	case MessageTypeSms:
		if !required(m.PhoneNumber, m.Message) {
			return "", fmt.Errorf("missing required fields for %s type", m.Type)
		}
		return SmsData(m.PhoneNumber, m.Message), nil
	case MessageTypeMail:
		if !required(m.Email, m.Subject, m.Body) {
			return "", fmt.Errorf("missing required fields for %s type", m.Type)
		}
		return MailData(m.Email, m.Subject, m.Body), nil
	case MessageTypeMailTo:
		if !required(m.Email, m.Subject, m.Body) {
			return "", fmt.Errorf("missing required fields for %s type", m.Type)
		}
		return MailDataTo(m.Email, m.Subject, m.Body), nil
	case MessageTypeGeo:
		if !required(m.Latitude, m.Longitude) {
			return "", fmt.Errorf("missing required fields for %s type", m.Type)
		}
		return GeoData(m.Latitude, m.Longitude, m.Label), nil
	case MessageTypeTel:
		if !required(m.PhoneNumber) {
			return "", fmt.Errorf("missing required fields for %s type", m.Type)
		}
		return TelData(m.PhoneNumber), nil
	case MessageTypeContact:
		if !required(m.Name, m.Phone, m.Email) {
			return "", fmt.Errorf("missing required fields for %s type", m.Type)
		}
		return ContactData(m.Name, m.Phone, m.Email), nil
	case MessageTypeVCard:
		if !required(m.Name, m.Phone, m.Email, m.Organization, m.Title, m.Address, m.Website) {
			return "", fmt.Errorf("missing required fields for %s type", m.Type)
		}
		return VCardData(m.Name, m.Phone, m.Email, m.Organization, m.Title, m.Address, m.Website), nil
	case MessageTypeCalendar:
		if !required(m.Summary, m.StartTime, m.EndTime) {
			return "", fmt.Errorf("missing required fields for %s type", m.Type)
		}
		return CalendarData(m.Summary, m.Location, m.Description, m.StartTime, m.EndTime), nil
	case MessageTypeICal:
		if !required(m.Summary, m.StartTime, m.EndTime) {
			return "", fmt.Errorf("missing required fields for %s type", m.Type)
		}
		return ICalData(m.Summary, m.StartTime, m.EndTime), nil
	default:
		if m.Text == "" {
			return "", fmt.Errorf("text is required for text type")
		}
		return m.Text, nil
	}
}
