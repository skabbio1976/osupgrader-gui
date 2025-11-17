package vcenter

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/yourusername/osupgrader-gui/internal/config"
)

var (
	clientMu   sync.Mutex
	cachedHost string
	cachedUser string
	cachedSess string
	cachedVim  *vim25.Client
)

// Client representerar en vCenter-klient
type Client struct {
	vim *vim25.Client
}

// Login gör faktisk inloggning mot vCenter via govmomi
func Login(cfg *config.VCenterConfig, password string) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config är nil")
	}
	if cfg.Host == "" || cfg.Username == "" {
		return nil, fmt.Errorf("host eller username saknas")
	}

	// Cache-check
	clientMu.Lock()
	if cachedVim != nil && cachedHost == cfg.Host && cachedUser == cfg.Username && cachedSess != "" {
		vim := cachedVim
		clientMu.Unlock()
		return &Client{vim: vim}, nil
	}
	clientMu.Unlock()

	// Bygg URL (lägg alltid till https:// och /sdk om det inte finns)
	raw := cfg.Host
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
	u.User = url.UserPassword(cfg.Username, password)

	soapClient := soap.NewClient(u, cfg.Insecure)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, fmt.Errorf("kunde inte skapa vim25 klient: %w", err)
	}
	sm := session.NewManager(vimClient)
	if err := sm.Login(ctx, u.User); err != nil {
		return nil, fmt.Errorf("inloggning misslyckades: %w", err)
	}
	us, err := sm.UserSession(ctx)
	if err != nil || us == nil {
		return nil, fmt.Errorf("kunde inte hämta user session: %w", err)
	}

	// Cache session
	sessID := randomID()
	clientMu.Lock()
	cachedHost = cfg.Host
	cachedUser = cfg.Username
	cachedSess = sessID
	cachedVim = vimClient
	clientMu.Unlock()

	return &Client{vim: vimClient}, nil
}

// GetVim returnerar vim25-klienten
func (c *Client) GetVim() *vim25.Client {
	return c.vim
}

// GetCachedClient returnerar den cachade klienten
func GetCachedClient() *vim25.Client {
	clientMu.Lock()
	defer clientMu.Unlock()
	return cachedVim
}

func hasSDKPath(u string) bool {
	return len(u) >= 4 && (u[len(u)-4:] == "/sdk" || u[len(u)-4:] == "sdk/")
}

func randomID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
