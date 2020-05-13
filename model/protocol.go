package model

const (
	HTTP Protocol = iota + 1
	HTTPS
	RDP
	SSH
	VNC
	TELNET
)

// Protocol represents an application-layer protocol.
type Protocol int

// SupportedProtocols is the list of currently supported protocols.
var SupportedProtocols = []struct {
	Id          Identity  `json:"id"`
	Name        string    `json:"name"`
	DefaultPort int       `json:"default_port"`
	SSOTypes    []SSOType `json:"sso_types"`
}{
	{Identity(HTTP), "HTTP", 80, []SSOType{None, Basic, Forms}},
	{Identity(HTTPS), "HTTPS", 443, []SSOType{None, Basic, Forms}},
	{Identity(SSH), "SSH", 22, []SSOType{None, Basic, PrivateKey}},
	{Identity(TELNET), "TELNET", 23, []SSOType{None, Basic}},
	{Identity(RDP), "RDP", 3389, []SSOType{None, Basic}},
	{Identity(VNC), "VNC", 5900, []SSOType{None, Basic}},
}

// String returns the protocol in lowercase string representation.
func (protocol Protocol) String() string {
	switch protocol {
	case HTTP:
		return "HTTP"
	case HTTPS:
		return "HTTPS"
	case RDP:
		return "RDP"
	case SSH:
		return "SSH"
	case VNC:
		return "VNC"
	case TELNET:
		return "TELNET"
	default:
		return "UNKNOWN"
	}
}

// Textual returns true if the protocol is a textual protocol.
func (protocol Protocol) Textual() bool {
	switch protocol {
	case HTTP, HTTPS, SSH, TELNET:
		return true
	}
	return false
}
