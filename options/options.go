package options

import (
	"github.com/jessevdk/go-flags"
)

var Opts struct {
	Bgp                    bool           `short:"b" long:"bgp" description:"Connect and Sync with GoBGP server"`
	Version                bool           `short:"v" long:"version" description:"Show loxilb version"`
	NoApi                  bool           `short:"a" long:"api" description:"Run Rest API server"`
	NoNlp                  bool           `short:"n" long:"nonlp" description:"Do not register with nlp"`
	NoExporter             bool           `short:"e" long:"noexporter" description:"Do not register flowExporter"`
	Host                   string         `long:"host" description:"the IP to listen on" default:"localhost" env:"HOST"`
	Port                   int            `long:"port" description:"the port to listen on for insecure connections" default:"11111" env:"PORT"`
	TLSHost                string         `long:"tls-host" description:"the IP to listen on for tls, when not specified it's the same as --host" env:"TLS_HOST"`
	TLSPort                int            `long:"tls-port" description:"the port to listen on for secure connections" default:"8091" env:"TLS_PORT"`
	TLSCertificate         flags.Filename `long:"tls-certificate" description:"the certificate to use for secure connections" default:"/opt/loxilb/cert/server.crt" env:"TLS_CERTIFICATE"`
	TLSCertificateKey      flags.Filename `long:"tls-key" description:"the private key to use for secure connections" default:"/opt/loxilb/cert/server.key" env:"TLS_PRIVATE_KEY"`
	FlowCollectorAddr      string         `long:"flow-collector-address" description:"Flow exporter collector IPv4 Address" default:"127.0.0.1:30473"`
	FlowCollectorProto     string         `long:"flow-collector-proto" description:"Flow exporter collector protocol" default:"tcp"`
	ActiveFlowTimeout      int            `long:"active-flow-timeout" description:"Flow exporter collector active time out" default:"10"`
	IdleFlowTimeout        int            `long:"idle-flow-timeout" description:"Flow exporter collector idle time out" default:"10"`
	StaleConnectionTimeout int            `long:"stale-connection-timeout" description:"Flow exporter collector stale connection time out" default:"10"`
	PollInterval           int            `long:"poll-interval" description:"Flow exporter collector polling interval time" default:"10"`
	ConnectUplinkToBridge  bool           `long:"connect-uplink-bridge" description:"Flow exporter collector uplink to bridge"`
}
