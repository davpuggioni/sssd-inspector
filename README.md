# README

# SSSD Inspector 🔍

**SSSD Inspector** is a native diagnostic utility designed to parse a supportconfig file generated on SLES or OpenSUSE and identify complex identity management failures. Unlike generic log viewers, this tool uses a specialized pattern-matching engine derived directly from the SSSD C source code to provide human-readable explanations for cryptic error messages.

---

## 🚀 Key Features

SSSD Inspector categorizes and explains failures across the entire SSSD stack, using insights extracted from the core service modules:

### 1. Frontend Responders
Identifies why the Linux OS is rejecting requests before they ever hit the network:
* **PAM & NSS:** Diagnoses offline authentication blocks, shell vetoes, and negative cache (ncache) rejections.
* **Sudo & SSH:** Troubleshoots missing sudo rules and SSH public key or certificate parsing errors.
* **InfoPipe (IFP) & PAC:** Detects D-Bus permission denials, attribute whitelist filtering, and Kerberos PAC validation failures.

### 2. Backend Providers
Decodes the "wire" protocols and communication logic used to talk to identity servers:
* **Active Directory & LDAP:** Explains USN rollbacks, machine password expiration, and LDAP pagination/size limit errors.
* **Kerberos:** Identifies KDC locator failures, clock skew, and FAST tunnel requirement errors.
* **FreeIPA/IdM:** Advanced diagnostics for Host-Based Access Control (HBAC), SELinux user mapping, and Cross-Forest AD trusts.
* **OAuth2/OIDC:** Maps modern Identity Provider failures like missing device endpoints or JSON parsing errors.

### 3. Internal Logic & Storage
Detects issues with SSSD's "brain" and local persistence:
* **Monitor:** Diagnoses service flapping, watchdog heartbeats, and process supervisor crashes.
* **SysDB & ConfDB:** Identifies LDB database corruption, schema upgrade failures during package updates, and `sssd.conf` syntax errors.
* **Cache Request Plugin:** Pinpoints frontend routing issues, such as UPN-to-domain mapping failures.

---

## 🛠️ Installation & Build

This project is built using **Wails (Go + Vite/TS)** for a lightweight, native desktop experience.

### Prerequisites
* **Go** 1.21+
* **Node.js** & NPM
* **Wails CLI** (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### Build the Production Binary
To compile the standalone binary for Linux:
```bash
wails build -platform linux/amd64 -ldflags "-w -s" -clean
```

---

## ⚖️ License
This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License version 3 as published by the Free Software Foundation.

## ⚠️ Disclaimer
This tool is intended for diagnostic purposes. It parses logs based on patterns found in the SSSD source code; however, always verify system configurations manually before applying changes in production environments.








