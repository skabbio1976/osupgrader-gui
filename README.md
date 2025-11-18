# OSUpgrader GUI

An automated tool for upgrading Windows Server 2016/2019 to Windows Server 2022/2025 Datacenter via VMware vCenter/vSphere.

**[🇸🇪 Swedish version / Svensk version](README-SWE.md)**

## Overview

OSUpgrader GUI is a Fyne-based graphical application that makes it easy to upgrade multiple Windows Server VMs simultaneously through an intuitive user interface.

## Features

- **Graphical user interface** with Fyne framework
  - Automatic DPI scaling for optimal display on all screens
- **vCenter login** with support for:
  - Password authentication (all platforms)
  - Windows SSPI/Kerberos single sign-on (Windows only) ✓ Tested and verified
  - Self-signed certificates
- **VM selection** with table view (Name, Folder, Domain, OS columns), search filtering and multi-select
- **Multi-domain support**:
  - Automatic domain append to username (e.g. `upgrade` → `upgrade@domain.local`)
  - Enables same service account across multiple domains
  - Shows VM's domain in table view
- **Snapshot management**:
  - Automatic snapshot before upgrade
  - Separate screen for managing and removing pre-upgrade snapshots
  - Batch removal of multiple snapshots simultaneously
  - Filtering by snapshot prefix
- **Parallel upgrades** with configurable concurrency
- **Progress tracking** with real-time logging and readable text
- **ISO validation** before upgrade starts
- **Configuration management** via GUI dialog with saved guest credentials
- **Debug logging** (optional with `-d/--debug` flag):
  - Detailed logging to `debuglogg.txt`
  - Safe logging (passwords never in cleartext)
  - Perfect for troubleshooting in airgapped environments
- **Secure authentication**:
  - Credential validation before upgrade
  - Prevents account lockout from failed attempts

## System Requirements

- Go 1.24 or later
- Linux/Windows/macOS
- Access to VMware vCenter
- Windows Server 2022/2025 Datacenter ISO on datastore

## Installation

### Build from source

```bash
# Clone the project
git clone https://github.com/yourusername/osupgrader-gui.git
cd osupgrader-gui

# Fetch dependencies
go mod tidy

# Build the application
go build -o osupgrader-gui ./cmd/osupgrader-gui

# Run the application
./osupgrader-gui
```

### Build for both Linux and Windows

```bash
# Use the build script (requires Docker for Windows builds)
./build.sh
```

### Manual Windows compilation

```bash
# Install fyne-cross first
go install github.com/fyne-io/fyne-cross@latest

# Build for Windows from Linux (requires Docker)
~/go/bin/fyne-cross windows -arch=amd64 -app-id com.example.osupgrader ./cmd/osupgrader-gui

# Extract
cd fyne-cross/dist/windows-amd64
unzip osupgrader-gui.exe.zip
```

## Usage

1. **Start the application**
   ```bash
   # Normal usage (no debug logging)
   ./osupgrader-gui

   # With debug logging (for troubleshooting)
   ./osupgrader-gui -d
   # or
   ./osupgrader-gui --debug
   ```

   When debug logging is enabled, `debuglogg.txt` is created in the same folder as the program with detailed information about all operations.

2. **Log in to vCenter**
   - Enter vCenter host (e.g. `vcenter.example.local`)
   - Choose authentication method:
     - **Password**: Enter username and password
     - **Windows SSPI/Kerberos**: Single sign-on with your Windows domain account (Windows only)
       - No password input required
       - Uses your Windows credentials automatically
       - Perfect for domain environments with integrated authentication
   - Check "Allow self-signed certificates" if necessary
   - Click "Log in"

3. **Select VMs to upgrade**
   - Table view shows all VMs with columns: Select, Name, Folder, Domain, OS
   - Search for VMs using the search field (searches all columns including domain)
   - Select VMs by checking the checkboxes in the first column
   - Use "Select all" / "Deselect all" for bulk operations
   - Click "Manage snapshots" to remove old pre-upgrade snapshots
   - Click "Continue to upgrade"

4. **Configure upgrade**
   - Enter guest admin user (e.g. `upgrade`)
     - **Multi-domain support**: Enter username only without domain (e.g. `upgrade`)
     - System automatically appends VM's domain: `upgrade@domain1.local`
     - Works perfectly with same service account across multiple domains
     - If you want to specify a specific domain, use `DOMAIN\user` or `user@domain.com`
   - Enter guest password
   - **💡 Tip**: Save guest credentials in Settings to avoid entering them every time!
   - Enter ISO datastore path (e.g. `[datastore1] iso/windows-server-2022.iso`)
   - Choose if snapshot should be created before upgrade
   - Click "Start upgrade"

5. **Monitor progress**
   - Progress bar shows progress
   - Real-time log shows detailed information (text is readable and can be selected/copied)
   - Status messages update continuously
   - Upgrade runs in background on guest OS

6. **Manage snapshots after upgrade** (recommended workflow)
   - After upgrade: Let application owners verify the system works
   - Go to "Manage snapshots"
   - Select pre-upgrade snapshots to remove
   - Confirm removal (cannot be undone!)
   - Free up disk space on datastore

## Configuration

Configuration is saved in `~/conf.json` and can be edited via the GUI's settings dialog:

```json
{
  "vcenter": {
    "vcenter_url": "vcenter.example.local",
    "username": "administrator@vsphere.local",
    "mode": "password",
    "insecure": true
  },
  "defaults": {
    "snapshot_name_prefix": "pre-upgrade",
    "iso_datastore_path": "[datastore1] iso/windows-server-2022.iso",
    "skip_memory_in_snapshot": true,
    "guest_username": "upgrade"
  },
  "upgrade": {
    "parallel": 2,
    "reboot": true,
    "timeout_minutes": 90,
    "precheck_disk_gb": 10
  },
  "timeouts": {
    "signal_script_seconds": 30,
    "signal_files_minutes": 30,
    "target_os_minutes": 20,
    "poweroff_minutes": 5
  },
  "logging": {
    "level": "info",
    "file": "osupgrader.log"
  },
  "ui": {
    "language": "sv",
    "dark_mode": false
  }
}
```

### Configuration Options

#### vCenter Settings
- **vcenter_url**: vCenter server hostname
- **username**: vCenter username
- **insecure**: Allow self-signed SSL certificates

#### Guest OS Credentials
- **guest_username**: Windows admin user on VMs (e.g. `upgrade`)
  - Saved in configuration file for convenience
  - If username lacks domain (`\` or `@`), VM's domain is added automatically
  - Example: `upgrade` → `upgrade@domain1.local` (based on VM's domain)
- **guest_password**: Windows admin password (masked in GUI with show/hide button)
  - ⚠️ **Security note**: Password is only stored in memory during application runtime and is NEVER written to the configuration file

#### Upgrade Settings
- **snapshot_name_prefix**: Prefix for snapshot names
- **iso_datastore_path**: Path to Windows Server 2022/2025 ISO
- **skip_memory_in_snapshot**: Skip memory in snapshot (faster)
- **parallel**: Number of parallel upgrades (1-10)
- **reboot**: Automatically reboot after upgrade
- **timeout_minutes**: Timeout for upgrade per VM
- **precheck_disk_gb**: Minimum free disk space (GB)

#### Timeout Settings
- **signal_script_seconds**: Wait time for signal task script completion
- **signal_files_minutes**: Wait time for scheduled task signal files to appear
- **target_os_minutes**: Max time to wait for target OS version
- **poweroff_minutes**: Max time to wait for guest shutdown before forced power off

## Upgrade Process

1. **Validation**
   - Validate guest credentials (prevents account lockout)
   - Check that ISO file exists on datastore
   - Check disk space on guest OS (at least 10 GB free)
   - Check that VM is powered on and VMware Tools is running

2. **Snapshot**
   - Create snapshot for recovery (optional)
   - Verify that snapshot was created correctly
   - Name format: `pre-upgrade-pre-YYYYMMDD-HHMM`

3. **ISO Mounting**
   - Mount Windows Server 2022/2025 ISO to CD-ROM
   - Verify that ISO is mounted

4. **Upgrade**
   - Run PowerShell upgrade script via VMware Tools
   - Script automatically detects OS edition (Datacenter/Standard, Core/Desktop)
   - Script sets appropriate GVLK key based on detected SKU
   - Selects correct WIM image index (1=Standard Core, 2=Standard Desktop, 3=Datacenter Core, 4=Datacenter Desktop)
   - Starts Windows Setup with `/auto upgrade /noreboot`
   - Waits for setup.exe to complete (with `-Wait`)
   - Schedules a graceful shutdown in Windows (60 seconds to clean up services)

5. **Monitoring**
   - Polling of PowerShell script exit and checking exit code
   - Waits for VM to go to `poweredOff`, forces `PowerOff` via vCenter if not within `poweroff_minutes`
   - Sleeps 60 seconds and powers on VM via vCenter before next phase
   - Polling VMware Tools/OS version every 45 seconds until Windows Server 2022/2025 is reported
   - Timeout after configured time (default: 90 minutes + configurable power-off timeout)

6. **Completion**
   - Waits for scheduled task signal files (task-based indicator) to see login environment is ready
   - Unmount ISO when upgrade is complete
   - Verify OS version is 2022 or 2025

## Project Structure

```
osupgrader-gui/
├── cmd/
│   └── osupgrader-gui/
│       └── main.go              # Main program (with -d/--debug flag)
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── debug/
│   │   └── logger.go            # Debug logging to file
│   ├── vcenter/
│   │   ├── client.go            # vCenter client and login
│   │   ├── inventory.go         # VM inventory management (with domain)
│   │   ├── snapshot.go          # Snapshot operations
│   │   └── types.go             # Data types (VMInfo with Domain)
│   ├── upgrade/
│   │   ├── upgrade.go           # Upgrade logic (auto-domain append)
│   │   ├── validators.go        # Validation functions
│   │   ├── iso.go               # ISO management
│   │   └── assets/
│   │       ├── upgradeos.ps1    # Upgrade PowerShell script
│   │       ├── cleanup.ps1      # Cleanup script
│   │       ├── createsignaltasks.ps1  # Signal task creation
│   │       └── processmonitor.ps1     # Process monitoring
│   └── gui/
│       ├── app.go               # Main application (DPI scaling)
│       ├── login.go             # Login screen
│       ├── vmselection.go       # VM selection screen (with Domain column)
│       ├── upgrade.go           # Upgrade workflow screen
│       ├── snapshots.go         # Snapshot management screen
│       └── settings.go          # Settings dialog
├── go.mod
├── go.sum
└── README.md
```

## Security Features

- **Passwords never stored** in configuration file
- **Safe debug logging**: Passwords never logged in cleartext (only length)
- **Credential validation**: Checks credentials before upgrade to prevent account lockout
- **Windows SSPI/Kerberos support** for secure single sign-on without password input
- **Snapshot verification** prevents data loss
- **Snapshot management with confirmation**: Confirmation dialog for snapshot removal
- **ISO validation** before snapshot saves time
- **Thread-safe** operations with mutex protection
- **VMware Tools crash recovery** handles restarts during upgrade
- **Timeout handling** prevents hanging upgrades
- **Multi-domain support**: Automatic domain append reduces risk of incorrect username

## Windows SSPI/Kerberos Authentication

SSPI (Security Support Provider Interface) is Microsoft's API for authentication and security in Windows. When using SSPI login:

1. **Transparent authentication**: Application uses your Windows credentials automatically
2. **No password input**: You don't need to enter password - perfect for smartcard/token users
3. **Domain integration**: Works seamlessly in Active Directory environments
4. **Kerberos protocol**: Secure ticket-based authentication to vCenter
5. **SPN-based**: Uses Service Principal Name `host/vcenter.domain.local` for authentication

**Technical details:**
- Implemented via `github.com/alexbrainman/sspi/negotiate`
- Supports multi-round SSPI handshake with `SSPIChallenge`
- Compatible with both PowerCLI and standard govmomi sessions
- Only available on Windows platform (stub on Linux/macOS)

## Troubleshooting

### Debug Logging
For detailed troubleshooting, start the application with the debug flag:
```bash
./osupgrader-gui -d
```

This creates `debuglogg.txt` in the same folder as the program with:
- All API calls to vCenter
- Guest operations details
- Authentication attempts (username and password length, but NOT the password)
- PowerShell script execution
- Snapshot operations
- ISO mounting/unmounting
- All errors with stack traces

**Important information in debug log:**
- Timestamps for all operations
- VM name and domain information
- Exit codes from PowerShell scripts
- OS version before and after upgrade

### Login Failed
- **Password authentication**:
  - Check vCenter URL and credentials
  - Enable "Allow self-signed certificates" if using self-signed cert
  - Check network access to vCenter
- **SSPI/Kerberos authentication**:
  - Only works on Windows
  - Requires you to be logged in with a domain account
  - vCenter server must be Windows-integrated (Active Directory)
  - Check that Kerberos SPN is correctly configured (`host/vcenter.domain.local`)
  - On Linux, only password authentication is used

### ISO Validation Failed
- Check that ISO path is correct: `[datastore1] iso/file.iso`
- Verify that datastore exists and is accessible
- Check that ISO file exists on datastore
- Use debug logging to see exactly which datastore is being searched

### Guest OS Authentication Failed
- **Account lockout issues**:
  - Application validates credentials BEFORE upgrade to prevent lockout
  - If credentials are incorrect, you get an error immediately without repeated attempts
- **Multi-domain usage**:
  - Enter username only without domain (e.g. `upgrade`)
  - System automatically appends VM's domain
  - Check that VM's domain is correct in table view
  - If auto-append doesn't work, use `DOMAIN\user` or `user@domain.com`
- **Debug tips**:
  - Run with `-d` flag
  - Check `debuglogg.txt` to see which username is actually used
  - Example: `Auto-appended domain to username: upgrade@domain1.local`

### Upgrade Fails
- Check that VMware Tools is installed and running
- Verify guest credentials are correct
- Check disk space on guest OS (at least 10 GB)
- See log file `C:\Windows\Temp\upgrade.log` on guest OS
- **PowerShell script issues**:
  - Check `C:\Windows\Temp\setup_stdout.log` and `setup_stderr.log`
  - Verify that setup.exe ran (check PID in debug log)
  - Check exit code from PowerShell script (should be 0)
- **Timeout issues**:
  - Default timeout is 90 minutes
  - Increase timeout in settings if upgrade takes longer
  - Slow VMs may need 120-180 minutes

### Snapshot Management
- **Cannot find snapshots**:
  - Check that `snapshot_name_prefix` in config matches snapshot names
  - Default prefix is `pre-upgrade`
  - Snapshot name format: `pre-upgrade-pre-YYYYMMDD-HHMM`
- **Removal failed**:
  - Check that no other operations are running on the VM
  - Verify vCenter permissions for snapshot removal
  - Some snapshots may be locked by backup jobs

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a detailed history of changes.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

We welcome contributions from the community! Whether you're fixing bugs, adding features, or improving documentation, your help makes OSUpgrader better for everyone.

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on:
- How to report bugs
- How to suggest enhancements
- Development workflow
- Code style guidelines
- Pull request process

## Community & Support

- **Bug Reports & Feature Requests**: Open an issue on GitHub using our [issue templates](.github/ISSUE_TEMPLATE/)
- **Questions**: Use GitHub Discussions or open an issue with the `question` label
- **Security Issues**: Please report privately to maintainers - see [SECURITY.md](SECURITY.md)
- **Contributing**: Read our [Contributing Guidelines](CONTRIBUTING.md)
- **Code of Conduct**: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

## Acknowledgments

Special thanks to:
- All the IT operations teams who provided feedback and use cases
- Contributors who help improve this tool
- The teams facing tight upgrade deadlines that inspired this project

## Roadmap

Planned features and improvements:
- [ ] English language UI (i18n support)
- [ ] Support for Windows Server Standard edition
- [x] Support for Windows Server 2025
- [ ] Automated rollback on failure
- [ ] Enhanced pre-flight checks
- [ ] Upgrade history database
- [ ] REST API for automation

Want to contribute to any of these? Check out [CONTRIBUTING.md](CONTRIBUTING.md)!

---

**Made with ❤️ for IT Operations teams everywhere**

If this tool saved you time or helped your organization, consider:
- ⭐ Starring the repository
- 🐛 Reporting bugs you encounter
- 💡 Suggesting improvements
- 🤝 Contributing code or documentation
