package middlewares

import (
	"log"
	"net"
	"net/http"
	"strings"
)

type WebhooksMiddleware struct {
	allowedIPNets []*net.IPNet
}

func NewWebhooksMiddleware() (*WebhooksMiddleware, error) {
	// These ip are specified in the Trello documentation here
	// https://developer.atlassian.com/cloud/trello/guides/rest-api/webhooks/
	trelloCIDRValues := []string{
		"107.23.104.115/32",
		"107.23.149.70/32",
		"54.152.166.250/32",
		"54.164.77.56/32",
		"54.209.149.230/32",
		"18.234.32.224/28",
	}

	wm := &WebhooksMiddleware{
		allowedIPNets: make([]*net.IPNet, 0),
	}

	for _, cidrValue := range trelloCIDRValues {
		_, ipNet, err := net.ParseCIDR(cidrValue)
		if err != nil {
			return nil, err
		}

		wm.allowedIPNets = append(wm.allowedIPNets, ipNet)
	}

	return wm, nil
}

func (wm WebhooksMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteIP := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])

		validRemoteAddr := false

		for _, allowedIPNet := range wm.allowedIPNets {
			if allowedIPNet.Contains(remoteIP) {
				validRemoteAddr = true
				break
			}
		}

		if validRemoteAddr {
			next.ServeHTTP(w, r)
		} else {
			log.Printf("Ip address (%v) not allowed", remoteIP)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}
