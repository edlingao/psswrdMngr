# Password Manager CLI

Simple command-line password manager.

## First Usage

On first run, `pwmngr` will prompt you to create a master password to encrypt your data.

## Usage

### List all entries
```bash
pwmngr list
```

### Add new entry
```bash
pwmngr add -t Title -f field1=value,field2=value,field3=value
```

### Set master password
```bash
pwmngr set -p Password
```

### Edit existing entry
```bash
pwmngr edit Title -f newField=value -e field1=newValue
```

## Examples

```bash
# Add a new entry
pwmngr add -t "GitHub" -f username=user@example.com,url=github.com

# Edit an entry (add new field and update existing)
pwmngr edit "GitHub" -f notes="Personal account" -e username=newuser@example.com
```
