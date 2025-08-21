# 🖥️ How Linux Services Work (Systemd Basics)

A **service** in Linux is just a program (script, binary, etc.) that is managed by the **init system** so it can:

* Start automatically on boot
* Be restarted if it crashes
* Be controlled with standard commands (`start`, `stop`, `status`)
* Run in the background without you needing to keep a terminal open

On modern Linux (Ubuntu, Fedora, Debian, CentOS, Arch etc.), the init system is **systemd**.

---

## 1️⃣ What is systemd?

* **systemd** is a daemon (runs in background) that starts right after the kernel boots.
* It manages **units** — things like services, sockets, mounts, timers, devices.
* A **service unit** is defined in a `.service` file, which tells systemd *how to run your program*.

---

## 2️⃣ Service File Structure

Service files live in `/etc/systemd/system/` (for custom services).
The format is **INI-style** sections:

### Example

```ini
[Unit]
Description=My Custom Service
After=network.target

[Service]
ExecStart=/usr/bin/python3 /opt/myapp/app.py
WorkingDirectory=/opt/myapp
Restart=always
User=myuser
Group=myuser

[Install]
WantedBy=multi-user.target
```

---

### Sections explained:

#### 🔹 `[Unit]`

* Metadata about the service
* `Description=` → human-readable description
* `After=network.target` → wait for network before starting

#### 🔹 `[Service]`

* Defines how the process runs
* `ExecStart=` → command to run your program
* `WorkingDirectory=` → directory where the program starts
* `Restart=always` → restart if it crashes
* `User=` / `Group=` → drop privileges (don’t run as root)

#### 🔹 `[Install]`

* Defines when it should start automatically
* `WantedBy=multi-user.target` → start on boot in “multi-user” (normal) mode

---

## 3️⃣ Lifecycle of a Service

Once your service file is created, you can:

```bash
# Reload systemd to pick up new service
sudo systemctl daemon-reload

# Start service now
sudo systemctl start myapp.service

# Enable service to run at boot
sudo systemctl enable myapp.service

# Stop service
sudo systemctl stop myapp.service

# Restart service
sudo systemctl restart myapp.service

# Check status
sudo systemctl status myapp.service

# View logs (journalctl integrates with systemd logging)
journalctl -u myapp.service -f
```

---

## 4️⃣ How systemd manages it

When you start a service:

1. **systemd forks** and launches your program as a child process.
2. It monitors the process ID (PID).
3. If `Restart=always` is set, systemd restarts it if it exits.
4. Logs (`stdout` and `stderr`) are captured by **journald**, which you read via `journalctl`.

---

## 5️⃣ Why not just use `nohup` or `&`?

* If you run `python3 app.py &`, the program runs in background but:
  ❌ Doesn’t restart on crash
  ❌ Doesn’t start on boot
  ❌ Hard to manage logs

Systemd solves all of that by supervising your process.

---

## 6️⃣ Special Features (Advanced)

Systemd can do much more:

* **Resource control** (`CPUQuota=50%`, `MemoryLimit=500M`)
* **Sandboxing** (`ProtectSystem=yes`, `PrivateTmp=yes`) for security
* **Dependencies** (start after databases, networking, etc.)
* **Timers** (cron replacement for scheduling)
* **Sockets** (on-demand service activation)

---

✅ **Summary**:
Creating a Linux service = writing a `.service` file → placing it in `/etc/systemd/system/` → managing it with `systemctl`.
Systemd ensures it runs reliably, securely, and integrates with logs, boot, and resource management.

---

Would you like me to draw you a **mental flowchart** (boot → systemd → service starts → monitored → logs → restart on crash) so you can *visualize* how services are managed internally?
