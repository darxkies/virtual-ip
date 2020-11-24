package pkg

import (
	"net"

	"github.com/j-keck/arping"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
)

type NetworkConfigurator interface {
	AddIP() error
	DeleteIP() error
	IsSet() (bool, error)
	IP() string
	Interface() string
}

type NetlinkNetworkConfigurator struct {
	address    *netlink.Addr
	link       netlink.Link
	_interface string
	ip         net.IP
}

func NewNetlinkNetworkConfigurator(_address, _interface string) (result NetlinkNetworkConfigurator, error error) {
	cidr := _address + "/32"

	ip, _, error := net.ParseCIDR(cidr)
	if error != nil {
		error = errors.Wrapf(error, "could not parse ip '%s'", _address)

		return
	}

	result = NetlinkNetworkConfigurator{_interface: _interface, ip: ip}

	result.address, error = netlink.ParseAddr(cidr)
	if error != nil {
		error = errors.Wrapf(error, "could not parse address '%s'", _address)

		return
	}

	result.link, error = netlink.LinkByName(_interface)
	if error != nil {
		error = errors.Wrapf(error, "could not get link for interface '%s'", _interface)

		return
	}

	return
}

func (configurator NetlinkNetworkConfigurator) AddIP() error {
	result, error := configurator.IsSet()
	if error != nil {
		return errors.Wrap(error, "ip check in AddIP failed")
	}

	// Already set
	if result {
		return nil
	}

	if error = netlink.AddrAdd(configurator.link, configurator.address); error != nil {
		return errors.Wrap(error, "could not add ip")
	}

	if error = arping.GratuitousArpOverIfaceByName(configurator.ip, configurator._interface); error != nil {
		return errors.Wrap(error, "gratuitous arp failed")
	}

	return nil
}

func (configurator NetlinkNetworkConfigurator) DeleteIP() error {
	result, error := configurator.IsSet()
	if error != nil {
		return errors.Wrap(error, "ip check in DeleteIP failed")
	}

	// Nothing to delete
	if !result {
		return nil
	}

	if error = netlink.AddrDel(configurator.link, configurator.address); error != nil {
		return errors.Wrap(error, "could not delete ip")
	}

	return nil
}

func (configurator NetlinkNetworkConfigurator) IsSet() (result bool, error error) {
	var addresses []netlink.Addr

	addresses, error = netlink.AddrList(configurator.link, 0)
	if error != nil {
		error = errors.Wrap(error, "could not list addresses")

		return
	}

	for _, address := range addresses {
		if address.Equal(*configurator.address) {
			return true, nil
		}
	}

	return false, nil
}

func (configurator NetlinkNetworkConfigurator) IP() string {
	return configurator.address.IP.String()
}

func (configurator NetlinkNetworkConfigurator) Interface() string {
	return configurator.link.Attrs().Name
}
