# AutoShip Backend Progress Log

## Overview

This document summarizes the backend work completed so far for the AutoShip dynamic website hosting platform. The backend automates deploying user-submitted repositories as live web services on an EC2 instance, supporting Node.js, Python, and Go projects.

---

## Features Implemented

### 1. Repository Handling
- **Cloning:**  
  - Users can submit a repository URL.
  - The backend clones the repository into a user-specific directory using the `CloneRepository` function.

### 2. Environment Setup
- **.env File Handling:**  
  - If the user provides `.env` content, it is saved into the cloned repository directory using the `SaveEnvFile` utility.

### 3. Environment Detection
- **Project Type Detection:**  
  - The backend inspects the repository to determine if it is a Node.js, Python, or Go project (see `DetectProjectType`).

### 4. Dockerfile Management
- **Template Copying:**  
  - Based on the detected environment, the corresponding Dockerfile template is copied into the repository directory.

### 5. Docker Build & Run
- **Image Building:**  
  - The backend builds a Docker image for the project using the generated Dockerfile.
- **Temporary Container:**  
  - A temporary container is run to help detect the internal port the application listens on.

### 6. Port Detection (Container)
- **Automatic Port Detection:**  
  - The backend attempts to detect the internal port by:
    - Checking common defaults (3000, 5000, 8080, etc.).
    - Running `netstat` inside the container if defaults fail.

### 7. Configuration Management
- **Environment Variables:**  
  - MongoDB URI, database/collection names, EC2 security group ID, and AWS region are now loaded from a `.env` file using the `godotenv` package, making configuration flexible and secure.

---

## Work In Progress / Not Yet Completed

### 1. Host Port Management
- **Current State:**  
  - The backend checks if the detected container port is available on the EC2 host.
- **To Do:**  
  - If the container port is not available, the backend should:
    - Find a free host port.
    - Map `<hostPort>:<containerPort>` when running the final container.
    - Open the selected host port in the EC2 security group.
    - Save the mapping (hostPort, containerPort) in MongoDB for tracking.

### 2. Final Container Launch
- **Current State:**  
  - The logic for running the final container with dynamic port mapping is not yet implemented.

### 3. Port Mapping Persistence
- **To Do:**  
  - Implement logic to save the mapping between the container port and the host port in MongoDB.

### 4. Routing with Traefik
- **To Do:**  
  - Integrate Traefik as a reverse proxy to handle routing for multiple containers and domains dynamically.

---

## Next Steps

1. **Complete Host Port Management:**
   - Implement logic to find a free host port if the container port is unavailable.
   - Update Docker run commands to use dynamic port mapping.
   - Open the host port in the EC2 security group.

2. **Persist Port Mappings:**
   - Save the mapping information in MongoDB for future reference and management.

3. **Add Routing with Traefik:**
   - Integrate Traefik for dynamic routing and domain management for hosted containers.

4. **Testing & Error Handling:**
   - Add robust error handling and logging.
   - Test the full pipeline with various project types and edge cases.

---

## Summary

- Repository cloning, environment detection, Dockerfile management, and container port detection are complete.
- Configuration is now environment-driven via `.env`.
- Host port mapping, persistence, and Traefik-based routing are the main remaining tasks.

---