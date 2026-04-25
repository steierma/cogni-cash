# 🪟 Windows Installation Guide

CogniCash is easiest to run on Windows using **Docker Desktop**.

## 🛠️ Step 1: Requirements
1. **Docker Desktop**: [Download & Install](https://www.docker.com/products/docker-desktop/).
   - Ensure **WSL 2** is enabled during installation.
   - Restart your computer if prompted.
2. **Git** (Optional but recommended): [Download & Install](https://git-scm.com/download/win).

---

## ⚡ Option 1: Quick Start (No Code Needed)
The fastest way to get CogniCash running.

1. **Create a folder** (e.g., `C:\Users\YourName\Documents\cognicash`).
2. **Download Config**:
   - Download [docker-compose.standalone.yml](https://raw.githubusercontent.com/steierma/cogni-cash/main/docker-compose.standalone.yml) and rename it to `docker-compose.yml`.
   - Download [.env.example](https://raw.githubusercontent.com/steierma/cogni-cash/main/backend/.env.example) and rename it to `.env`.
3. **Configure**:
   - Open `.env` with Notepad.
   - Set `JWT_SECRET` to any random string.
   - Set `POSTGRES_PASSWORD` and `ADMIN_PASSWORD`.
   - Set `OLLAMA_URL=http://host.docker.internal:11434` (if using Ollama on Windows).
4. **Run**:
   - Open **PowerShell** in that folder.
   - Type: `docker compose up -d`

Access the app at **http://localhost**.

---

## Option 2: Full Installation (Clone Repository)
Recommended if you want the full source code or to use the Makefile.

1. Open PowerShell and run:
   ```powershell
   git clone https://github.com/steierma/cogni-cash.git
   cd cogni-cash
   cp backend/.env.example backend/.env
   # Edit backend/.env as described above
   docker compose up -d
   ```

---

## 💡 Windows Tips & Troubleshooting

### Ollama Integration
If you have Ollama installed directly on Windows, ensure it is running in your taskbar. Docker connects to it using the special address: `http://host.docker.internal:11434`.

### "is a directory" error
If Docker complains that `enable-banking-prod.pem` is a directory:
1. Stop the app in Docker Desktop.
2. Delete the **folder** named `enable-banking-prod.pem` in your cognicash directory.
3. Create a new **empty file** with that exact name (using Notepad).
4. Start the app again.

### Accessing the App
By default, the app is available at **http://localhost**. If port 80 is used by another Windows service (like IIS), you may need to stop that service or change the ports in `docker-compose.yml`.
