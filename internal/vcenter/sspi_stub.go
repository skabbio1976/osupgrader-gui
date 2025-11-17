//go:build !windows

package vcenter

import (
	"errors"

	"github.com/yourusername/osupgrader-gui/internal/config"
)

var ErrSSPINotSupported = errors.New("SSPI is only supported on Windows")

// LoginSSPI is not available on non-Windows platforms
func LoginSSPI(cfg *config.VCenterConfig) (*Client, error) {
	return nil, ErrSSPINotSupported
}
