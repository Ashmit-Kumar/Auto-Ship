# 🚀 AutoShip

AutoShip is a **one-click hosting platform** that lets you deploy your GitHub repositories instantly. Inspired by platforms like Vercel and Netlify, AutoShip automates the entire process—from cloning your repo to serving your app live with a unique URL.

---

## ⚡️ How It Works

1. **Submit Your Repo**
   - Enter your GitHub repository URL and (optionally) your `.env` file content.
2. **Automatic Setup**
   - AutoShip clones your repository, detects the project type (Node.js, Python, Go), and prepares the environment.
3. **Build & Deploy**
   - The system builds your project using Docker, runs it in an isolated container, and detects the correct internal port.
4. **Live Hosting**
   - Your app is hosted instantly. You get a unique live link (e.g., `username.autoship.app` or `autoship.app/project/12345`).
5. **Smart Routing**
   - Traefik reverse proxy automatically routes traffic to your running container.

---

## 🛠️ Tech Stack

- **Frontend:** Next.js, Tailwind CSS
- **Backend:** Go (Gin/Echo), Docker
- **Reverse Proxy:** Traefik (dynamic routing)
- **Database:** MongoDB (for project and port management)
- **Cloud:** AWS EC2

---

## 🔒 Security & Reliability

- **Input Validation:** Prevents malicious input and attacks.
- **Port Management:** Dynamically maps container ports to available host ports.
- **EC2 Security:** Automatically opens required ports in the AWS security group.
- **Cleanup:** Old builds and containers are removed automatically.
- **Logs:** Real-time build and deployment logs for transparency.

---

## 🚩 MVP Features

- [x] One-click GitHub repo deployment
- [x] Automatic environment detection (Node.js, Python, Go)
- [x] Docker-based build and run
- [x] Live link generation
- [x] Dynamic port and routing management
- [ ] Traefik-based smart routing (in progress)
- [ ] Project/port mapping persistence in MongoDB (in progress)

---

## 🤝 Contributing

Contributions and feedback are welcome!  
Feel free to open issues or submit pull requests.

---

AutoShip makes deployment effortless—just submit your repo and go live