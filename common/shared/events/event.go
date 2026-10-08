package events

import "htqrcode/common/module"

// Event identifies the module that owns a domain event.
type Event interface{ Module() module.Name }
