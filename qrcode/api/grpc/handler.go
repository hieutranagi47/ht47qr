package grpcapi

import (
	"context"
	"errors"
	"net/http"
	"regexp"

	"google.golang.org/protobuf/proto"
	"htqrcode/common"
	transport "htqrcode/common/grpc"
	"htqrcode/qrcode/api/input"
	"htqrcode/qrcode/app"
	"htqrcode/qrcode/domain"
)

type Handler struct {
	UnimplementedQRCodeServiceServer
	Generator app.Generator
}

var (
	_        QRCodeServiceServer = Handler{}
	hexColor                     = regexp.MustCompile(`^#?[0-9a-fA-F]{6}$`)
)

func Register(s *transport.Server, generator app.Generator) {
	RegisterQRCodeServiceServer(s, Handler{Generator: generator})
}

func invalid(message, field string) error {
	return common.NewInvalidInputError("invalid_input", "%s", message).WithDetails([]common.ErrorDetails{{
		EntityType: "request_field", EntityID: field, ErrorSlug: "invalid_input", Message: message,
	}})
}

func (h Handler) GenerateQRCode(ctx context.Context, r *GenerateQRCodeRequest) (*GenerateQRCodeResponse, error) {
	if r == nil || r.Message == nil {
		return nil, invalid("message is required", "message")
	}
	if proto.Size(r) > transport.MaxRequestSize {
		return nil, invalid("request exceeds 8 MiB", "request")
	}
	m := r.Message
	message := domain.MessageRequest{
		Type:         domain.MessageType(m.Type),
		WifiName:     m.WifiName,
		Password:     m.Password,
		PhoneNumber:  m.PhoneNumber,
		Message:      m.Message,
		Email:        m.Email,
		Subject:      m.Subject,
		Body:         m.Body,
		Latitude:     m.Latitude,
		Longitude:    m.Longitude,
		Label:        m.Label,
		Name:         m.Name,
		Phone:        m.Phone,
		Organization: m.Organization,
		Title:        m.Title,
		Address:      m.Address,
		Website:      m.Website,
		Summary:      m.Summary,
		Location:     m.Location,
		Description:  m.Description,
		StartTime:    m.StartTime,
		EndTime:      m.EndTime,
		Text:         m.Text,
	}
	if _, err := message.Payload(); err != nil {
		return nil, invalid(err.Error(), "message")
	}
	options := app.Options{Width: 10, Border: 10}
	if o := r.Options; o != nil {
		if o.QrWidth != nil {
			if *o.QrWidth < 6 || *o.QrWidth > 255 {
				return nil, invalid("qr_width must be between 6 and 255", "qr_width")
			}
			options.Width = uint8(*o.QrWidth)
		}
		if o.BorderWidth != nil {
			if *o.BorderWidth > 1024 {
				return nil, invalid("border_width must be between 0 and 1024", "border_width")
			}
			options.Border = int(*o.BorderWidth)
		}
		options.Circle, options.Custom, options.Foreground = o.IsCircleShape, o.IsCustomShape, o.ForegroundColor
		if options.Foreground != "" && !hexColor.MatchString(options.Foreground) {
			return nil, invalid("foreground_color must be a six-digit hex color", "foreground_color")
		}
		if len(o.LogoImg) != 0 {
			img, err := input.DecodeImage(o.LogoImg, 1<<20, 500)
			if err != nil {
				return nil, invalid(err.Error(), "logo_img")
			}
			options.Logo = img
		}
		if len(o.HalftoneImg) != 0 {
			if _, err := input.DecodeImage(o.HalftoneImg, 5<<20, 4096); err != nil {
				return nil, invalid(err.Error(), "halftone_img")
			}
			options.Halftone = o.HalftoneImg
		}
	}
	data, err := h.Generator.Generate(ctx, message, options)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, common.Error{HttpErrorCode: http.StatusInternalServerError, PublicError: "failed to generate qr code", ErrorSlug: "qr_code_generation_failed", InternalError: err}
	}
	return &GenerateQRCodeResponse{Image: data, ContentType: "image/png"}, nil
}
