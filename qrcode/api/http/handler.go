package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"mime"
	"mime/multipart"
	nethttp "net/http"
	"regexp"
	"strconv"

	"htqrcode/common"
	"htqrcode/qrcode/api/input"
	"htqrcode/qrcode/app"
	"htqrcode/qrcode/domain"

	"github.com/labstack/echo/v5"
)

type Handler struct{ generator app.Generator }

var _ StrictServerInterface = Handler{}

func Register(router common.EchoRouter, generator app.Generator, prefix string) {
	strict := NewStrictHandler(Handler{generator: generator}, nil)

	RegisterHandlersWithBaseURL(
		router,
		multipartServer{
			ServerInterface: strict,
		},
		prefix,
	)
}

// multipartServer validates multipart binding before invoking the strict handler.
type multipartServer struct{ ServerInterface }

func (s multipartServer) PostV1GenerateQrCode(c *echo.Context) error {
	c.Request().Body = nethttp.MaxBytesReader(c.Response(), c.Request().Body, 8<<20)
	mediaType, params, err := mime.ParseMediaType(c.Request().Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/form-data" || params["boundary"] == "" {
		return invalidInput("invalid multipart request", "data").WithInternalError(err)
	}
	return s.ServerInterface.PostV1GenerateQrCode(c)
}

func (Handler) GetHealth(context.Context, GetHealthRequestObject) (GetHealthResponseObject, error) {
	return GetHealth200TextResponse("OK"), nil
}

func (Handler) GetV1VeryFirstApi(context.Context, GetV1VeryFirstApiRequestObject) (GetV1VeryFirstApiResponseObject, error) {
	return GetV1VeryFirstApi200TextResponse("My Handler v1"), nil
}

func invalidInput(message, field string) common.Error {
	return common.NewInvalidInputError("invalid_input", "%s", message).WithDetails([]common.ErrorDetails{{
		EntityType: "form_field",
		EntityID:   field,
		ErrorSlug:  "invalid_input",
		Message:    message,
	}})
}

func bad(message, field string) (PostV1GenerateQrCodeResponseObject, error) {
	return nil, invalidInput(message, field)
}

func (h Handler) PostV1GenerateQrCode(ctx context.Context, request PostV1GenerateQrCodeRequestObject) (PostV1GenerateQrCodeResponseObject, error) {
	form, err := request.Body.ReadForm(6 << 20)
	if err != nil {
		return bad("invalid multipart request", "data")
	}
	defer form.RemoveAll()
	value := func(name string) string {
		v := form.Value[name]
		if len(v) == 0 {
			return ""
		}
		return v[0]
	}
	if value("data") == "" {
		return bad("data is required", "data")
	}
	var body MessageRequest
	if err := json.Unmarshal([]byte(value("data")), &body); err != nil {
		return bad("invalid data format", "data")
	}
	message := domain.MessageRequest{
		Type:         domain.MessageType(body.Type),
		WifiName:     body.WifiName,
		Password:     body.Password,
		PhoneNumber:  body.PhoneNumber,
		Message:      body.Message,
		Email:        body.Email,
		Subject:      body.Subject,
		Body:         body.Body,
		Latitude:     body.Latitude,
		Longitude:    body.Longitude,
		Label:        body.Label,
		Name:         body.Name,
		Phone:        body.Phone,
		Organization: body.Organization,
		Title:        body.Title,
		Address:      body.Address,
		Website:      body.Website,
		Summary:      body.Summary,
		Location:     body.Location,
		Description:  body.Description,
		StartTime:    body.StartTime,
		EndTime:      body.EndTime,
		Text:         body.Text,
	}
	if _, err := message.Payload(); err != nil {
		return bad(err.Error(), "data")
	}
	options := app.Options{Width: 10, Border: 10}
	for _, param := range []struct {
		name     string
		min, max int
		target   *int
	}{
		{"qr_width", 6, 255, nil}, {"border_width", 0, 1024, &options.Border},
	} {
		if raw := value(param.name); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil || n < param.min || n > param.max {
				return bad(fmt.Sprintf("%s must be between %d and %d", param.name, param.min, param.max), param.name)
			}
			if param.target == nil {
				options.Width = uint8(n)
			} else {
				*param.target = n
			}
		}
	}
	for _, param := range []struct {
		name   string
		target *bool
	}{{"is_circle_shape", &options.Circle}, {"is_custom_shape", &options.Custom}} {
		if raw := value(param.name); raw != "" {
			b, err := strconv.ParseBool(raw)
			if err != nil {
				return bad("invalid boolean value", param.name)
			}
			*param.target = b
		}
	}
	options.Foreground = value("foreground_color")
	if options.Foreground != "" && !hexColor.MatchString(options.Foreground) {
		return bad("foreground_color must be a six-digit hex color", "foreground_color")
	}
	for _, name := range []string{"logo_img", "halftone_img"} {
		files := form.File[name]
		if len(files) == 0 {
			continue
		}
		limit, maxDimension := int64(5<<20), 4096
		if name == "logo_img" {
			limit, maxDimension = 1<<20, 500
		}
		data, img, err := readImage(files[0], limit, maxDimension)
		if err != nil {
			return bad(err.Error(), name)
		}
		if name == "logo_img" {
			options.Logo = img
		} else {
			options.Halftone = data
		}
	}
	data, err := h.generator.Generate(ctx, message, options)
	if err != nil {
		return nil, common.Error{
			HttpErrorCode: nethttp.StatusInternalServerError,
			PublicError:   "failed to generate qr code",
			ErrorSlug:     "qr_code_generation_failed",
			InternalError: err,
		}
	}
	return PostV1GenerateQrCode200ImagepngResponse{Body: bytes.NewReader(data), ContentLength: int64(len(data))}, nil
}

var hexColor = regexp.MustCompile(`^#?[0-9a-fA-F]{6}$`)

func readImage(file *multipart.FileHeader, limit int64, maxDimension int) ([]byte, image.Image, error) {
	if file.Size > limit {
		return nil, nil, fmt.Errorf("image exceeds maximum file size")
	}
	f, err := file.Open()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot open image")
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil || int64(len(data)) > limit {
		return nil, nil, fmt.Errorf("cannot read image within size limit")
	}
	img, err := input.DecodeImage(data, int(limit), maxDimension)
	if err != nil {
		return nil, nil, err
	}
	return data, img, nil
}
