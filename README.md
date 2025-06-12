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

SaSHa (SSH Assistant) is a terminal-based SSH connection manager designed to simplify managing and connecting to multiple SSH servers. It provides a beautiful and intuitive text user interface (TUI) that organizes servers in a hierarchical structure with customizable colors and navigation.

## Features

- ✅ **Intuitive TUI** with colors and navigation
- ✅ **Hierarchical organization** of servers using groups and subgroups
- ✅ **Connection history** to quickly access recently used servers
- ✅ **Favorites system** to mark and easily access important servers
- ✅ **Filtering** to quickly find servers in large configurations
- ✅ **Custom SSH options** per server or group (port, user, additional arguments)
- ✅ **Theme customization** with colors for servers and groups
- ✅ **Global configuration inheritance** with settings at the root level that cascade to all groups and servers
- ✅ **Group-level configuration inheritance** propagating settings from groups to subgroups and servers
- ✅ **Import functionality** to modularize your server configurations with remote URL support
- ✅ **Authentication for remote imports** with token and basic auth support
- ✅ **Smart caching system** for remote imports with configurable timeout
- ✅ **Keyboard shortcuts** for efficient navigation and operation

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
# Define history size (number of entries to keep)
# Set to 0 to disable history
history_size: 20

# Enable or disable favorites feature
favorites_enabled: true

# Define root-level groups
groups:
  - name: Production
    color: "#FF5733"  # Custom color for this group
    hosts:
      - name: Web Server
        host: web.example.com
        user: admin
      - name: Database
        host: db.example.com
        user: dbadmin
        port: 2222

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

# Define root-level standalone servers
hosts:
  - name: Home Server
    host: homeserver.local
    user: admin
    color: "#5733FF"  # Custom color for this server
```

### Advanced Configuration with Global Settings

For larger setups, you can define global settings that will be inherited by all groups and servers:

```yaml
# Optional global settings - inherited by everything unless overridden
user: admin                    # Default user for all servers
port: 2222                     # Default port for all servers  
color: "#3366FF"               # Default color theme
ssh_binary: ssh                # Default SSH binary
extra_args: ["-o StrictHostKeyChecking=no"]  # Default SSH arguments
auth:                          # Default auth for remote imports
  token: global_token

groups:
  - name: Production
    color: "#FF5733"  # Overrides global color for this group and its children
    hosts:
      - name: Web Server
        host: web.example.com
        # Inherits user, port, auth from global settings
        # But uses Production group's color
      - name: Database
        host: db.example.com
        user: dbadmin  # Overrides global user
        port: 3306     # Overrides global port

hosts:
  - name: Home Server
    host: homeserver.local
    # Inherits all global settings (user, port, color, ssh_binary, extra_args)
```

### Configuration Components

#### Global Configuration Options

At the root level of your configuration, you can optionally define global settings that will be inherited by all groups and servers:

- `user`: Default SSH username for all servers (optional)
- `port`: Default SSH port for all servers (optional)
- `color`: Default color theme for all groups and servers (optional)
- `extra_args`: Default additional SSH command-line arguments for all servers (optional)
- `ssh_binary`: Default SSH binary to use for all servers (optional)
- `auth`: Default authentication configuration for remote imports (optional)
- `no_cache`: Disable caching for all remote imports (optional)

#### Server Configuration Options

Each server can have the following properties:

- `name`: Display name for the server (required)
- `host`: Hostname or IP address (required)
- `user`: SSH username (optional, inherits from parent group or global)
- `port`: SSH port (optional, inherits from parent group or global, defaults to 22)
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
- `extra_args`: Default additional SSH arguments for all servers in this group (optional, inherits from parent group or global)
- `ssh_binary`: Default SSH binary for all servers in this group (optional, inherits from parent group or global)
- `auth`: Authentication configuration for imports within this group (optional, inherits from parent group or global)
- `no_cache`: Disable caching for imports within this group (optional, inherits from parent group or global)

### Configuration Features

#### Multi-Level Configuration Inheritance

SaSHa implements a comprehensive multi-level inheritance system for configuration properties. Settings can be defined at multiple levels and are automatically propagated down the hierarchy unless explicitly overridden.

**Inheritance Order (from highest to lowest priority):**
1. **Global Level** (root of config file)
2. **Parent Group Level**
3. **Current Group Level**
4. **Server Level** (highest priority, final override)

This hierarchical approach is especially powerful in enterprise environments, where organizational standards (like company-wide SSH keys or connection settings) can be defined once at the global level and automatically applied throughout the entire infrastructure.

**The inheritance applies to the following properties:**
- `user`: SSH username
- `port`: SSH port
- `extra_args`: Additional SSH command-line arguments
- `ssh_binary`: Custom SSH binary
- `color`: Visual theme color
- `auth`: Authentication configuration for remote imports
- `no_cache`: Caching behavior for remote imports

**Complete inheritance example:**

```yaml
# Optional global settings - only define what you want to standardize
user: company-admin              # Optional: set if you want a default user
auth:                           # Optional: set if you use authenticated remote imports
  token: company_access_token

groups:
  - name: Production
    # Only override what you need to change
    color: "#FF0000"  # Make production red for visibility
    
    hosts:
      - name: Web Server
        host: web.prod.company.com
        # Only 'host' is required, everything else inherits or uses defaults
        
      - name: Database
        host: db.prod.company.com
        user: db-admin      # Override only if different from global
        port: 5432          # Override only if different from default (22)

    groups:
      - name: Europe
        color: "#FF6600"    # Optional: different color for European servers
        user: eu-admin      # Optional: different user for European team
        
        hosts:
          - name: EU Web Server
            host: web.eu.prod.company.com
            # Minimal config - inherits everything else

# Minimal server config - just needs name and host
hosts:
  - name: Jump Host
    host: jump.company.com
    # Inherits global user if defined, uses defaults for everything else
```

This multi-level inheritance system allows you to:
- **Start simple**: Just define `name` and `host` for servers - everything else is optional
- **Set standards**: Define global defaults only for settings you want to standardize
- **Override selectively**: Change only what's different at each level
- **Maintain consistency**: Ensure common settings are applied automatically
- **Stay flexible**: Override any setting at any level when needed

#### Importing Configurations

For complex setups, you can split your configuration across multiple files and import them. This is particularly useful for team environments where configurations can be shared, or for organizing large server infrastructures:

```yaml
# Main config file - define only what you want to standardize globally
user: company-admin     # Optional: only if you want a default user
auth:                   # Optional: only if you use authenticated remote imports
  token: company_access_token

imports:
  - file: ~/.sasha/production-servers.yaml
  - file: ~/.sasha/development-servers.yaml
    user: devuser  # Override global user for all imported dev servers
  - file: https://config.company.com/shared-servers.yaml
    # Uses global auth automatically if defined

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
    hosts:
      - name: Web Server
        host: web.example.com
        # Will inherit user, port, auth from main config
        # Will inherit color from Production group
hosts:
  - name: Standalone Server
    host: server.example.com
    # Will inherit all settings from main config
```

Using version control for your configuration files allows you to track infrastructure changes over time and easily share server access configurations with team members. This approach treats your SSH access management as "configuration as code."

**Each import directive supports all the same properties as groups:**

- `file`: Path to the YAML file to import (required)
- `path`: Group path where the imported items will be placed (optional)
- `user`: Override username for all imported servers (optional)
- `port`: Override port for all imported servers (optional)
- `extra_args`: Override additional SSH arguments for all imported servers (optional)
- `ssh_binary`: Override SSH binary for all imported servers (optional)
- `color`: Override color for all imported groups and servers (optional)
- `auth`: Override authentication for nested remote imports (optional)
- `no_cache`: Override caching behavior for this import (optional)

#### Remote Imports and Caching

SaSHa supports importing configurations from remote URLs, allowing teams to share server configurations from central repositories:

```yaml
# Global auth used by all remote imports unless overridden
auth:
  token: company_access_token

imports:
  - file: https://config-server.example.com/team-servers.yaml
    # Uses global auth automatically

  - file: https://other-server.example.com/special-servers.yaml
    # Override auth for this specific import
    auth:
      username: user
      password: pass
      # Or use a token instead
      # token: different_access_token
      # header: Authorization  # Optional, defaults to "Authorization"
```

Remote imports are cached locally to improve performance and allow offline use. You can control caching behavior at multiple levels:

```yaml
# Global cache settings
cache_timeout: 48  # Cache timeout in hours (default: 24)
no_cache: false    # Global caching behavior

groups:
  - name: Dynamic Config
    no_cache: true  # Disable caching for this group's imports
    imports:
      - file: https://config-server.example.com/dynamic-servers.yaml
        # Will not be cached due to group setting

  - name: Stable Config
    imports:
      - file: https://config-server.example.com/stable-servers.yaml
        # Uses global cache settings (48 hours)

      - file: https://config-server.example.com/frequently-updated.yaml
        no_cache: true  # Override to disable cache for this specific import
```

You can manually manage the cache using:

```bash
sasha --clear-cache      # Clear cache and exit
sasha --refresh-cache    # Clear cache but continue loading
```

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

The history size can be configured in the configuration file:

```yaml
# Store up to 20 history entries
history_size: 20

# Disable history
# history_size: 0
```

You can mark servers as favorites for quick access. To toggle a server's favorite status, select it and press `f`. Access all of your favorites by pressing `Tab`.

Enable or disable the favorites feature in the configuration file:

```yaml
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
  -clear-cache       Clear the import cache and exit
  -refresh-cache     Clear the cache but continue loading the application
  -clear-history     Clear connection history
  -clear-favorites   Clear favorites
  -version           Print version information
  -help              Show this help message

Environment variables:
  SASHA_HOME         Path to SaSHa home directory (default: ~/.sasha)
```

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

When you select a server, SaSHa will build the SSH command and execute it for you.

## License

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see http://www.gnu.org/licenses/.