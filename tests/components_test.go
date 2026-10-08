package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/makiuchi-d/gozxing"
	decoder "github.com/makiuchi-d/gozxing/qrcode"
	"github.com/stretchr/testify/require"
	"htqrcode"
	"htqrcode/common"
)

func startService(t *testing.T) *httptest.Server {
	t.Helper()
	service, err := htqrcode.New(context.Background(), htqrcode.ExternalServices{})
	require.NoError(t, err)
	server := httptest.NewServer(service.HTTPHandler())
	t.Cleanup(server.Close)
	return server
}

func generate(t *testing.T, server *httptest.Server, fields map[string]string, files map[string][]byte) *http.Response {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for k, v := range fields {
		require.NoError(t, writer.WriteField(k, v))
	}
	for k, v := range files {
		part, err := writer.CreateFormFile(k, "untrusted.png")
		require.NoError(t, err)
		_, err = part.Write(v)
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	response, err := server.Client().Post(server.URL+"/api/v1/generate-qrcode", writer.FormDataContentType(), &body)
	require.NoError(t, err)
	t.Cleanup(func() { response.Body.Close() })
	return response
}

func decodePNG(t *testing.T, response *http.Response) image.Image {
	t.Helper()
	require.Equal(t, 200, response.StatusCode)
	require.Equal(t, "image/png", response.Header.Get("Content-Type"))
	img, err := png.Decode(response.Body)
	require.NoError(t, err)
	return img
}

func TestHealthAndGreeting(t *testing.T) {
	server := startService(t)
	for path, expected := range map[string]string{"/health": "OK", "/api/v1/very-first-api": "My Handler v1"} {
		t.Run(path, func(t *testing.T) {
			response, err := server.Client().Get(server.URL + path)
			require.NoError(t, err)
			defer response.Body.Close()
			require.Equal(t, 200, response.StatusCode)
			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.Equal(t, expected, string(body))
		})
	}
}

var qrMessageCases = []struct{ name, data, payload string }{
	{"text", `{"type":"text","text":"hello"}`, "hello"},
	{"default", `{"text":"fallback"}`, "fallback"},
	{"unknown", `{"type":"legacy","text":"fallback"}`, "fallback"},
	{"wifi", `{"type":"wifi","wifi_name":"network","password":"secret"}`, "WIFI:T:WPA;S:network;P:secret;;"},
	{"sms", `{"type":"sms","phone_number":"123","message":"hello"}`, "SMSTO:123:hello"},
	{"mail", `{"type":"mail","email":"a@b.com","subject":"hi","body":"hello"}`, "MATMSG:TO:a@b.com;SUB:hi;BODY:hello;;"},
	{"mailto", `{"type":"mailto","email":"a@b.com","subject":"hi","body":"hello"}`, "mailto:a@b.com?subject=hi&body=hello;;"},
	{"geo", `{"type":"geo","latitude":"10","longitude":"20","label":"home"}`, "GEO:10,20?q=home"},
	{"tel", `{"type":"tel","phone_number":"123"}`, "TEL:123"},
	{"contact", `{"type":"contact","name":"Jo","phone":"123","email":"a@b.com"}`, "MECARD:N:Jo;TEL:123;EMAIL:a@b.com;;"},
	{"vcard", `{"type":"vcard","name":"Jo","phone":"123","email":"a@b.com","organization":"Org","title":"Dev","address":"Home","website":"example.com"}`, "BEGIN:VCARD\nVERSION:3.0\nN:Jo\nORG:Org\nTITLE:Dev\nTEL:123\nEMAIL:a@b.com\nADR:Home\nURL:example.com\nEND:VCARD"},
	{"calendar", `{"type":"calendar","summary":"Meet","location":"Home","description":"Hi","start_time":"20261008T090000","end_time":"20261008T100000"}`, "BEGIN:VCALENDAR\nVERSION:2.0\nBEGIN:VEVENT\nSUMMARY:Meet\nLOCATION:Home\nDESCRIPTION:Hi\nDTSTART:20261008T090000\nDTEND:20261008T100000\nEND:VEVENT\nEND:VCALENDAR"},
	{"ical", `{"type":"ical","summary":"Meet","start_time":"20261008T090000","end_time":"20261008T100000"}`, "BEGIN:EVENT\nSUMMARY:Meet\nDTSTART:20261008T090000\nDTEND:20261008T100000\nEND:EVENT"},
}

func TestQRMessageTypes(t *testing.T) {
	server := startService(t)

	for _, tc := range qrMessageCases {
		t.Run(tc.name, func(t *testing.T) {
			img := decodePNG(t, generate(t, server, map[string]string{"data": tc.data}, nil))
			bitmap, err := gozxing.NewBinaryBitmapFromImage(img)
			require.NoError(t, err)
			decoded, err := decoder.NewQRCodeReader().Decode(bitmap, nil)
			require.NoError(t, err)
			require.Equal(t, tc.payload, decoded.GetText())
		})
	}
}

func testImage(t *testing.T, size int, jpegFormat bool) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, color.RGBA{R: 90, G: 100, B: 150, A: 255})
		}
	}
	var data bytes.Buffer
	var err error
	if jpegFormat {
		err = jpeg.Encode(&data, img, nil)
	} else {
		err = png.Encode(&data, img)
	}
	require.NoError(t, err)
	return data.Bytes()
}

func TestQROptions(t *testing.T) {
	server := startService(t)
	for _, shape := range []string{"square", "is_circle_shape", "is_custom_shape"} {
		t.Run(shape, func(t *testing.T) {
			fields := map[string]string{"data": `{"text":"hello"}`, "qr_width": "12", "border_width": "0", "foreground_color": "#123456"}
			if shape != "square" {
				fields[shape] = "true"
			}
			img := decodePNG(t, generate(t, server, fields, nil))
			require.Equal(t, 252, img.Bounds().Dx())
			require.Equal(t, 252, img.Bounds().Dy())
			found := false
			for y := 0; y < img.Bounds().Dy() && !found; y++ {
				for x := 0; x < img.Bounds().Dx(); x++ {
					r, g, b, _ := img.At(x, y).RGBA()
					if r == 0x1212 && g == 0x3434 && b == 0x5656 {
						found = true
						break
					}
				}
			}
			require.True(t, found, "foreground color must affect QR pixels")
		})
	}
	for _, format := range []bool{false, true} {
		for _, field := range []string{"logo_img", "halftone_img"} {
			t.Run(field+map[bool]string{false: "png", true: "jpeg"}[format], func(t *testing.T) {
				decodePNG(t, generate(t, server, map[string]string{"data": `{"text":"hello"}`}, map[string][]byte{field: testImage(t, 20, format)}))
			})
		}
	}
}

func assertBad(t *testing.T, response *http.Response, field string) {
	t.Helper()
	require.Equal(t, 400, response.StatusCode)
	require.Contains(t, response.Header.Get("Content-Type"), "application/json")
	body := decodeError(t, response)
	require.Equal(t, "invalid_input", body.Slug)
	require.NotEmpty(t, body.Message)
	require.Equal(t, []common.HttpErrorDetail{{
		EntityType: "form_field",
		EntityID:   field,
		ErrorSlug:  "invalid_input",
		Message:    body.Message,
	}}, body.Details)
}

func TestQRValidation(t *testing.T) {
	server := startService(t)
	for _, data := range []string{"", "{", "null", "[]", `{"text":""}`} {
		t.Run("data/"+data, func(t *testing.T) { assertBad(t, generate(t, server, map[string]string{"data": data}, nil), "data") })
	}
	for _, typ := range []string{"wifi", "sms", "mail", "mailto", "geo", "tel", "contact", "vcard", "calendar", "ical"} {
		t.Run("missing/"+typ, func(t *testing.T) {
			assertBad(t, generate(t, server, map[string]string{"data": `{"type":"` + typ + `"}`}, nil), "data")
		})
	}
	for field, values := range map[string][]string{"qr_width": {"5", "0", "-1", "256", "abc"}, "border_width": {"-1", "1025", "abc"}, "is_circle_shape": {"no"}, "is_custom_shape": {"no"}, "foreground_color": {"oops", "#12345", "#12345678"}} {
		for _, value := range values {
			t.Run(field+"/"+value, func(t *testing.T) {
				assertBad(t, generate(t, server, map[string]string{"data": `{"text":"hello"}`, field: value}, nil), field)
			})
		}
	}
	for _, field := range []string{"logo_img", "halftone_img"} {
		for name, data := range map[string][]byte{"invalid": []byte("not an image"), "truncated": testImage(t, 20, false)[:30], "oversize": bytes.Repeat([]byte("x"), map[string]int{"logo_img": (1 << 20) + 1, "halftone_img": (5 << 20) + 1}[field])} {
			t.Run(field+"/"+name, func(t *testing.T) {
				assertBad(t, generate(t, server, map[string]string{"data": `{"text":"hello"}`}, map[string][]byte{field: data}), field)
			})
		}
	}
	assertBad(t, generate(t, server, map[string]string{"data": `{"text":"hello"}`}, map[string][]byte{"logo_img": testImage(t, 501, false)}), "logo_img")
	for _, contentType := range []string{"application/json", "multipart/form-data", "multipart/form-data; boundary=broken", "multipart/form-data; boundary=\"broken"} {
		response, err := server.Client().Post(server.URL+"/api/v1/generate-qrcode", contentType, strings.NewReader("invalid"))
		require.NoError(t, err)
		defer response.Body.Close()
		assertBad(t, response, "data")
	}
}

func TestQRRenderingFailure(t *testing.T) {
	server := startService(t)
	response := generate(t, server, map[string]string{"data": `{"text":"` + strings.Repeat("a", 4000) + `"}`}, nil)
	require.Equal(t, 500, response.StatusCode)
	body := decodeError(t, response)
	require.Equal(t, "qr_code_generation_failed", body.Slug)
	require.Equal(t, "failed to generate qr code", body.Message)
	require.Empty(t, body.Details)
}

func TestSSEHeartbeat(t *testing.T) {
	service, err := htqrcode.New(context.Background(), htqrcode.ExternalServices{})
	require.NoError(t, err)
	server := httptest.NewServer(service.SSEHandler())
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/sse/qrcode/heartbeat", nil)
	require.NoError(t, err)
	response, err := server.Client().Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, 200, response.StatusCode)
	require.Contains(t, response.Header.Get("Content-Type"), "text/event-stream")
	data := make([]byte, len(": connected\n\n"))
	_, err = io.ReadFull(response.Body, data)
	require.NoError(t, err)
	require.Equal(t, ": connected\n\n", string(data))
	cancel()
}

func TestQRTemporaryFilesAreRemoved(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TMPDIR", dir)
	server := startService(t)
	for _, fields := range []map[string]string{
		{"data": `{"text":"hello"}`},
		{"data": `{"text":"hello"}`, "is_circle_shape": "true"},
		{"data": `{"text":"` + strings.Repeat("a", 1000) + `"}`, "qr_width": "255"},
	} {
		response := generate(t, server, fields, map[string][]byte{"halftone_img": testImage(t, 20, false)})
		_, err := io.Copy(io.Discard, response.Body)
		require.NoError(t, err)
		if fields["qr_width"] == "255" {
			require.Equal(t, 500, response.StatusCode)
		} else {
			require.Equal(t, 200, response.StatusCode)
		}
		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		require.Empty(t, entries)
	}
}

// decodeError checks the exact public contract, including a non-null details array.
func decodeError(t *testing.T, response *http.Response) common.HttpErrorResponse {
	t.Helper()
	require.Contains(t, response.Header.Get("Content-Type"), "application/json")
	data, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	var fields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &fields))
	require.Len(t, fields, 3)
	require.Contains(t, fields, "message")
	require.Contains(t, fields, "slug")
	require.Contains(t, fields, "details")
	var body common.HttpErrorResponse
	require.NoError(t, json.Unmarshal(data, &body))
	require.NotNil(t, body.Details)
	return body
}

func TestRoutingErrorsUseSharedContract(t *testing.T) {
	service, err := htqrcode.New(context.Background(), htqrcode.ExternalServices{})
	require.NoError(t, err)
	for _, router := range []struct {
		name    string
		handler http.Handler
		paths   []string
	}{
		{"http", service.HTTPHandler(), []string{"/health", "/api/v1/very-first-api", "/api/v1/generate-qrcode"}},
		{"sse", service.SSEHandler(), []string{"/health", "/sse/qrcode/heartbeat"}},
	} {
		t.Run(router.name, func(t *testing.T) {
			server := httptest.NewServer(router.handler)
			t.Cleanup(server.Close)
			for _, path := range append(router.paths, "/missing") {
				t.Run(path, func(t *testing.T) {
					request, err := http.NewRequest(http.MethodDelete, server.URL+path, nil)
					require.NoError(t, err)
					response, err := server.Client().Do(request)
					require.NoError(t, err)
					defer response.Body.Close()
					status := http.StatusMethodNotAllowed
					if path == "/missing" {
						status = http.StatusNotFound
					}
					require.Equal(t, status, response.StatusCode)
					body := decodeError(t, response)
					require.Equal(t, http.StatusText(status), body.Message)
					require.Equal(t, strings.ToLower(strings.ReplaceAll(body.Message, " ", "_")), body.Slug)
					require.Empty(t, body.Details)
				})
			}
		})
	}
}

func TestQROversizedMultipartUsesSharedContract(t *testing.T) {
	server := startService(t)
	assertBad(t, generate(t, server, map[string]string{"data": strings.Repeat("x", 8<<20)}, nil), "data")
}
