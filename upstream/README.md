# Upstream SSSD Man Pages & API Drop-Zone

This directory is an input drop-zone for importing SSSD configuration definitions from upstream SSSD releases. It is git-ignored (only this `README.md` is tracked in git).

The generated, versioned catalog is compiled into the binary at `sssd_catalog/catalog.json`, with a human-readable summary at `sssd_catalog/catalog.md`.

## How to Import / Update

Copy the upstream files into this directory. Any combination of the following is supported:

### 1. From an SSSD git checkout (recommended)
```bash
# From the SSSD source checkout root (e.g. ~/Desktop/sssd-master):
cp src/config/etc/sssd.api.conf upstream/
cp -r src/config/etc/sssd.api.d upstream/
cp src/man/*.xml upstream/
cp -r src/man/include upstream/
cp version.m4 upstream/ 2>/dev/null || true
```

### 2. From installed man pages on an openSUSE/SLES or Fedora system
```bash
cp /usr/share/man/man5/sssd*.5* upstream/
```

### 3. Generate the catalog
Run:
```bash
sssd-inspector -gen-catalog upstream
```
or if building with `go run`:
```bash
go run . -gen-catalog upstream
```

This will parse the files in `upstream/`, cross-check option names and types, and update:
- `sssd_catalog/catalog.json` (machine-readable, embedded in the inspector binary)
- `sssd_catalog/catalog.md` (clean, human-readable reference documentation grouped by section)

Then recompile the binary or run tests:
```bash
go test ./...
```
