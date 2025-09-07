<h2 align="center">SaSHa</h2>
<p align="center">A TUI SSH connection manager with hierarchical organization</p>
<p align="center">
    <a href="#about">About</a> •
    <a href="#features">Features</a> •
    <a href="#installation">Installation</a> •
    <a href="#configuration">Configuration</a> •
    <a href="#usage">Usage</a> •
    <a href="#license">License</a>
</p>

## About

SaSHa is a terminal-based SSH connection manager designed to simplify managing and connecting to multiple SSH servers. It provides a beautiful and intuitive text user interface (TUI) that organizes servers in a hierarchical structure with customizable colors and navigation.

## Features

- ✅ **Intuitive TUI** with colors and navigation
- ✅ **Hierarchical organization** of servers using groups and subgroups
- ✅ **Connection history** to quickly access recently used servers
- ✅ **Favorites system** to mark and easily access important servers
- ✅ **Filtering** to quickly find servers in large configurations
- ✅ **Custom SSH options** per server or group (port, user, password, additional arguments)
- ✅ **Automatic password authentication** with secure PTY handling
- ✅ **Theme customization** with colors for servers and groups
- ✅ **Global configuration inheritance** with settings at the root level that cascade to all groups and servers
- ✅ **Group-level configuration inheritance** propagating settings from groups to subgroups and servers
- ✅ **Inheritance with explicit overrides**
- ✅ **Import functionality** to modularize your server configurations with remote URL support
- ✅ **Authentication for remote imports** with token and basic auth support
- ✅ **Smart caching system** for remote imports with flexible cron-based scheduling
- ✅ **Keyboard shortcuts** for efficient navigation and operation
- ✅ **Top-level group filtering** to limit display to specific groups

## Installation

### Using Pre-compiled Binaries

SaSHa provides pre-compiled binaries for various platforms through GitHub releases:

```bash
# Linux (x64)
curl -L https://github.com/gaetanlhf/SaSHa/releases/latest/download/sasha-linux-amd64 -o sasha
chmod +x sasha
sudo mv sasha /usr/local/bin/

# Linux (ARM64)
curl -L https://github.com/gaetanlhf/SaSHa/releases/latest/download/sasha-linux-arm64 -o sasha
chmod +x sasha
sudo mv sasha /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/gaetanlhf/SaSHa/releases/latest/download/sasha-darwin-amd64 -o sasha
chmod +x sasha
sudo mv sasha /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/gaetanlhf/SaSHa/releases/latest/download/sasha-darwin-arm64 -o sasha
chmod +x sasha
sudo mv sasha /usr/local/bin/
```

You can also manually download the appropriate binary for your system from the [GitHub Releases page](https://github.com/gaetanlhf/SaSHa/releases).

### Building from Source

If you prefer to build from source, ensure you have **Golang** (1.23 or newer) installed on your machine.

```bash
# Clone the repository
git clone https://github.com/gaetanlhf/SaSHa.git
cd SaSHa

# Build the application
make build

# Install the application
sudo make install
```

## Configuration

SaSHa uses a YAML configuration file to define server groups, individual servers, and application preferences.

### Configuration File Location

By default, SaSHa looks for the configuration file at `~/.sasha/config.yaml`. You can specify a different location by setting the `SASHA_HOME` environment variable:

```bash
export SASHA_HOME=/path/to/your/sasha/directory
```

### Basic Configuration Example

Here's a simple example showing the main configuration structure:

```yaml
# Application features configuration
features:
  history_size: 20           # Number of history entries to keep (0 to disable)
  favorites_enabled: true    # Enable or disable favorites feature
  cache_schedule: "0 8 * * *" # Cache update schedule using cron format
  allow_web_imports: false   # Enable/disable remote HTTP/HTTPS imports

# Define root-level groups
groups:
  - name: Production
    color: "#FF5733"  # Custom color for this group
    hosts:
      - name: Web Server
        host: web.example.com
        user: admin
        password: mypassword123  # Automatic password authentication
      - name: Database
        host: db.example.com
        user: dbadmin
        port: 2222
        password: dbpass456

  - name: Development
    color: "#33FF57"
    hosts:
      - name: Dev Server
        host: dev.example.com
        user: developer
    groups:
      - name: Testing
        hosts:
          - name: QA Server
            host: qa.example.com
            user: tester
            password: testpass789

# Define root-level standalone servers
hosts:
  - name: Home Server
    host: homeserver.local
    user: admin
    password: homepass123
    color: "#5733FF"  # Custom color for this server
```

### Advanced Configuration with Global Settings

For larger setups, you can define global settings that will be inherited by all groups and servers:

```yaml
# Optional global settings - inherited by everything unless overridden
user: admin                    # Default user for all servers
port: 2222                     # Default port for all servers  
password: default_password     # Default password for all servers
color: "#3366FF"               # Default color theme
ssh_binary: ssh                # Default SSH binary
extra_args: ["-o StrictHostKeyChecking=no"]  # Default SSH arguments
no_cache: false                # Global caching behavior
auth:                          # Default auth for remote imports
  token: global_token

# Application features
features:
  history_size: 20             # Keep 20 connection history entries
  favorites_enabled: true      # Enable favorites system
  cache_schedule: "0 8 * * *"  # Daily cache updates at 8 AM
  allow_web_imports: true      # Allow remote HTTP/HTTPS imports

groups:
  - name: Production
    color: "#FF5733"  # Overrides global color for this group and its children
    password: prod_password  # Override global password for production servers
    hosts:
      - name: Web Server
        host: web.example.com
        # Inherits user, port, password from global settings
        # But uses Production group's color and password
      - name: Database
        host: db.example.com
        user: dbadmin  # Overrides global user
        port: 3306     # Overrides global port
        password: special_db_password  # Overrides group password

hosts:
  - name: Home Server
    host: homeserver.local
    # Inherits all global settings (user, port, password, color, ssh_binary, extra_args)
```

### Features Configuration

The `features` section controls application behavior and security settings:

#### History Configuration
```yaml
features:
  history_size: 20    # Keep 20 recent connections (default: 20)
  # history_size: 0   # Disable history completely
  # history_size: -1  # Use default (20)
```

#### Favorites System
```yaml
features:
  favorites_enabled: true   # Enable favorites (default: false)
  # favorites_enabled: false # Disable favorites
```

#### Cache Management
```yaml
features:
  cache_schedule: "0 8 * * *"  # Daily at 8 AM (cron format)
  # cache_schedule: "*/30 * * * *"  # Every 30 minutes
  # cache_schedule: "0 9 * * 1"     # Weekly on Monday at 9 AM
```

#### Web Import Security
```yaml
features:
  allow_web_imports: false  # Disable remote HTTP/HTTPS imports (default: false)
  # allow_web_imports: true # Allow remote imports
```

When `allow_web_imports` is set to `false`, SaSHa will block all HTTP/HTTPS imports for security, while still allowing local file imports. This provides granular control over remote configuration access.

### Password Authentication

SaSHa supports automatic password authentication for SSH connections. When a password is configured, SaSHa will:

1. Automatically detect password prompts during SSH connection
2. Provide the password securely without user interaction
3. Handle authentication failures gracefully
4. Display a clean "Connecting to [hostname]..." message

**Password configuration examples:**

```yaml
# Global password for all servers
password: company_default_password

groups:
  - name: Production
    password: production_password  # Override for production servers
    hosts:
      - name: Web Server
        host: web.example.com
        # Uses production_password

      - name: Special Server
        host: special.example.com
        password: unique_password  # Override with server-specific password

      - name: Key-based Server
        host: keyauth.example.com
        password: ""  # Explicit override: disable password auth for this server

  - name: Development
    # No password specified - inherits global password
    hosts:
      - name: Dev Server
        host: dev.example.com
        # Uses company_default_password from global config
```

**Security considerations:**
- Passwords are stored in plain text in the configuration file
- Ensure proper file permissions on your configuration file (`chmod 600 ~/.sasha/config.yaml`)
- Consider using SSH keys instead of passwords when possible
- Use different passwords for different environment tiers (development, staging, production)

**Password inheritance behavior:**
- Like other settings, passwords follow the inheritance hierarchy
- Setting `password: ""` explicitly disables password authentication for that server
- Servers without password configuration will attempt key-based authentication

### Configuration Components

#### Global Configuration Options

At the root level of your configuration, you can optionally define global settings that will be inherited by all groups and servers:

- `user`: Default SSH username for all servers (optional)
- `port`: Default SSH port for all servers (optional)
- `password`: Default SSH password for all servers (optional)
- `color`: Default color theme for all groups and servers (optional)
- `extra_args`: Default additional SSH command-line arguments for all servers (optional)
- `ssh_binary`: Default SSH binary to use for all servers (optional)
- `auth`: Default authentication configuration for remote imports (optional)
- `no_cache`: Disable caching for all remote imports (optional)

#### Cache Schedule Configuration

The `cache_schedule` setting uses cron format to control when cached remote imports should be refreshed. This provides flexible scheduling based on your needs:

```yaml
# Common cache schedule examples:

# Every 6 hours (default if not specified)
features:
  cache_schedule: "0 */6 * * *"

# Daily at 8 AM
features:
  cache_schedule: "0 8 * * *"

# Daily at 9 PM
features:
  cache_schedule: "0 21 * * *"

# Twice daily (8 AM and 8 PM)
features:
  cache_schedule: "0 8,20 * * *"

# Weekly on Monday at 9 AM
features:
  cache_schedule: "0 9 * * 1"

# Every hour
features:
  cache_schedule: "0 * * * *"

# Every 30 minutes
features:
  cache_schedule: "*/30 * * * *"
```

**Cron format:** `minute hour day-of-month month day-of-week`
- minute: 0-59
- hour: 0-23
- day-of-month: 1-31
- month: 1-12
- day-of-week: 0-7 (0 and 7 are Sunday)

#### Server Configuration Options

Each server can have the following properties:

- `name`: Display name for the server (required)
- `host`: Hostname or IP address (required)
- `user`: SSH username (optional, inherits from parent group or global)
- `port`: SSH port (optional, inherits from parent group or global, defaults to 22)
- `password`: SSH password for automatic authentication (optional, inherits from parent group or global)
- `color`: Custom color for this server (optional, inherits from parent group or global)
- `extra_args`: Additional SSH command-line arguments (optional, inherits from parent group or global)
- `ssh_binary`: Custom SSH binary to use (optional, inherits from parent group or global)

#### Group Configuration Options

Each group can have the following properties:

- `name`: Display name for the group (required)
- `hosts`: List of servers in this group
- `groups`: List of subgroups
- `color`: Custom color for this group (optional, inherits from parent group or global)
- `user`: Default username for all servers in this group (optional, inherits from parent group or global)
- `port`: Default port for all servers in this group (optional, inherits from parent group or global)
- `password`: Default password for all servers in this group (optional, inherits from parent group or global)
- `extra_args`: Default additional SSH arguments for all servers in this group (optional, inherits from parent group or global)
- `ssh_binary`: Default SSH binary for all servers in this group (optional, inherits from parent group or global)
- `auth`: Authentication configuration for imports within this group (optional, inherits from parent group or global)
- `no_cache`: Disable caching for imports within this group (optional, inherits from parent group or global)

### Configuration Features

#### Multi-Level Configuration Inheritance

SaSHa implements a comprehensive multi-level inheritance system with precise control over when inheritance occurs. Settings can be defined at multiple levels and are automatically propagated down the hierarchy unless explicitly overridden.

**Inheritance Behavior:**
- **Field not specified**: Inherits from parent (group or global)
- **Field set to a value**: Uses that specific value
- **Field set to empty string (`""`) or zero (`0`)**: Explicitly prevents inheritance and excludes the parameter from the SSH command

This allows for precise control - you can inherit most settings while explicitly disabling specific ones where needed.

**Inheritance Order (from highest to lowest priority):**
1. **Global Level** (root of config file)
2. **Parent Group Level**
3. **Current Group Level**
4. **Server Level** (highest priority, final override)

This hierarchical approach is especially powerful in enterprise environments, where organizational standards (like company-wide SSH keys or connection settings) can be defined once at the global level and automatically applied throughout the entire infrastructure.

**The inheritance applies to the following properties:**
- `user`: SSH username
- `port`: SSH port
- `password`: SSH password
- `extra_args`: Additional SSH command-line arguments
- `ssh_binary`: Custom SSH binary
- `color`: Visual theme color
- `auth`: Authentication configuration for remote imports
- `no_cache`: Caching behavior for remote imports

**Complete inheritance example with explicit overrides:**

```yaml
# Global settings - define your organizational defaults
user: company-admin              # Standard admin user for all servers
port: 2222                      # Company standard SSH port
password: company_default_pass   # Default password for password-auth servers
ssh_binary: ssh                 # Standard SSH client
auth:                           # Company API token for remote configs
  token: company_access_token

features:
  cache_schedule: "0 8 * * *"     # Daily cache updates at 8 AM
  allow_web_imports: true         # Allow remote configuration imports

groups:
  - name: Production
    color: "#FF0000"              # Make production red for visibility
    password: secure_prod_pass    # Override with production-specific password
    
    hosts:
      - name: Web Server
        host: web.prod.company.com
        # Inherits: user=company-admin, port=2222, password=secure_prod_pass, ssh_binary=ssh, color=#FF0000
        
      - name: Database
        host: db.prod.company.com
        user: db-admin              # Override: use db-admin instead of company-admin
        port: 5432                  # Override: PostgreSQL port instead of 2222
        password: special_db_pass   # Override: database-specific password
        
      - name: Key-only Server
        host: secure.prod.company.com
        password: ""                # Explicit override: disable password auth, use keys only
        port: 0                     # Explicit override: use default SSH port (22)
        # Result: ssh company-admin@secure.prod.company.com (no password, default port)

  - name: Development
    color: "#00CC66"              # Green for development
    user: dev-user                # Override global user for dev servers
    password: dev_password        # Override global password for dev servers
    
    hosts:
      - name: Dev Server
        host: dev.company.com
        # Inherits: user=dev-user, port=2222, password=dev_password, color=#00CC66
        
      - name: Public Dev Server  
        host: public-dev.company.com
        user: ""                    # Explicit override: no username needed
        password: ""                # Explicit override: no password needed
        port: 0                     # Explicit override: use standard port 22
        # Result: ssh public-dev.company.com (no auth)

hosts:
  - name: Jump Host
    host: jump.company.com
    # Inherits all global settings: user=company-admin, port=2222, password=company_default_pass, etc.
    
  - name: Public Server
    host: public.company.com  
    user: ""                      # Explicit override: no username required
    password: ""                  # Explicit override: no password required
    # Result: ssh public.company.com (no auth)
```

**SSH Command Examples from the above config:**
- Web Server: `ssh -p 2222 company-admin@web.prod.company.com` (with password: secure_prod_pass)
- Database: `ssh -p 5432 db-admin@db.prod.company.com` (with password: special_db_pass)
- Key-only Server: `ssh company-admin@secure.prod.company.com` (key-based auth, default port)
- Dev Server: `ssh -p 2222 dev-user@dev.company.com` (with password: dev_password)
- Public Dev Server: `ssh public-dev.company.com` (no auth, default port)
- Jump Host: `ssh -p 2222 company-admin@jump.company.com` (with password: company_default_pass)
- Public Server: `ssh public.company.com` (no auth)

This smart inheritance system allows you to:
- **Start simple**: Just define `name` and `host` for servers - everything else is optional
- **Set standards**: Define global defaults only for settings you want to standardize
- **Override selectively**: Change only what's different at each level
- **Explicitly disable**: Use empty strings or zero values to prevent inheritance
- **Maintain consistency**: Ensure common settings are applied automatically
- **Stay flexible**: Override any setting at any level when needed
- **Secure by default**: Define secure passwords globally and override only when necessary

#### Importing Configurations

For complex setups, you can split your configuration across multiple files and import them. This is particularly useful for team environments where configurations can be shared, or for organizing large server infrastructures:

```yaml
# Main config file - define only what you want to standardize globally
user: company-admin     # Optional: only if you want a default user
password: company_pass  # Optional: only if you want a default password
auth:                   # Optional: only if you use authenticated remote imports
  token: company_access_token

features:
  cache_schedule: "0 8 * * *"  # Optional: daily updates at 8 AM
  allow_web_imports: true      # Allow remote configuration imports

imports:
  - file: ~/.sasha/production-servers.yaml
  - file: ~/.sasha/development-servers.yaml
    user: devuser      # Override global user for all imported dev servers
    password: devpass  # Override global password for all imported dev servers
  - file: https://config.company.com/shared-servers.yaml
    # Uses global auth and cache schedule automatically if defined

groups:
  - name: Local
    imports:
      - file: ~/.sasha/team-servers.yaml
        # Imported servers inherit Local group settings + global settings
    hosts:
      - name: Localhost
        host: 127.0.0.1
        # Minimal config - just name and host required
```

In the imported files, simply define groups and servers that will inherit from the importing context:

```yaml
# production-servers.yaml
groups:
  - name: Production
    color: "#FF0000"
    password: prod_secure_pass  # Override imported password for production
    hosts:
      - name: Web Server
        host: web.example.com
        # Will inherit user, port, password from Production group
        # Will inherit color from Production group

      - name: No-Auth Server
        host: public.example.com
        user: ""        # Explicit override: no username needed
        password: ""    # Explicit override: no password needed
        # Result: ssh public.example.com

hosts:
  - name: Standalone Server
    host: server.example.com
    # Will inherit all settings from main config including password
```

Using version control for your configuration files allows you to track infrastructure changes over time and easily share server access configurations with team members. This approach treats your SSH access management as "configuration as code."

**Each import directive supports all the same properties as groups:**

- `file`: Path to the YAML file to import (required)
- `path`: Group path where the imported items will be placed (optional)
- `user`: Override username for all imported servers (optional)
- `port`: Override port for all imported servers (optional)
- `password`: Override password for all imported servers (optional)
- `extra_args`: Override additional SSH arguments for all imported servers (optional)
- `ssh_binary`: Override SSH binary for all imported servers (optional)
- `color`: Override color for all imported groups and servers (optional)
- `auth`: Override authentication for nested remote imports (optional)
- `no_cache`: Override caching behavior for this import (optional)

#### Remote Imports and Caching

SaSHa supports importing configurations from remote URLs, allowing teams to share server configurations from central repositories. Remote imports can be controlled via the `allow_web_imports` feature flag:

```yaml
features:
  allow_web_imports: true      # Enable remote HTTP/HTTPS imports
  cache_schedule: "0 9 * * *"  # Daily updates at 9 AM

# Global auth used by all remote imports unless overridden
auth:
  token: company_access_token

imports:
  - file: https://config-server.example.com/team-servers.yaml
    # Uses global auth and cache schedule automatically

  - file: https://other-server.example.com/special-servers.yaml
    # Override auth for this specific import
    auth:
      username: user
      password: pass
      # Or use a token instead
      # token: different_access_token
      # header: Authorization  # Optional, defaults to "Authorization"
```

**Security Note**: When `allow_web_imports` is set to `false` (default), all HTTP/HTTPS imports will be blocked with a clear error message, while local file imports continue to work normally. This provides granular security control over remote configuration access.

Remote imports are cached locally to improve performance and allow offline use. The cache is updated based on the cron schedule you configure:

```yaml
features:
  cache_schedule: "0 */12 * * *"  # Update every 12 hours

# Global caching behavior
no_cache: false                 

groups:
  - name: Dynamic Config
    no_cache: true  # Disable caching for this group's imports
    imports:
      - file: https://config-server.example.com/dynamic-servers.yaml
        # Will not be cached due to group setting

  - name: Stable Config
    imports:
      - file: https://config-server.example.com/stable-servers.yaml
        # Uses global cache schedule (every 12 hours)

      - file: https://config-server.example.com/frequently-updated.yaml
        no_cache: true  # Override to disable cache for this specific import
```

You can manually manage the cache using:

```bash
sasha --clear-cache      # Clear cache and exit
sasha --refresh-cache    # Clear cache but continue loading
```

**Cache Schedule Examples:**

- `"0 8 * * *"` - Daily at 8 AM
- `"0 9 * * 1"` - Weekly on Monday at 9 AM
- `"*/30 * * * *"` - Every 30 minutes
- `"0 8,20 * * *"` - Twice daily at 8 AM and 8 PM
- Not specified - No automatic refresh (cache only updated manually)

#### Authentication for Remote Imports

SaSHa supports authenticated imports with several authentication methods. Authentication can be configured at any level and will be inherited down the hierarchy:

```yaml
# Global authentication - used by all remote imports
auth:
  token: global_access_token

groups:
  - name: External Team
    # Override global auth for this team's imports
    auth:
      username: team_user
      password: team_pass
    imports:
      - file: https://external-config.example.com/team-servers.yaml
        # Uses team auth automatically

  - name: Special Projects
    imports:
      - file: https://config-server.example.com/public-servers.yaml
        # Uses global auth

      - file: https://special-server.example.com/private-servers.yaml
        # Override with specific auth for this import only
        auth:
          token: special_project_token
          header: X-API-Key  # Custom header name
```

**Authentication methods supported:**
- **Basic Authentication**: `username` and `password`
- **Token Authentication**: `token` with optional custom `header` (defaults to "Authorization")

#### Theme Customization

You can customize the appearance of SaSHa by assigning colors at any level. Colors are specified in hexadecimal format (e.g., `#FF5733`) and follow the same inheritance rules as other properties.

```yaml
# Global color theme
color: "#0066CC"  # Company blue for everything

groups:
  - name: Production
    color: "#FF0000"  # Override to red for production - critical systems
    hosts:
      - name: Critical Server
        host: critical.prod.example.com
        color: "#FF6600"  # Override to orange - extra attention needed

  - name: Development
    color: "#00CC66"  # Override to green for development - safe to experiment
    groups:
      - name: Staging
        # Inherits green from Development
        hosts:
          - name: Staging Server
            host: staging.example.com
            # Inherits green color from Development group
```

SaSHa's interface will dynamically update its theme based on your current location in the group hierarchy, reflecting the color associated with your current group or server. This provides visual context about where you are in your server organization and the criticality level of your current context.

#### History and Favorites

SaSHa automatically maintains a history of your connections. You can access it by pressing `h`. The history records:

- Server name and connection details
- Path to the server in your group hierarchy
- Timestamp of the connection

The history size can be configured in the features section:

```yaml
features:
  # Store up to 20 history entries
  history_size: 20

  # Disable history
  # history_size: 0
```

You can mark servers as favorites for quick access. To toggle a server's favorite status, select it and press `f`. Access all of your favorites by pressing `Tab`.

Enable or disable the favorites feature in the features section:

```yaml
features:
  # Enable favorites
  favorites_enabled: true

  # Disable favorites
  # favorites_enabled: false
```

## Usage

Once configured, simply run the `sasha` command to launch the application:

```bash
sasha
```

### Command Line Options

SaSHa supports several command-line options:

```
Usage: sasha [options]

Options:
  -clear-cache                 Clear the import cache and exit
  -refresh-cache               Clear the cache but continue loading the application
  -clear-history               Clear connection history
  -clear-favorites             Clear favorites
  -filter-top-level-groups     Only load specified top-level groups (comma-separated)
  -version                     Print version information
  -help                        Show this help message

Environment variables:
  SASHA_HOME                   Path to SaSHa home directory (default: ~/.sasha)
```

#### Top-Level Group Filtering

The `--filter-top-level-groups` option allows you to limit SaSHa to only display specific top-level groups from your configuration. This is particularly useful for separating different contexts (e.g., work vs. personal servers) or when you only need to work with a subset of your infrastructure.

```bash
# Load only work-related servers
sasha --filter-top-level-groups=Work

# Load only personal servers
sasha --filter-top-level-groups=Personal

# Load multiple contexts
sasha --filter-top-level-groups=Work,Personal

# Load only client-specific servers
sasha --filter-top-level-groups=ClientA
```

When using this option:
- Only the specified top-level groups will be loaded
- If only one group is specified, SaSHa will start directly inside that group (making it the new "Home")
- History and favorites will be filtered to only show entries from the loaded groups

**Example with configuration:**
```yaml
groups:
  - name: Work
    hosts:
      - name: Company Server
        host: server.company.com
  - name: Personal
    hosts:
      - name: Home Server
        host: homeserver.local
  - name: ClientA
    hosts:
      - name: Client Server
        host: client.example.com
```

Using `sasha --filter-top-level-groups=Work` will only show the Work group and its contents, effectively hiding Personal and ClientA groups for this session.

### Navigation

- Use arrow keys or `j`/`k` to navigate up and down
- Press `Enter` to select a server or enter a group
- Press `Esc` or `Backspace` to go back
- Press `/` to filter servers
- Press `?` to toggle help display
- Press `h` to access connection history
- Press `Tab` to access favorites
- Press `f` to toggle favorite status of the selected server
- Press `q` to quit

When you select a server, SaSHa will display "Connecting to [hostname]..." and execute the SSH connection for you. If a password is configured, it will be provided automatically during the authentication process.

## License

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see http://www.gnu.org/licenses/.