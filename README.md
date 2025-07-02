# Forum Project ✨ *✧˚ · . 🚀 ⋆ ˚｡⋆ 🌟 ⋆ · ⋆ ✦ . * ✵

A web-based forum application built with Go and styled using Tailwind CSS.

## Project Setup

- This project uses Tailwind CSS for styling.
- The `static/css/output.css` file is precompiled and included in the repository.
- You don't need Node.js or Tailwind CLI to review this project, as all styles are already generated.

## Development

If you're making changes to the templates or styles:

1. Install dependencies:
   ```bash
   npm install
   ```

2. Run the Tailwind CSS build process:
   ```bash
   # For one-time build:
   npm run build

   # For development with auto-rebuild on changes:
   npm run watch
   ```

This will automatically rebuild the CSS file whenever you make changes to your HTML templates.

# Running the Project

## Project Structure

- `main.go` - Main application file
- `static/` - Static assets including CSS
- `templates/` - HTML templates
- `database/` - Database related files

## Dependencies

- Go standard library
- SQLite for database
- Tailwind CSS (precompiled) 

## Prerequisites

- Docker installed on your system
- Docker Compose installed
- (Optional) SQLite3 CLI for database inspection

## Setup Instructions

### 1. Add Your API Keys

Create a file named `apiKeys.sh` in the project root with your API keys:

```bash
#!/bin/bash
export NASA_API_KEY="your_nasa_api_key_here"
export GITHUB_CLIENT_ID="your_github_client_id_here"
export GITHUB_CLIENT_SECRET="your_github_client_secret_here"
export GOOGLE_CLIENT_ID="your_google_client_id_here"
export GOOGLE_CLIENT_SECRET="your_google_client_secret_here"
echo "All API keys have been exported."
```

**Do not commit this file to git.**

### 2. Make the Setup Script Executable (First Time Only)

If you see a permission error when running the script, make it executable:

```sh
chmod +x forum/run.sh
```

### 3. Run the Setup Script

From the project root, run:

```sh
./run.sh
```

This script will:
- Ensure the database file is set up correctly
- Source your API keys
- Build and start the Docker containers
- Show container status and logs


### 4. Access the Application

Once running, access the forum at:
- http://localhost:8080

## Stopping and Restarting

To stop the containers:
```sh
docker-compose down
```

To restart:
```sh
docker-compose restart
```

To rebuild:
```sh
docker-compose up -d --build
```

## Troubleshooting
- If you see a database error about `forum.db` being a directory, delete the `forum.db` directory and rerun the setup script.
- If you see a permission error running the script, use `chmod +x forum/run.sh` or `sh forum/run.sh`.
- If port 8080 is in use, stop the other service or change the port mapping in `docker-compose.yml`.

## Security Notes
- **Never commit API keys to Git.**