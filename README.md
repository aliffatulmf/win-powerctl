# win-powerctl

HTTP service for remote system power control.

## Features

- Shutdown, restart, power off via HTTP API
- Windows service support
- Health check with DLL status
- Structured logging with zerolog
- Password authentication

## Build

### PowerShell (recommended)

```powershell
.\build.ps1
```

With custom MSVC path:

```powershell
.\build.ps1 -MSVCPath "C:\path\to\vcvarsall.bat"
```

Output will be in `dist/`:

```
dist/
├── config.ini
├── poweroff.dll
└── win-powerctl.exe
```

### Go only (without DLL)

```bash
go build -o dist/win-powerctl.exe ./cmd/win-powerctl
```

## Usage

1. Edit `config.ini`:

```ini
[server]
host = 0.0.0.0
port = 10125

[auth]
password = changeme
```

2. Run:

```bash
win-powerctl.exe
```

## Commands

| Command     | Description                       |
| ----------- | --------------------------------- |
| _(no args)_ | Run HTTP server                   |
| `install`   | Install as Windows service        |
| `uninstall` | Remove from Windows services      |
| `start`     | Start the service                 |
| `stop`      | Stop the service                  |
| `restart`   | Restart the service               |
| `service`   | Run as Windows service (internal) |
| `version`   | Print version                     |

## API

| Endpoint                        | Description         |
| ------------------------------- | ------------------- |
| `GET /shutdown?auth=<password>` | Graceful shutdown   |
| `GET /health`                   | Health check (JSON) |

### Health Response

```json
{
  "status": "ok",
  "server": { "host": "0.0.0.0", "port": 10125 },
  "poweroff": { "loaded": true }
}
```

## Logs

Logs are written to `logs/<date>.log` in plain text format.
