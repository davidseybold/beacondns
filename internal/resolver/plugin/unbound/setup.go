package unbound

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() {
	caddy.RegisterPlugin("unbound", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	u, err := unboundParse(c)
	if err != nil {
		return plugin.Error("unbound", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		u.Next = next
		return u
	})
	c.OnShutdown(u.Stop)

	return nil
}

//nolint:gocognit // This was copied from the coredns plugin.
func unboundParse(c *caddy.Controller) (*Unbound, error) {
	u := New()

	i := 0
	for c.Next() {
		// Return an error if unbound block specified more than once
		if i > 0 {
			return nil, plugin.ErrOnce
		}
		i++

		u.from = c.RemainingArgs()
		if len(u.from) == 0 {
			u.from = make([]string, len(c.ServerBlockKeys))
			copy(u.from, c.ServerBlockKeys)
		}
		for i, str := range u.from {
			u.from[i] = plugin.Host(str).Normalize()
		}

		for c.NextBlock() {
			var args []string
			var err error

			switch c.Val() {
			case "except":
				except := c.RemainingArgs()
				if len(except) == 0 {
					return nil, c.ArgErr()
				}
				for j := 0; j < len(except); j++ {
					except[j] = plugin.Host(except[j]).Normalize()
				}
				u.except = except
			case "option":
				args = c.RemainingArgs()
				if len(args) != 2 {
					return nil, c.ArgErr()
				}
				if err = u.setOption(args[0], args[1]); err != nil {
					return nil, err
				}
			case "config":
				args = c.RemainingArgs()
				if len(args) != 1 {
					return nil, c.ArgErr()
				}
				if err = u.config(args[0]); err != nil {
					return nil, err
				}
			default:
				return nil, c.ArgErr()
			}
		}
	}
	return u, nil
}
