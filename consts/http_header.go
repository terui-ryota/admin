package consts

import "go.opencensus.io/plugin/ochttp/propagation/b3"

const (
	HeaderNameAuthorization = "authorization"
	HeaderNameContentType   = "content-type"
	HeaderNameUserAgent     = "user-agent"
	HeaderNameXB3TraceID    = b3.TraceIDHeader
	HeaderNameXB3SpanID     = b3.SpanIDHeader
	HeaderNameXB3Sampled    = b3.SampledHeader

	HeaderNameFeature             = "x-feature"
	HeaderNameFeatureVia          = "x-feature-via"
	HeaderNameNantesDevelopmentID = "x-nantes-development-id"

	HeaderNameXClientIPAddress = "x-client-ip-address"
	HeaderNameXClientUserAgent = "x-client-user-agent"
)

const (
	MediaTypeApplicationJson           = "application/json"
	MediaTypeApplicationFormURLEncoded = "application/x-www-form-urlencoded"
)
