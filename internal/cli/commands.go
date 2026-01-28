package cli

import (
	"os"

	"github.com/bastibuck/am-i-home-cli/internal/router"
)

// ListDevices prints devices from the provided RouterClient
func ListDevices(c router.RouterClient) error {
	devs, err := c.ListConnected()
	if err != nil {
		return err
	}

	return PrintStructTable(os.Stdout, devs, []string{"MAC", "IP", "Hostname", "Active"})
}

// activeDevice is a display struct for active devices (without Active field)
type activeDevice struct {
	MAC      string
	IP       string
	Hostname string
}

// ListActive prints only devices marked as active by the router
func ListActive(c router.RouterClient) error {
	devs, err := c.ListConnected()
	if err != nil {
		return err
	}

	var active []activeDevice
	for _, d := range devs {
		if d.Active {
			active = append(active, activeDevice{MAC: d.MAC, IP: d.IP, Hostname: d.Hostname})
		}
	}

	return PrintStructTable(os.Stdout, active, []string{"MAC", "IP", "Hostname"})
}

func CheckByMatcher(c router.RouterClient, matcher string) (bool, error) {
	devs, err := c.ListConnected()
	if err != nil {
		return false, err
	}
	for _, d := range devs {
		if d.Active && (router.MatchMAC(d.MAC, matcher) || d.Hostname == matcher || d.IP == matcher) {
			return true, nil
		}
	}
	return false, nil
}
