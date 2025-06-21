# Security Policy

## Supported Versions

Use this section to tell people about which versions of your project are currently being supported with security updates.

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability within Go Active Record, please send an email to security@forester.co. All security vulnerabilities will be promptly addressed.

Please do not create a public GitHub issue for security vulnerabilities.

### What to include in your report

- A description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Suggested fix (if any)

### Response timeline

- **Initial response**: Within 48 hours
- **Status update**: Within 1 week
- **Fix release**: Within 30 days (depending on severity)

### Responsible disclosure

We follow responsible disclosure practices:

1. **Private reporting**: Report vulnerabilities privately
2. **Timeline**: We'll work with you to establish a timeline for disclosure
3. **Credit**: We'll credit you in the security advisory
4. **Coordination**: We'll coordinate the public disclosure

## Security Best Practices

When using Go Active Record:

1. **Keep dependencies updated**: Regularly update your dependencies
2. **Use HTTPS**: Always use HTTPS in production
3. **Validate input**: Always validate and sanitize user input
4. **Use prepared statements**: The library uses prepared statements by default
5. **Follow principle of least privilege**: Use database users with minimal required permissions

## Security Features

Go Active Record includes several security features:

- **SQL injection protection**: Uses prepared statements
- **Input validation**: Built-in validation framework
- **Type safety**: Strong typing prevents many security issues
- **Connection security**: Supports SSL/TLS connections

## Reporting Security Issues

**Do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to security@forester.co.

You should receive a response within 48 hours. If for some reason you do not, please follow up via email to ensure we received your original message.

Please include the requested information listed above (as much as you can provide) to help us better understand the nature and scope of the possible issue. 