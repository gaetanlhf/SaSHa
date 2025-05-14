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
- ✅ **Configuration inheritance** propagating settings from groups to subgroups and servers
- ✅ **Import functionality** to modularize your server configurations
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

### Configuration Components

#### Server Configuration Options

Each server can have the following properties:

- `name`: Display name for the server (required)
- `host`: Hostname or IP address (required)
- `user`: SSH username (optional)
- `port`: SSH port (optional, defaults to 22)
- `color`: Custom color for this server (optional, in hex format)
- `extra_args`: Additional SSH command-line arguments (optional)
- `ssh_binary`: Custom SSH binary to use (optional, defaults to "ssh")

#### Group Configuration Options

Each group can have the following properties:

- `name`: Display name for the group (required)
- `hosts`: List of servers in this group
- `groups`: List of subgroups
- `color`: Custom color for this group (optional, in hex format)
- `user`: Default username for all servers in this group (optional)
- `port`: Default port for all servers in this group (optional)
- `extra_args`: Default additional SSH arguments for all servers in this group (optional)
- `ssh_binary`: Default SSH binary for all servers in this group (optional)

### Configuration Features

#### Configuration Inheritance

SaSHa implements a powerful inheritance system for configuration properties. Settings defined at a group level are automatically propagated to all of its children (both subgroups and servers) unless explicitly overridden.

This approach is especially useful in team environments, where common settings (like a team username or standard SSH options) can be defined once and automatically applied to all servers within that group.

The inheritance applies to the following properties:
- `user`: SSH username
- `port`: SSH port
- `extra_args`: Additional SSH command-line arguments
- `ssh_binary`: Custom SSH binary
- `color`: Visual theme color

For example:

```yaml
groups:
  - name: Production
    user: prod-user         # This user will be inherited by all child items
    port: 2222              # This port will be inherited by all child items
    ssh_binary: ssh-custom  # This binary will be inherited by all child items
    color: "#FF5733"        # This color will be inherited by all child items
    extra_args: ["-v"]      # These arguments will be inherited by all child items

    hosts:
      - name: Web Server
        host: web.example.com
        # Inherits user, port, ssh_binary, color and extra_args from parent

      - name: Database
        host: db.example.com
        user: db-admin      # Overrides the inherited user
        port: 3333          # Overrides the inherited port

    groups:
      - name: Europe
        color: "#3366FF"    # Overrides the inherited color
        hosts:
          - name: EU Server
            host: eu.example.com
            # Inherits user, port, ssh_binary from Production
            # But inherits color from Europe
```

This inheritance system makes it easy to define common settings once at a higher level and have them automatically applied to all nested items, while still allowing for flexibility when you need to override specific settings for individual servers or subgroups.

#### Importing Configurations

For complex setups, you can split your configuration across multiple files and import them. This is particularly useful for team environments where configurations can be shared, or for organizing large server infrastructures:

```yaml
# Main config file
imports:
  - file: ~/.sasha/production-servers.yaml
  - file: ~/.sasha/development-servers.yaml
    user: devuser  # Override settings for all imported servers

groups:
  - name: Local
    hosts:
      - name: Localhost
        host: 127.0.0.1
```

In the imported files, simply define groups and servers:

```yaml
# production-servers.yaml
groups:
  - name: Production
    hosts:
      - name: Web Server
        host: web.example.com
        user: admin
hosts:
  - name: Standalone Server
    host: server.example.com
```

You can also import within groups, which helps organize servers by department, team, or project:

```yaml
groups:
  - name: Development
    imports:
      - file: ~/.sasha/dev-servers.yaml
        # Imported servers will be placed in the Development group
```

Using version control for your configuration files allows you to track infrastructure changes over time and easily share server access configurations with team members. This approach treats your SSH access management as "configuration as code."

Each import directive can have the following properties:

- `file`: Path to the YAML file to import (required)
- `path`: Group path where the imported items will be placed (optional)
- `user`: Default username for all imported servers (optional)
- `port`: Default port for all imported servers (optional)
- `extra_args`: Default additional SSH arguments for all imported servers (optional)
- `ssh_binary`: Default SSH binary for all imported servers (optional)
- `color`: Default color for all imported groups and servers (optional)

#### Remote Imports and Caching

SaSHa supports importing configurations from remote URLs, allowing teams to share server configurations from central repositories:

```yaml
imports:
  - file: https://config-server.example.com/team-servers.yaml
    # Credentials for authenticated imports
    auth:
      username: user
      password: pass
      # Or use a token instead
      # token: access_token
      # header: Authorization  # Optional, defaults to "Authorization"
```

Remote imports are cached locally to improve performance and allow offline use. You can control caching behavior:

```yaml
# Global cache timeout in hours (default: 24)
cache_timeout: 48

# Disable caching for specific imports
imports:
  - file: https://config-server.example.com/team-servers.yaml
    no_cache: true  # Always fetch the latest version

# Disable caching for all imports
no_cache: true
```

You can manually clear the cache using:

```bash
sasha --clear-cache
# Or refresh without exiting
sasha --refresh-cache
```

#### Authentication for Imports

SaSHa supports authenticated imports with several authentication methods:

```yaml
imports:
  - file: https://config-server.example.com/team-servers.yaml
    auth:
      # Basic authentication
      username: user
      password: pass

      # Or token-based authentication
      # token: your_access_token
      # header: Authorization  # Optional, defaults to "Authorization"
```

Authentication can also be specified at the group level and will be used for all imports within that group:

```yaml
groups:
  - name: Team
    auth:
      token: team_access_token
    imports:
      - file: https://config-server.example.com/team-servers.yaml
        # Will use the group's auth configuration
```

#### Theme Customization

You can customize the appearance of SaSHa by assigning colors to groups and servers. Colors are specified in hexadecimal format (e.g., `#FF5733`).

Colors defined in groups are inherited by their subgroups and servers unless overridden. This is part of SaSHa's comprehensive inheritance system that applies to colors, SSH settings, and other properties.

```yaml
groups:
  - name: Production
    color: "#FF5733"  # All servers and subgroups will use this color unless overridden
    hosts:
      - name: Important Server
        host: important.example.com
        color: "#3366FF"  # This server overrides the parent group's color
```

SaSHa's interface will dynamically update its theme based on your current location in the group hierarchy, reflecting the color associated with your current group or server. This provides visual context about where you are in your server organization.

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