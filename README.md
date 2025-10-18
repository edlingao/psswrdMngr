# Password Manager

Secure TUI-based password manager built with Go and SQLite.

## Features

- **Master Password Protection** - All data encrypted with master password
- **Groups** - Organize entries into categories (Work, Personal, etc.)
- **Flexible Fields** - Store any key-value pairs (username, password, URL, notes, etc.)
- **Terminal UI** - Interactive interface using Bubble Tea
- **Local Storage** - SQLite database stored locally at `~/.mngr/main.db`
- **Configurable** - Customize settings via `~/.mngr/config.yaml`

## Installation

### Prerequisites

- Go 1.21+
- SQLite3

### Build from Source

```bash
git clone <repository-url>
cd password_mngr
./setup.sh
go build -o psswrdMngr
```

The `setup.sh` script will:
- Create `~/.mngr/` directory
- Generate default `config.yaml`
- Initialize `main.db` with schema

### Manual Setup

```bash
mkdir -p ~/.mngr
cat > ~/.mngr/config.yaml << EOF
db_path: ~/.mngr/main.db
EOF
sqlite3 ~/.mngr/main.db < db/schema.sql
go build -o psswrdMngr
```

## First Run

On first launch, you'll be prompted to create a master password. This password encrypts all your data and is required for every session.

```bash
./psswrdMngr
```

## Usage

### Main Interface

The TUI provides an interactive menu-driven interface:

1. **Login** - Enter master password
2. **Groups** - Manage entry categories
3. **Secured Entries** - Manage password entries
4. **Fields** - Add/edit/delete fields within entries

### Workflow Example

1. Start the application: `./psswrdMngr`
2. Enter master password
3. Create a group (e.g., "Work", "Personal")
4. Create a secured entry (e.g., "GitHub Account")
5. Add fields to the entry:
   - `username`: your_email@example.com
   - `password`: your_secure_password
   - `url`: https://github.com
   - `notes`: Personal account

### Navigation

- **Arrow keys** - Navigate menu options
- **Enter** - Select option
- **Esc** - Go back/Cancel
- **Tab** - Switch between input fields

## Configuration

Edit `~/.mngr/config.yaml` to customize settings:

```yaml
db_path: ~/.mngr/main.db
```

### Custom Database Location

```yaml
db_path: /path/to/custom/database.db
```

## Database Schema

The application uses four tables:

- **Password** - Stores encrypted master password
- **Groups** - Categories for organizing entries
- **Secured** - Password entries with titles and group assignments
- **Fields** - Key-value pairs for each secured entry

## Security Notes

- Master password is hashed and stored securely
- Database is stored locally (not cloud-synced)
- Recommended: Regular backups of `~/.mngr/main.db`
- No password recovery - master password must be remembered

## Backup & Restore

### Backup

```bash
cp ~/.mngr/main.db ~/.mngr/main.db.backup
```

### Restore

```bash
cp ~/.mngr/main.db.backup ~/.mngr/main.db
```

## Development

### Project Structure

```
password_mngr/
├── internal/
│   ├── config/         # Configuration management
│   ├── group/          # Group domain logic
│   ├── password/       # Master password logic
│   └── secured/        # Secured entries & fields
├── configurator/       # Dependency injection
├── db/
│   ├── schema.sql      # Database schema
│   └── queries/        # SQL queries
├── main.go             # Entry point
└── setup.sh            # Setup script
```

### Build

```bash
go build -o psswrdMngr
```

### Run Tests

```bash
go test ./...
```

## Troubleshooting

### "Failed to initialize config"

Ensure `~/.mngr/` directory exists and is writable:

```bash
mkdir -p ~/.mngr
chmod 755 ~/.mngr
```

### "Error connecting to DB"

Check database path in config and file permissions:

```bash
ls -la ~/.mngr/main.db
sqlite3 ~/.mngr/main.db "SELECT 1;"
```

### Reset Master Password

Run the reset script:

```bash
sqlite3 ~/.mngr/main.db < db/queries/resetPassword.sql
```

**Warning**: This will require you to set a new master password but won't decrypt existing data encrypted with the old password.

## License

MIT License - Free and open source. See LICENSE file for details.
