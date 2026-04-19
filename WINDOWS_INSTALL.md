# 🪟 Windows Installation Guide

This guide will help you get **CogniCash** running on your Windows computer using Docker. No programming knowledge required!

---

## 🛠️ Step 1: Pre-requisites (The "Must-Haves")

Before we start, you need to install two main tools:

### 1. Install Docker Desktop
Docker is like a "shipping container" system for software. It lets CogniCash run on your Windows PC exactly like it does on a server.
- **Download:** [Docker Desktop for Windows](https://www.docker.com/products/docker-desktop/)
- **Installation:** Run the installer. If it asks about **WSL 2** (Windows Subsystem for Linux), say **YES** and let it install.
- **Restart:** You will likely need to restart your computer.
- **Check:** Once restarted, look for the little whale icon in your taskbar. It should say "Docker Desktop is running".

### 2. Install Git
Git is used to download the code and keep it updated.
- **Download:** [Git for Windows](https://git-scm.com/download/win)
- **Installation:** Click "Next" on all default options.

---

## 🔑 Step 2: Creating Your Security Keys

Security keys are like unique passwords for different parts of the app. We will use **PowerShell** (Windows' built-in command tool) to create them.

1.  Press the **Windows Key**, type `PowerShell`, and press **Enter**.
2.  Type or copy-paste these commands one by one:

### Create a JWT Secret (Your App's Signature)
```powershell
-join ((48..57) + (97..102) | Get-Random -Count 64 | ForEach-Object {[char]$_})
```
*👉 **Action:** Copy the random string of letters and numbers it gives you. Save it in a Notepad file for a moment.*

---

## 📥 Step 3: Downloading CogniCash

1.  Open **File Explorer** and go to a folder where you want to keep the app (e.g., `C:\Users\YourName\Documents`).
2.  Right-click in the empty space and select **"Open Git Bash here"** (or use the PowerShell window from Step 2).
3.  Type this command:
    ```bash
    git clone https://github.com/steierma/cogni-cash.git
    cd cogni-cash
    ```

---

## 📝 Step 4: Configuration (The .env file)

We need to tell the app what its settings are.

1.  Inside the `cogni-cash` folder, go into the `backend` folder.
2.  Find the file named `.env.example`.
3.  **Right-click it → Copy**, then **Right-click → Paste**. 
4.  Rename the copy to exactly `.env` (delete the `.example` part).
5.  **Right-click `.env` → Open with → Notepad**.
6.  Update these specific lines:

```dotenv
# Paste that random string you got in Step 2 here:
JWT_SECRET=your_random_string_here

# Choose a username and a VERY strong password for your first login:
ADMIN_USERNAME=admin
ADMIN_PASSWORD=choose_a_strong_password

# Choose a strong password for the database (you won't need to type this often):
POSTGRES_PASSWORD=another_strong_password

# If you have Ollama installed on Windows:
OLLAMA_URL=http://host.docker.internal:11434
```

---

## 🚀 Step 5: Starting the App

Now for the magic part.

1.  Go back to your PowerShell or Git Bash window (make sure you are inside the `cogni-cash` folder).
2.  Type this command:
    ```bash
    docker compose up -d
    ```
    *Note: This will take a few minutes the first time as it downloads everything needed.*

3.  **Wait until it finishes.** You can check the status in the **Docker Desktop Dashboard**.

---

## 🌐 Step 6: Accessing CogniCash

Once everything is "Running" (Green in Docker Desktop):

1.  Open your browser (Chrome, Edge, Firefox).
2.  Go to: **[http://localhost:3000](http://localhost:3000)**
3.  Log in using the `ADMIN_USERNAME` and `ADMIN_PASSWORD` you chose in Step 4.

---

## 🏦 Optional: Live Banking (Enable Banking)

If you want to sync real bank accounts:

1.  Register at [enablebanking.com](https://enablebanking.com).
2.  Create an application to get your `APP_ID`.
3.  Download your **Private Key** (a file ending in `.pem`).
4.  Rename that file to `enable-banking-prod.pem`.
5.  Place it directly in the main `cogni-cash` folder (the one containing `backend`, `frontend`, etc.).
6.  Open `backend/.env` again and add your `ENABLE_BANKING_APP_ID=...`.
7.  Restart the app by typing `docker compose restart` in your terminal.

---

## 🛠️ Advanced: Local HTTPS & "Hosts File" Workaround (Enable Banking)

If you are using **Enable Banking** for real-time sync, it strictly requires an **HTTPS** redirect URL. Even if you are running the app locally, you can "trick" your computer into using a custom domain name like `cognicash.local`.

### 1. Edit the Windows "Hosts" File
This tells Windows that `cognicash.local` points directly to your own computer.

1.  Press the **Windows Key**, type `Notepad`, then right-click it and select **"Run as Administrator"** (this is important!).
2.  In Notepad, go to **File → Open**.
3.  Paste this path into the address bar at the top and press Enter: `C:\Windows\System32\drivers\etc`.
4.  Change the file type dropdown in the bottom right from `Text Documents (*.txt)` to `All Files (*.*)`.
5.  Select the file named **`hosts`** and open it.
6.  Add this line at the very bottom:
    ```text
    127.0.0.1    cognicash.local
    ```
7.  **Save** the file and close Notepad.

### 2. Update your CogniCash Configuration
1.  Open `backend/.env` again in Notepad.
2.  Update the `DOMAIN_NAME` line:
    ```dotenv
    DOMAIN_NAME=cognicash.local
    ```
3.  Go to your [Enable Banking Dashboard](https://enablebanking.com) and update your application's **Redirect URL** to: `https://cognicash.local/api/v1/bank/finish/` (Note the **https** and the custom domain).

### 3. Accessing via HTTPS
1.  Restart your app in the terminal: `docker compose restart`.
2.  In your browser, go to **[https://cognicash.local:3000](https://cognicash.local:3000)**.
3.  Your browser will show a **"Your connection is not private"** warning because the security certificate is local. This is normal! Click **"Advanced"** and then **"Proceed to cognicash.local (unsafe)"**.

---

## ❓ Troubleshooting (For Windows Users)

- **"is a directory" error:** If you see an error about `enable-banking-prod.pem` being a directory, it means you started the app before creating the file. Delete the folder named `enable-banking-prod.pem` in your main directory and create the real file as explained in Step 4/Optional.
- **Port 3000 is busy:** Make sure you don't have other web development tools running on port 3000.
- **Ollama connection:** Ensure Ollama is running on your Windows taskbar. If it's not connecting, check your firewall settings to allow Docker to talk to your PC.

**Need Help?** Open an issue on GitHub!
