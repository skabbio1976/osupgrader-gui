//go:build windows

package vcenter

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"time"

	"github.com/alexbrainman/sspi/negotiate"
	"github.com/vmware/govmomi/vim25"
	vimmethods "github.com/vmware/govmomi/vim25/methods"
	vimsoap "github.com/vmware/govmomi/vim25/soap"
	vimtypes "github.com/vmware/govmomi/vim25/types"
	"github.com/skabbio1976/osupgrader-gui/internal/config"
)

// LoginSSPI etablerar en vCenter-session med Windows-integrerad autentisering (Kerberos/SSPI)
func LoginSSPI(cfg *config.VCenterConfig) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config är nil")
	}

	h := normalizeServerHost(cfg.Host)
	if h == "" {
		return nil, fmt.Errorf("tom host – inget vCenter angivet")
	}

	// Cache-check
	clientMu.Lock()
	if cachedVim != nil && cachedHost == h && cachedSess != "" {
		vim := cachedVim
		clientMu.Unlock()
		return &Client{vim: vim}, nil
	}
	clientMu.Unlock()

	// Bygg URL med https:// och /sdk
	raw := h
	if !(len(raw) >= 8 && (raw[:8] == "https://" || raw[:7] == "http://")) {
		raw = "https://" + raw
	}
	// Ta bort eventuell avslutande slash
	if raw[len(raw)-1] == '/' {
		raw = raw[:len(raw)-1]
	}
	if !hasSDKPath(raw) {
		raw = raw + "/sdk"
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("ogiltig URL: %w", err)
	}

	soapClient := vimsoap.NewClient(u, cfg.Insecure)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, fmt.Errorf("vim25.NewClient misslyckades: %w", err)
	}

	// Förbered SSPI (Kerberos) kontext för SPN host/<vcenter>
	cred, err := negotiate.AcquireCurrentUserCredentials()
	if err != nil {
		return nil, fmt.Errorf("AcquireCurrentUserCredentials: %w", err)
	}
	defer cred.Release()

	target := "host/" + h
	secctx, outToken, err := negotiate.NewClientContext(cred, target)
	if err != nil {
		return nil, fmt.Errorf("NewClientContext: %w", err)
	}
	defer secctx.Release()

	var sess vimtypes.UserSession
	for {
		req := vimtypes.LoginBySSPI{
			This:        *c.ServiceContent.SessionManager,
			Locale:      "en_US",
			Base64Token: base64.StdEncoding.EncodeToString(outToken),
		}

		resp, err := vimmethods.LoginBySSPI(ctx, c, &req)
		if err == nil {
			sess = resp.Returnval
			break
		}

		// Hantera SSPIChallenge: uppdatera klientens säkerhetskontext och fortsätt
		if vimsoap.IsSoapFault(err) {
			if vf := vimsoap.ToVimFault(err); vf != nil {
				if ch, ok := vf.(*vimtypes.SSPIChallenge); ok {
					in, _ := base64.StdEncoding.DecodeString(ch.Base64Token)
					done, next, uerr := secctx.Update(in)
					if uerr != nil {
						return nil, fmt.Errorf("SSPI Update: %w", uerr)
					}
					outToken = next
					if done && len(outToken) == 0 {
						outToken = []byte{}
					}
					continue
				}
			}
		}
		return nil, fmt.Errorf("LoginBySSPI: %w", err)
	}

	if sess.Key == "" {
		return nil, fmt.Errorf("SSPI inloggning saknar session")
	}

	// Cache session
	clientMu.Lock()
	cachedVim = c
	cachedHost = h
	cachedUser = sess.UserName
	cachedSess = sess.Key
	clientMu.Unlock()

	return &Client{vim: c}, nil
}
