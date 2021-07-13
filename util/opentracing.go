package util

import (
	"github.com/valyala/fasthttp"
)

// Carrier satisfies both TextMapWriter and TextMapReader.
//
// Example usage for server side:
//
//      carrier := opentracefasthttp.New(&ctx.Request.Header)
//      clientContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
//
// Example usage for client side:
//
//     carrier := opentracefasthttp.New(&req.Header)
//     err := tracer.Inject(
//         span.Context(),
//         opentracing.HTTPHeaders,
//         carrier)
//
type Carrier struct {
	h *fasthttp.RequestHeader
}

func NewCarrier(h *fasthttp.RequestHeader) Carrier {
	return Carrier{h: h}
}

// Set conforms to the TextMapWriter interface.
func (c Carrier) Set(key, val string) {
	c.h.Set(key, val)
}

// ForeachKey conforms to the TextMapReader interface.
func (c Carrier) ForeachKey(handler func(key, val string) error) (err error) {
	c.h.VisitAll(func(key, value []byte) {
		if err != nil {
			return
		}

		if err = handler(string(key), string(value)); err != nil {
			return
		}
	})
	return err
}
