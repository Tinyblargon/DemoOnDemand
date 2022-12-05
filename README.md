# DemoOnDemand

# Production

configure config file
update docker volume mappings for 'backend' and 'db'

```bash
docker-compose up -d --build
```

# Development
## Backend
install go 1.18
```bash
cd backend
```
### Copy example config file
```bash
cp config.yml.example config.yml
```
configure settings in config file
### Running

run in debug mode with vscode

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
### Configure setting in .env file
```bash
cp .env.example .env.local
```
change 'VUE_APP_ROOT_API=' in frontend/.env
#### Compiles and hot-reloads for development
```bash
npm run serve
```
#### Compiles and minifies for production
```bash
npm run build
```
#### Lints and fixes files
```bash
npm run lint
```