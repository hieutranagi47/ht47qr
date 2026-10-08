package events

import "htqrcode/common/module"

type QRCodeCreated struct {
	UUID string `json:"uuid"`
}

func (QRCodeCreated) Module() module.Name { return module.QRCode }
