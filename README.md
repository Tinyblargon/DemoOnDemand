# DemoOnDemand
### Application for creating VM-templates and deploying said templates in a vSphere environment.
# Production

## Prerequisites

Install the following packages:
- docker-ce
- docker-ce-cli
- containerd.io
- docker-compose-plugin

## Setup

1. Configure the config under `backend` inside the file `docker-compose.yml`.
2. Update docker volume mappings for `backend` and `db` in the `docker-compose.yml` file.
3. Update environment variables for `frontend` and `db` in the `docker-compose.yml` file.
4. Run docker-compose
```bash
docker compose up -d --build
```

# Development
## Backend

### install go 1.18

### Copy the example config file
```bash
cp backend/config.yml.example backend/config.yml
```
### configure settings in `backend/config.yml`

## Frontend

### Install npm
#### Debian
```bash
apt update
apt install npm
```
### Project setup
```bash
cd frontend
npm install
```
### Configure the URL of the API in .env file
```bash
cp .env.example .env.local
```
#### change `VUE_APP_ROOT_API=` in frontend/.env.local
### Compile and hot-load for development
```bash
npm run serve
```
