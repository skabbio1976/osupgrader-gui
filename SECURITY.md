# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.2.x   | :white_check_mark: |
| < 0.2   | :x:                |

## Reporting a Vulnerability

**Please DO NOT report security vulnerabilities through public GitHub issues.**

If you discover a security vulnerability in OSUpgrader, please follow these steps:

### 1. Private Disclosure

Send an email to the project maintainers with:

- A description of the vulnerability
- Steps to reproduce the issue
- Potential impact of the vulnerability
- Any suggested fixes (if you have them)

**Please allow up to 48 hours for an initial response.**

### 2. Do Not Disclose Publicly

Please do not publicly disclose the vulnerability until:
- We have confirmed the vulnerability
- We have developed and tested a fix
- We have released a patched version
- Sufficient time has passed for users to upgrade (typically 7-14 days)

### 3. Coordinated Disclosure

We will work with you to:
- Understand the full scope of the vulnerability
- Develop an appropriate fix
- Test the fix thoroughly
- Prepare a security advisory
- Credit you for the discovery (if you wish)

## Security Considerations for Users

### Credentials

- **vCenter Password**: Never stored in configuration files (only in memory during session)
- **Guest Password**: Never stored in configuration files (only in memory during execution)
- **Debug Logs**: Passwords are never logged in clear text (only password length is logged)

### Network Security

- **SSL/TLS**: By default, certificate verification is enabled. Only disable with `insecure: true` if absolutely necessary
- **SSPI/Kerberos**: When using Windows authentication, credentials are handled by the OS security subsystem

### VMware Environment

- **Least Privilege**: Use a service account with minimal required permissions:
  - VM power operations
  - VM snapshot operations
  - VM reconfiguration (for ISO mounting)
  - Guest operations (for running scripts)

- **Audit Logging**: All upgrade operations are logged for audit purposes

### Configuration Files

- **File Permissions**: Config files should be readable only by the user running OSUpgrader
  - Linux/macOS: `chmod 600 ~/conf.json`
  - Windows: Restrict NTFS permissions appropriately

### Guest Scripts

- **Code Review**: All PowerShell scripts are embedded in the binary and can be reviewed in the source code
- **Execution Policy**: Scripts run with `-ExecutionPolicy Bypass` to avoid policy conflicts
- **Cleanup**: Signal files and temporary scripts are cleaned up after upgrade

## Known Limitations

### Credential Validation

- Multiple failed credential attempts can lead to account lockout
- The tool validates credentials before upgrade to minimize this risk
- Ensure you test credentials in a non-production environment first

### Snapshot Security

- Snapshots may contain sensitive data in VM memory (if memory snapshots are enabled)
- Use `skip_memory_in_snapshot: true` if this is a concern
- Snapshots consume significant storage - monitor datastore space

### Domain Credentials

- Auto-domain-append feature requires accurate DNS/domain detection
- Verify domain membership before running large-scale upgrades
- Use explicit `DOMAIN\user` or `user@domain.com` format if auto-append fails

## Security Best Practices

1. **Test First**: Always test in a non-production environment
2. **Backup**: Verify backups exist before running production upgrades
3. **Snapshots**: Use pre-upgrade snapshots (enabled by default)
4. **Monitoring**: Monitor upgrade progress and check logs
5. **Rollback Plan**: Have a tested rollback procedure
6. **Access Control**: Limit who can run OSUpgrader in your organization
7. **Network Isolation**: Run from a secure management network
8. **Audit Trail**: Review logs after upgrades complete

## Compliance Considerations

### GDPR / Data Protection

- OSUpgrader processes:
  - Server names
  - Administrator credentials (temporarily, in memory only)
  - Domain information
  - Network information

- No personal data is transmitted outside your environment
- All data remains within your vCenter infrastructure

### Logging

- Debug logs may contain:
  - Server names
  - Usernames (but never passwords)
  - Network paths
  - Error messages

- Review debug logs before sharing externally
- Redact sensitive information if sharing logs for troubleshooting

## Updates and Patches

Subscribe to:
- GitHub release notifications
- Security advisories (when available)

To update:
```bash
# Download latest release
# Backup your config
cp ~/conf.json ~/conf.json.backup

# Replace binary
# Verify version
./osupgrader-gui --version
```

## Questions?

If you have security questions that don't involve reporting a vulnerability, please:
- Open a GitHub issue with the `security` label
- Start a GitHub Discussion in the Security category

---

**Security is everyone's responsibility. Thank you for helping keep OSUpgrader and its users safe!**
