// Package contracts lives in a sub-package of common/module (not in common/module itself)
// to break the import cycle: common/module/module.go imports this package,
// and each module's module.go implements RegisterContracts — which means
// common/module cannot also define contracts (that would create a cycle).
package contracts

import (
	"errors"

	qrcode "htqrcode/qrcode/api/module/client"
	shortenURL "htqrcode/shorten_url/api/module/client"
)

type Contracts struct {
	qrcode.QRCode
	shortenURL.ShortenURL
}

func (c *Contracts) Verify() error {
	var err error
	if c.QRCode == nil {
		err = errors.Join(err, errors.New("qrcode module contract is empty"))
	}

	return err
}
