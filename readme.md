# Codify

A secure, isolated playground for rapid prototyping in React and Node.js.

## Overview

Codify provides developers with isolated environments for quick experimentation and learning. Built with security and ease of use in mind, each playground runs in its own Kubernetes container, allowing safe and efficient code testing.



https://github.com/user-attachments/assets/ace1bc01-1029-4bbd-85d0-2265f730e440




## Architecture
![arch](https://github.com/user-attachments/assets/9caaff8a-30bc-4d89-9189-2c17761512eb)



- **Frontend**: Next.js-based interface for code editing and execution
- **Backend**: Golang server managing user authentication and code processing
- **Codeground Manager**: Spins up codegrounds on demand(k8s orchestration)
- **Infrastructure**: Kubernetes-based containerization for isolated playground environments

## Key Features

- **Language Support**: Full development environments for React and Node.js
- **Isolated Execution**: Each playground runs in a separate Kubernetes pod
- **Access Control**: Controlled external access through dedicated ingress routes
- **Real-time Preview**: Instant access to running applications via temporary URLs

