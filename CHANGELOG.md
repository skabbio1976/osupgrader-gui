# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- English language support for UI (i18n) - In Progress
- Support for Windows Server Standard edition - Planned
- Support for Windows Server 2025 - Planned

### Changed
- All code comments translated from Swedish to English for international contributors

## [0.2.0] - 2025-01-18

### Added
- **Community & Open Source**
  - MIT License for open source distribution
  - Contributing guidelines (CONTRIBUTING.md)
  - Code of Conduct (CODE_OF_CONDUCT.md)
  - Security policy (SECURITY.md)
  - GitHub issue templates (bug report, feature request, question)
  - GitHub pull request template
  - Comprehensive .gitignore

- **Core Features**
  - Graphical interface with Fyne framework
  - Automatic DPI scaling for optimal display
  - vCenter login with password and Windows SSPI/Kerberos support
  - Unsigned certificate support
  - VM selection with table view (Name, Folder, Domain, OS columns)
  - Search filtering and multi-select
  - Multi-domain support with automatic domain append
  - Snapshot management (create, list, delete)
  - Parallel upgrades with configurable concurrency
  - Real-time progress tracking and logging
  - ISO validation before upgrade
  - Configuration management via GUI dialog
  - Debug logging with `-d/--debug` flag

- **Upgrade Process**
  - Credential validation before upgrade (prevents account lockout)
  - Disk space precheck (configurable minimum)
  - Optional snapshot creation before upgrade
  - Automatic detection of Server Core vs Desktop Experience
  - ISO mounting and unmounting
  - PowerShell-based upgrade execution
  - Post-reboot signaling via scheduled tasks
  - OS version verification
  - Automatic cleanup of temporary files

- **PowerShell Scripts**
  - `uppgradeos.ps1` - Main upgrade script with auto-detection
  - `createsignaltasks.ps1` - Post-reboot signal task setup
  - `cleanup.ps1` - Cleanup of signal files and temporary scripts

### Changed
- All code comments translated to English
- Improved error messages and logging
- Enhanced debug logging with safe credential handling

### Security
- Passwords never stored in configuration files
- Safe debug logging (passwords never in clear text)
- Credential validation before upgrade attempts
- Thread-safe operations with mutex protection
- VMware Tools crash recovery
- Timeout handling for all operations

### Documentation
- Comprehensive README in Swedish with English sections
- Detailed troubleshooting guide
- Security features documentation
- Multi-domain setup instructions
- SSPI/Kerberos authentication guide

## [0.1.0] - 2024-12-XX (Internal Release)

### Added
- Initial internal version
- Basic upgrade functionality
- Swedish-only interface
- Manual testing phase

---

## Version History Notes

### Semantic Versioning Guide

Given a version number MAJOR.MINOR.PATCH:
- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality (backwards-compatible)
- **PATCH**: Bug fixes (backwards-compatible)

### Change Categories

- **Added**: New features
- **Changed**: Changes to existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Security vulnerability fixes

---

## How to Update CHANGELOG

When making changes:

1. Add changes under `[Unreleased]` section
2. Use appropriate category (Added, Changed, Fixed, etc.)
3. Write clear, concise descriptions
4. Include issue/PR numbers when applicable
5. Move changes to a new version section on release

### Example Entry Format

```markdown
### Added
- New feature description (#123)
- Another feature with more details
  - Sub-point with implementation details
  - Technical note about breaking changes

### Fixed
- Bug description and impact (#456)
- Another bug fix

### Security
- Security vulnerability description (CVE-XXXX-YYYY)
```

---

## Migration Guides

### Upgrading to 0.2.0 from 0.1.0

No breaking changes. Configuration file format remains compatible.

**Recommended actions:**
- Review new security documentation
- Update vCenter service account permissions if using SSPI
- Enable debug logging for first production run

---

## Future Roadmap

See README.md for detailed roadmap. Highlights:

- [ ] English UI (v0.3.0)
- [ ] Standard edition support (v0.3.0)
- [ ] Windows Server 2025 support (v0.4.0)
- [ ] Automated rollback (v0.5.0)
- [ ] REST API (v1.0.0)

---

**For the latest changes, see:** [GitHub Releases](https://github.com/YOURORG/osupgrader-gui/releases)
