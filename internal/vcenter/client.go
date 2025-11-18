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

// Client represents a vCenter client
type Client struct {
	vim *vim25.Client
}

// Login performs actual login to vCenter via govmomi
func Login(cfg *config.VCenterConfig, password string) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if cfg.Host == "" || cfg.Username == "" {
		return nil, fmt.Errorf("host or username missing")
	}

	// Cache check
	clientMu.Lock()
	if cachedVim != nil && cachedHost == cfg.Host && cachedUser == cfg.Username && cachedSess != "" {
		vim := cachedVim
		clientMu.Unlock()
		return &Client{vim: vim}, nil
	}
	clientMu.Unlock()

	// Build URL (always add https:// and /sdk if not present)
	raw := cfg.Host
	if !(len(raw) >= 8 && (raw[:8] == "https://" || raw[:7] == "http://")) {
		raw = "https://" + raw
	}
	// Remove trailing slash if present
	if raw[len(raw)-1] == '/' {
		raw = raw[:len(raw)-1]
	}
	if !hasSDKPath(raw) {
		raw = raw + "/sdk"
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	u.User = url.UserPassword(cfg.Username, password)

	soapClient := soap.NewClient(u, cfg.Insecure)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, fmt.Errorf("could not create vim25 client: %w", err)
	}
	sm := session.NewManager(vimClient)
	if err := sm.Login(ctx, u.User); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}
	us, err := sm.UserSession(ctx)
	if err != nil || us == nil {
		return nil, fmt.Errorf("could not fetch user session: %w", err)
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

// GetVim returns the vim25 client
func (c *Client) GetVim() *vim25.Client {
	return c.vim
}

// GetCachedClient returns the cached client
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
