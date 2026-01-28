package router

// Device represents a device connected to the router
type Device struct {
	MAC      string
	IP       string
	Hostname string
	Active   bool
}

// RouterClient abstracts fetching connected devices
type RouterClient interface {
	ListConnected() ([]Device, error)
}

// normalizeMAC returns a canonical MAC format used for comparisons:
// lowercase with no separators.
func normalizeMAC(mac string) string {
	b := make([]byte, 0, len(mac))
	for i := 0; i < len(mac); i++ {
		c := mac[i]
		if c == ':' || c == '-' || c == '.' || c == ' ' {
			continue
		}
		// uppercase to lowercase
		if c >= 'A' && c <= 'Z' {
			c = c - 'A' + 'a'
		}
		b = append(b, c)
	}
	return string(b)
}

// MatchMAC compares two MAC addresses for equality after normalization
func MatchMAC(a, b string) bool {
	return normalizeMAC(a) == normalizeMAC(b)
}
