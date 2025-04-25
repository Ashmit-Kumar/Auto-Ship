/backend
├── cmd/
│   └── server/                # Main app entry point
│       └── main.go
├── internal/
│   ├── api/                   # API route handlers
│   │   ├── auth.go            # Login, signup, GitHub OAuth handlers
│   │   └── projects.go        # Clone/build/host endpoints
│   ├── services/              # Core logic (cloning, building, hosting)
│   │   ├── clone.go
│   │   ├── build.go
│   │   └── host.go
│   ├── models/                # Structs for users, projects, responses
│   ├── db/                    # DB connection and queries
│   │   ├── postgres.go
│   │   └── migrations/
│   ├── config/                # Config loader (env vars, secrets)
│   ├── middleware/            # Auth middleware, logging, rate limiting
│   └── utils/                 # Helper functions (validators, etc.)
├── scripts/                   # Bash scripts for setup/build
├── static/                    # Hosted project folders
├── Dockerfile
├── go.mod
├── go.sum
└── .env
