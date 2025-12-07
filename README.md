# fscomposer

[![Go Reference](https://pkg.go.dev/badge/github.com/absfs/fscomposer.svg)](https://pkg.go.dev/github.com/absfs/fscomposer)
[![Go Report Card](https://goreportcard.com/badge/github.com/absfs/fscomposer)](https://goreportcard.com/report/github.com/absfs/fscomposer)
[![CI](https://github.com/absfs/fscomposer/actions/workflows/ci.yml/badge.svg)](https://github.com/absfs/fscomposer/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**Visual Filesystem Composition Studio for the absfs Ecosystem**

A modern, isometric drag-and-drop interface for building complex filesystem stacks from composable [absfs](https://github.com/absfs/absfs) nodes. Think of it as a visual IDE for filesystem architecture.

## Vision

Create arbitrarily complex filesystem compositions through an intuitive visual interface with isometric 2.5D blocks, similar to modern game world builders. Connect backends (S3, SFTP, local disk), apply transformations (encryption, compression, caching), enforce policies (permissions, quotas), and mount the result via FUSE, WebDAV, or NFS.

### Example Composition

```
[FUSE Mount] ← Visual representation of:
      ↑
[metricsfs] (Prometheus)
      ↑
  [permfs] (ACL for /public, /private)
      ↑
 [cachefs] (1GB LRU cache)
      ↑
 [retryfs] (3x exponential backoff)
      ↑
[encryptfs] (AES-256-GCM)
      ↑
  ┌──────┴──────┐
  │             │
[s3fs]     [sftpfs]
(primary)  (fallback)
```

**Result:** Encrypted, cached, permissioned cloud storage mounted as local filesystem with metrics and automatic retry.

## Design Philosophy

### User Experience
- **Visual First:** Drag and drop blocks like building with LEGO
- **Dark Theme:** Black-on-black with neon accent colors (cyan, magenta, yellow, green)
- **Isometric 2.5D:** Modern game-like visual style
- **Real-time Preview:** See configuration changes immediately
- **Export Anywhere:** Deploy to FUSE, Docker, Kubernetes, or standalone binary

### Technical Approach
- **Hybrid Architecture:** Core wrappers built-in, plugin system for extensions
- **Multiple Deploy Options:** Runtime composition or compiled binary
- **Type-Safe:** Validate connections at design time
- **Production Ready:** Generate deployment-ready code/containers

## UI Design Concept

### Node Types & Visual Language

#### Backend Nodes (Foundation - Dark Blue/Purple)
- **osfs** - Local disk (folder icon)
- **memfs** - In-memory (RAM chip icon)
- **s3fs** - AWS S3 (cloud icon)
- **sftpfs** - SFTP server (network icon)
- **webdavfs** - WebDAV (web icon)
- **boltfs** - BoltDB (database icon)
- **httpfs** - HTTP (globe icon)

#### Wrapper Nodes (Middleware - Various Colors)
- **cachefs** - Cache layer (⚡ lightning - yellow)
- **encryptfs** - Encryption (🔒 lock - red)
- **compressfs** - Compression (📦 package - orange)
- **retryfs** - Retry logic (🔄 circular arrows - orange)
- **metricsfs** - Observability (📊 graph - green)
- **unionfs** - Multi-layer (📚 layers - cyan)
- **permfs** - Access control (👤 user - magenta)
- **quotafs** - Storage limits (💾 disk - blue)
- **switchfs** - Path routing (🔀 switch - purple)
- **logfs** - Audit logging (📝 log - white)

#### Mount Point Nodes (Output - Bright White)
- **FUSE Mount** - Local mount point
- **WebDAV Server** - HTTP/WebDAV endpoint
- **NFS Server** - NFS export
- **HTTP API** - REST API endpoint

### Connection Rules

Nodes have typed input/output ports:

**Single Input/Output (Pass-through wrappers):**
```
┌─────────────┐
│  encryptfs  │
│             │
│   ▲    ▼    │
│             │
└─────────────┘
```

**Multiple Inputs (Composition wrappers):**
```
┌─────────────┐
│  unionfs    │
│             │
│ ▲ ▲ ▲  ▼   │
│             │
└─────────────┘
```

**Multiple Outputs (Routing wrappers):**
```
┌─────────────┐
│  switchfs   │
│             │
│   ▲    ▼▼▼  │
│             │
└─────────────┘
```

### Configuration Panels

Clicking a node opens a context-sensitive configuration panel:

**Example: encryptfs Configuration**
```
┌────────────────────────────┐
│ Encryption Configuration   │
├────────────────────────────┤
│ Algorithm: AES-256-GCM ▼   │
│ Key Source: Environment ▼  │
│   ENV_VAR: ENCRYPT_KEY     │
│ ☑ Encrypt filenames        │
│ ☑ Encrypt metadata         │
│ Cipher Mode: GCM ▼         │
│                            │
│ [Validate] [Save] [Cancel] │
└────────────────────────────┘
```

**Example: cachefs Configuration**
```
┌────────────────────────────┐
│ Cache Configuration        │
├────────────────────────────┤
│ Cache Size: 1024 MB        │
│ Policy: LRU ▼              │
│   LRU / LFU / ARC          │
│ TTL: 300 seconds           │
│ ☑ Cache metadata           │
│ ☐ Write-through            │
│ ☑ Write-back               │
│                            │
│ [Validate] [Save] [Cancel] │
└────────────────────────────┘
```

## Architecture Options

### Option 1: Pre-Built Service (Fast, Limited)

**Architecture:**
```
┌─────────────────────────────────────┐
│  Web UI (Svelte/React)              │
│  - Drag & drop canvas               │
│  - Configuration panels             │
│  - Composition persistence          │
└────────────┬────────────────────────┘
             │ REST API / WebSocket
┌────────────▼────────────────────────┐
│  Backend Service (Go)               │
│  - All absfs implementations        │
│  - Dynamic composition engine       │
│  - Runtime configuration            │
│  - FUSE/WebDAV/NFS servers          │
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│  Composed Filesystem                │
│  - Live mount points                │
│  - Hot-reload on config change      │
└─────────────────────────────────────┘
```

**Pros:**
- Instant deployment
- No compilation needed
- Live reconfiguration without restart
- Hot-reload compositions

**Cons:**
- All nodes must be pre-compiled into binary
- Larger binary size (~100MB+)
- Can't use custom/third-party wrappers without rebuild
- Higher memory footprint

**Best for:** Quick prototyping, development, simple deployments

---

### Option 2: Code Generator + Hot Compile (Flexible, Slower)

**Architecture:**
```
┌─────────────────────────────────────┐
│  Composition UI (Electron/Web)      │
│  - Visual node graph editor         │
│  - Export to composition spec       │
│  - Template management              │
└────────────┬────────────────────────┘
             │ Composition Spec (JSON/YAML)
┌────────────▼────────────────────────┐
│  Build Container (Docker)           │
│  - Go template engine               │
│  - Code generation from spec        │
│  - go build + optimization          │
│  - docker build (optional)          │
└────────────┬────────────────────────┘
             │ Compiled Binary / Container
┌────────────▼────────────────────────┐
│  Custom Service Image               │
│  - Only needed dependencies         │
│  - Optimized binary (10-30MB)       │
│  - Single-purpose deployment        │
└─────────────────────────────────────┘
```

**Pros:**
- Smallest final binary (only used modules)
- Can use any Go module from any source
- No runtime composition overhead
- Production-optimized builds
- Can integrate custom/third-party wrappers

**Cons:**
- Build time (30-90 seconds)
- Requires Docker/Podman or Go toolchain
- Configuration changes need full rebuild
- More complex deployment pipeline

**Best for:** Production deployments, embedded systems, microservices

---

### Option 3: Hybrid Plugin System (Recommended)

**Architecture:**
```
┌─────────────────────────────────────┐
│  Web UI + Desktop App               │
│  - Svelte/SvelteKit frontend        │
│  - Electron wrapper (optional)      │
└────────────┬────────────────────────┘
             │ REST API + Plugin Registry
┌────────────▼────────────────────────┐
│  Core Service (Go)                  │
│  ├─ Built-in Wrappers               │
│  │  - cachefs, retryfs, etc.        │
│  ├─ Plugin Loader                   │
│  │  - Dynamic .so loading           │
│  │  - Version management            │
│  └─ Mount Handlers                  │
│     - FUSE, WebDAV, NFS             │
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│  Filesystem Stack                   │
│  - Runtime composition              │
│  - Hot-reload with validation       │
└─────────────────────────────────────┘
```

**Pros:**
- Core wrappers always available (fast)
- Can add custom plugins without rebuild
- Moderate binary size (~50MB)
- Fast startup time
- Extensible ecosystem

**Cons:**
- Go plugin system platform-specific
- Plugin versioning complexity
- Some overhead for plugin calls
- Requires careful ABI management

**Best for:** Balanced approach, community extensions, desktop app

---

## Implementation Plan

### Phase 1: Core Backend Engine

**Goal:** Build the composition engine that can dynamically construct filesystem stacks

**Components:**
1. **Composition Spec Format** (JSON/YAML)
   ```yaml
   version: "1.0"
   name: "encrypted-s3-cache"
   description: "Encrypted S3 with local cache"

   nodes:
     - id: s3-backend
       type: s3fs
       config:
         bucket: my-bucket
         region: us-east-1

     - id: encryption
       type: encryptfs
       config:
         algorithm: AES-256-GCM
         keySource: env
         keyEnv: ENCRYPT_KEY

     - id: cache
       type: cachefs
       config:
         size: 1073741824  # 1GB
         policy: LRU
         ttl: 300

   connections:
     - from: s3-backend
       to: encryption
     - from: encryption
       to: cache

   mount:
     type: fuse
     path: /mnt/composed
     root: cache
   ```

2. **Composition Engine** (`engine/`)
   - Spec parser and validator
   - Node registry (type → constructor mapping)
   - Dependency graph builder
   - Cycle detection
   - Stack instantiation

3. **Node Registry** (`registry/`)
   - Register all absfs wrapper constructors
   - Type validation
   - Configuration schema per node type
   - Version management

4. **Mount Handlers** (`mount/`)
   - FUSE mount wrapper
   - WebDAV server wrapper
   - NFS server wrapper
   - HTTP API wrapper

**Files to Create:**
```
engine/
  spec.go         # Composition spec structs
  parser.go       # YAML/JSON parsing
  validator.go    # Validate spec (cycles, types)
  builder.go      # Build filesystem stack

registry/
  registry.go     # Node type registration
  nodes.go        # All absfs wrapper registrations
  schema.go       # Configuration schemas

mount/
  fuse.go         # FUSE mounting
  webdav.go       # WebDAV server
  nfs.go          # NFS server
  api.go          # HTTP API

cmd/fscomposer/
  main.go         # CLI entry point
```

**Deliverables:**
- CLI tool that loads composition spec and mounts filesystem
- Comprehensive test suite
- Example composition specs

**Testing:**
```bash
# Load composition from spec
./fscomposer mount --spec examples/encrypted-s3.yaml

# Validate spec without mounting
./fscomposer validate --spec my-stack.yaml

# List available node types
./fscomposer nodes list
```

---

### Phase 2: REST API Server

**Goal:** Expose composition engine via REST API for UI integration

**Endpoints:**

**Composition Management:**
- `POST /api/compositions` - Create new composition
- `GET /api/compositions` - List all compositions
- `GET /api/compositions/{id}` - Get composition spec
- `PUT /api/compositions/{id}` - Update composition
- `DELETE /api/compositions/{id}` - Delete composition

**Lifecycle:**
- `POST /api/compositions/{id}/start` - Mount filesystem
- `POST /api/compositions/{id}/stop` - Unmount filesystem
- `GET /api/compositions/{id}/status` - Get mount status
- `POST /api/compositions/{id}/reload` - Hot-reload config

**Registry:**
- `GET /api/nodes` - List available node types
- `GET /api/nodes/{type}` - Get node schema/docs

**Validation:**
- `POST /api/validate` - Validate composition spec

**Monitoring:**
- `GET /api/compositions/{id}/metrics` - Get filesystem metrics (if metricsfs used)
- `GET /api/compositions/{id}/logs` - Stream logs (if logfs used)

**WebSocket:**
- `WS /api/compositions/{id}/events` - Real-time status updates

**Files to Create:**
```
api/
  server.go       # HTTP server setup
  handlers.go     # Request handlers
  middleware.go   # Auth, logging, CORS
  websocket.go    # WebSocket support

storage/
  compositions.go # Persist compositions (BoltDB/SQLite)

cmd/fscomposesd/
  main.go         # Daemon entry point
```

**Deliverables:**
- REST API server with OpenAPI spec
- WebSocket support for live updates
- Persistent composition storage
- Authentication/authorization hooks

**Testing:**
```bash
# Start server
./fscomposesd --port 8080

# Create composition via API
curl -X POST http://localhost:8080/api/compositions \
  -H "Content-Type: application/json" \
  -d @my-stack.json

# Mount it
curl -X POST http://localhost:8080/api/compositions/abc123/start
```

---

### Phase 3: Web UI - Canvas & Node Editor

**Goal:** Build the visual drag-and-drop interface

**Tech Stack:**
- **Framework:** Svelte + SvelteKit (smaller, faster than React)
- **Node Editor:** Custom canvas or [Rete.js](https://rete.js.org/)
- **Styling:** Tailwind CSS with dark theme
- **Isometric Rendering:** Custom SVG components
- **Build:** Vite

**UI Components:**

1. **Canvas Workspace** (`src/lib/canvas/`)
   - Drag-and-drop node placement
   - Connection line drawing
   - Zoom/pan navigation
   - Grid snapping
   - Undo/redo

2. **Node Palette** (`src/lib/palette/`)
   - Categorized node list (Backends, Wrappers, Mounts)
   - Search/filter
   - Drag to canvas

3. **Configuration Panel** (`src/lib/config/`)
   - Context-sensitive forms per node type
   - Real-time validation
   - Schema-driven UI generation

4. **Toolbar** (`src/lib/toolbar/`)
   - New/Open/Save composition
   - Validate/Deploy buttons
   - Settings

5. **Status Bar** (`src/lib/status/`)
   - Connection status
   - Validation errors
   - Active mounts

**Pages:**
- `/` - Dashboard (list compositions)
- `/compose` - Visual editor
- `/compose/:id` - Edit existing composition
- `/deploy` - Deployment options
- `/settings` - App settings

**Files to Create:**
```
frontend/
  src/
    lib/
      canvas/
        Canvas.svelte
        Node.svelte
        Connection.svelte
      palette/
        NodePalette.svelte
        NodeCard.svelte
      config/
        ConfigPanel.svelte
        FieldTypes.svelte
      api/
        client.ts        # API wrapper
    routes/
      +page.svelte       # Dashboard
      compose/
        +page.svelte     # Editor
      settings/
        +page.svelte     # Settings
    app.css              # Global styles
```

**Deliverables:**
- Fully functional visual editor
- Dark theme with neon accents
- Responsive design
- Export compositions to JSON/YAML

**Visual Design:**

**Color Palette:**
```css
/* Base */
--bg-primary: #0a0a0a;     /* Near black */
--bg-secondary: #1a1a1a;   /* Dark gray */
--bg-tertiary: #2a2a2a;    /* Medium gray */

/* Accents */
--accent-cyan: #00f0ff;
--accent-magenta: #ff00ff;
--accent-yellow: #ffff00;
--accent-green: #00ff00;
--accent-red: #ff0040;

/* Node Types */
--node-backend: #4a3a8a;   /* Deep purple */
--node-wrapper: #2a4a6a;   /* Deep blue */
--node-mount: #f0f0f0;     /* White */
```

---

### Phase 4: Deployment Integrations

**Goal:** Enable one-click deployment to various targets

**Deployment Targets:**

1. **Local FUSE Mount**
   - Direct mount via fscomposer CLI
   - Systemd service generation
   - Auto-start on boot

2. **Docker Container**
   - Generate Dockerfile
   - Build and push to registry
   - docker-compose.yml generation

3. **Kubernetes**
   - Generate K8s manifests (Deployment, Service, ConfigMap)
   - Helm chart generation
   - Volume mounts and secrets

4. **Standalone Binary**
   - Code generation
   - Cross-compilation
   - Packaging (tar.gz, deb, rpm)

**Deployment Wizard:**
```
┌────────────────────────────────┐
│ Deploy: encrypted-s3-cache     │
├────────────────────────────────┤
│ Target: Docker Container ▼     │
│                                │
│ Registry: ghcr.io/user/repo    │
│ Tag: latest                    │
│                                │
│ Environment Variables:         │
│   ENCRYPT_KEY: ••••••••        │
│   AWS_ACCESS_KEY: ••••••••     │
│                                │
│ ☑ Build multi-arch (amd64,arm) │
│ ☑ Push to registry             │
│                                │
│ [Build] [Build & Deploy]       │
└────────────────────────────────┘
```

**Code Generation Templates:**

**Dockerfile Template:**
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o fscomposer .

FROM alpine:latest
RUN apk --no-cache add fuse
COPY --from=builder /build/fscomposer /usr/local/bin/
COPY composition.yaml /etc/fscomposer/
ENTRYPOINT ["fscomposer", "mount", "--spec", "/etc/fscomposer/composition.yaml"]
```

**Kubernetes Deployment Template:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Name}}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{.Name}}
  template:
    metadata:
      labels:
        app: {{.Name}}
    spec:
      containers:
      - name: fscomposer
        image: {{.Image}}
        securityContext:
          privileged: true
        volumeMounts:
        - name: composition
          mountPath: /etc/fscomposer
      volumes:
      - name: composition
        configMap:
          name: {{.Name}}-config
```

**Files to Create:**
```
deploy/
  docker/
    template.go      # Dockerfile generation
    build.go         # Docker build/push
  kubernetes/
    template.go      # K8s manifest generation
    helm.go          # Helm chart generation
  systemd/
    template.go      # Service unit generation
  codegen/
    template.go      # Go code generation
    build.go         # Cross-compilation
```

**Deliverables:**
- Deployment wizard in UI
- Template-based code generation
- Multi-target deployment support
- Environment variable management

---

### Phase 5: Plugin System & Extension Marketplace

**Goal:** Allow community-contributed wrappers and backends

**Plugin Architecture:**

**Go Plugin Interface:**
```go
// plugin/interface.go
package plugin

type FSPlugin interface {
    // Metadata
    Name() string
    Version() string
    Description() string

    // Schema for configuration UI
    ConfigSchema() []ConfigField

    // Factory
    New(config map[string]interface{}, underlying absfs.FileSystem) (absfs.FileSystem, error)
}

type ConfigField struct {
    Name        string
    Type        string // "string", "int", "bool", "select"
    Required    bool
    Default     interface{}
    Description string
    Options     []string // For "select" type
}
```

**Plugin Discovery:**
```go
// Load plugin from .so file
plugin, err := plugin.Open("./plugins/myfs.so")

// Lookup exported symbol
sym, err := plugin.Lookup("Plugin")

// Type assert to interface
fsPlugin := sym.(FSPlugin)

// Register with registry
registry.Register(fsPlugin.Name(), fsPlugin)
```

**Plugin Development Kit:**
```
sdk/
  plugin.go       # Plugin interface
  helpers.go      # Common helpers
  testing.go      # Test utilities
  examples/
    wrapper/      # Example wrapper plugin
    backend/      # Example backend plugin
```

**Plugin Registry Service:**
- Central repository of community plugins
- Version management
- Security scanning
- Automatic updates

**UI Integration:**
- Browse plugin marketplace
- One-click install
- Version pinning
- Plugin settings

**Files to Create:**
```
plugin/
  interface.go    # Plugin interface
  loader.go       # Dynamic loading
  registry.go     # Plugin registry
  validator.go    # Plugin validation

marketplace/
  api.go          # Marketplace API client
  install.go      # Plugin installation
  update.go       # Plugin updates
```

**Deliverables:**
- Plugin SDK with examples
- Plugin marketplace (web UI)
- Secure plugin loading
- Documentation for plugin authors

---

### Phase 6: Advanced Features

**Goal:** Production-ready features for enterprise use

**Features:**

1. **Composition Templates**
   - Pre-built stacks for common use cases
   - Template marketplace
   - Template parameterization

2. **Version Control Integration**
   - Git integration for composition specs
   - Diff/merge compositions
   - Rollback support

3. **Team Collaboration**
   - Multi-user support
   - Role-based access control
   - Shared compositions

4. **Monitoring Dashboard**
   - Real-time metrics visualization (if metricsfs used)
   - Performance graphs
   - Alert configuration

5. **Testing & Simulation**
   - Dry-run mode (validate without mounting)
   - Performance simulation
   - Load testing tools

6. **Migration Tools**
   - Import from other systems
   - Export to different formats
   - Data migration helpers

**Files to Create:**
```
templates/
  library.go      # Template management
  marketplace.go  # Template sharing

vcs/
  git.go          # Git integration
  diff.go         # Composition diffing

monitoring/
  dashboard.go    # Metrics dashboard
  alerts.go       # Alert management

testing/
  simulator.go    # Performance simulation
  validator.go    # Integration testing
```

**Deliverables:**
- Template library with 10+ common patterns
- Git integration for version control
- Monitoring dashboard
- Testing/simulation tools

---

## Technical Specifications

### Composition Spec Format

**JSON Schema:**
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["version", "name", "nodes", "connections", "mount"],
  "properties": {
    "version": {
      "type": "string",
      "pattern": "^[0-9]+\\.[0-9]+$"
    },
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 64
    },
    "description": {
      "type": "string"
    },
    "nodes": {
      "type": "array",
      "items": {
        "$ref": "#/definitions/node"
      }
    },
    "connections": {
      "type": "array",
      "items": {
        "$ref": "#/definitions/connection"
      }
    },
    "mount": {
      "$ref": "#/definitions/mount"
    }
  },
  "definitions": {
    "node": {
      "type": "object",
      "required": ["id", "type"],
      "properties": {
        "id": {
          "type": "string",
          "pattern": "^[a-z0-9-]+$"
        },
        "type": {
          "type": "string",
          "enum": ["osfs", "memfs", "s3fs", "sftpfs", "webdavfs",
                   "cachefs", "encryptfs", "retryfs", "metricsfs",
                   "permfs", "unionfs", "switchfs"]
        },
        "config": {
          "type": "object"
        }
      }
    },
    "connection": {
      "type": "object",
      "required": ["from", "to"],
      "properties": {
        "from": {
          "type": "string"
        },
        "to": {
          "type": "string"
        }
      }
    },
    "mount": {
      "type": "object",
      "required": ["type", "root"],
      "properties": {
        "type": {
          "type": "string",
          "enum": ["fuse", "webdav", "nfs", "api"]
        },
        "path": {
          "type": "string"
        },
        "root": {
          "type": "string"
        },
        "options": {
          "type": "object"
        }
      }
    }
  }
}
```

### Node Type Specifications

Each node type has a configuration schema:

**s3fs:**
```yaml
type: s3fs
schema:
  - name: bucket
    type: string
    required: true
    description: S3 bucket name

  - name: region
    type: string
    required: true
    default: us-east-1
    description: AWS region

  - name: endpoint
    type: string
    required: false
    description: Custom S3 endpoint (for MinIO, etc.)

  - name: credentials
    type: select
    required: true
    default: env
    options: [env, config, iam]
    description: Credential source
```

**cachefs:**
```yaml
type: cachefs
schema:
  - name: size
    type: int
    required: true
    default: 1073741824
    min: 1048576
    max: 107374182400
    description: Cache size in bytes

  - name: policy
    type: select
    required: true
    default: LRU
    options: [LRU, LFU, ARC]
    description: Eviction policy

  - name: ttl
    type: int
    required: false
    default: 300
    description: Entry TTL in seconds

  - name: writeThrough
    type: bool
    required: false
    default: false
    description: Enable write-through mode
```

**encryptfs:**
```yaml
type: encryptfs
schema:
  - name: algorithm
    type: select
    required: true
    default: AES-256-GCM
    options: [AES-256-GCM, ChaCha20-Poly1305]
    description: Encryption algorithm

  - name: keySource
    type: select
    required: true
    default: env
    options: [env, file, kms]
    description: Key source

  - name: keyEnv
    type: string
    required: false
    description: Environment variable for key (if keySource=env)

  - name: keyFile
    type: string
    required: false
    description: Path to key file (if keySource=file)

  - name: encryptNames
    type: bool
    required: false
    default: false
    description: Encrypt file names
```

### API Specification

**OpenAPI 3.0:**
```yaml
openapi: 3.0.0
info:
  title: FSComposer API
  version: 1.0.0
  description: Filesystem composition engine

paths:
  /api/compositions:
    get:
      summary: List all compositions
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Composition'
    post:
      summary: Create composition
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CompositionSpec'
      responses:
        '201':
          description: Created

  /api/compositions/{id}:
    get:
      summary: Get composition
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Success

  /api/compositions/{id}/start:
    post:
      summary: Mount filesystem
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Started

components:
  schemas:
    Composition:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        status:
          type: string
          enum: [stopped, starting, running, stopping, error]
        createdAt:
          type: string
          format: date-time
        updatedAt:
          type: string
          format: date-time
```

---

## Example Use Cases

### Use Case 1: Encrypted Cloud Backup

**Scenario:** Backup local files to S3 with encryption, compression, and retry logic

**Composition:**
```yaml
version: "1.0"
name: "encrypted-backup"

nodes:
  - id: local-disk
    type: osfs
    config:
      root: /home/user/backup

  - id: compression
    type: compressfs
    config:
      algorithm: zstd
      level: 3

  - id: encryption
    type: encryptfs
    config:
      algorithm: AES-256-GCM
      keySource: env
      keyEnv: BACKUP_KEY

  - id: retry
    type: retryfs
    config:
      maxRetries: 3
      backoff: exponential

  - id: s3-backend
    type: s3fs
    config:
      bucket: my-backups
      region: us-west-2

connections:
  - from: local-disk
    to: compression
  - from: compression
    to: encryption
  - from: encryption
    to: retry
  - from: retry
    to: s3-backend

mount:
  type: fuse
  path: /mnt/backup
  root: local-disk
```

**Visual Representation:**
```
[Local Disk] → [Compress] → [Encrypt] → [Retry] → [S3]
     ↑
  [FUSE Mount: /mnt/backup]
```

---

### Use Case 2: Multi-Tier Storage with Routing

**Scenario:** Hot data in memfs, warm in local disk, cold in S3

**Composition:**
```yaml
version: "1.0"
name: "tiered-storage"

nodes:
  - id: hot-cache
    type: memfs

  - id: warm-storage
    type: osfs
    config:
      root: /mnt/ssd

  - id: cold-storage
    type: s3fs
    config:
      bucket: archive
      region: us-east-1

  - id: router
    type: switchfs
    config:
      routes:
        - pattern: "/hot/**"
          target: hot-cache
        - pattern: "/warm/**"
          target: warm-storage
        - pattern: "/cold/**"
          target: cold-storage

  - id: metrics
    type: metricsfs
    config:
      prometheus: true
      port: 9090

connections:
  - from: hot-cache
    to: router
  - from: warm-storage
    to: router
  - from: cold-storage
    to: router
  - from: router
    to: metrics

mount:
  type: webdav
  port: 8080
  root: metrics
```

**Visual Representation:**
```
      [switchfs Router]
         /    |    \
      /hot  /warm  /cold
       |      |      |
   [memfs] [osfs] [s3fs]
       \      |      /
          [metricsfs]
              ↑
         [WebDAV :8080]
```

---

### Use Case 3: Multi-User Shared Storage

**Scenario:** Team shared storage with per-user quotas and permissions

**Composition:**
```yaml
version: "1.0"
name: "team-storage"

nodes:
  - id: base-storage
    type: osfs
    config:
      root: /data/shared

  - id: permissions
    type: permfs
    config:
      rules:
        - path: "/public/**"
          allow: [read]
          users: ["*"]
        - path: "/users/alice/**"
          allow: [read, write]
          users: ["alice"]
        - path: "/admin/**"
          allow: [read, write, delete]
          users: ["admin"]

  - id: quotas
    type: quotafs
    config:
      limits:
        - user: alice
          size: 10737418240  # 10GB
        - user: bob
          size: 5368709120   # 5GB

  - id: audit
    type: logfs
    config:
      level: info
      output: /var/log/fscomposer/audit.log

  - id: metrics
    type: metricsfs
    config:
      prometheus: true

connections:
  - from: base-storage
    to: permissions
  - from: permissions
    to: quotas
  - from: quotas
    to: audit
  - from: audit
    to: metrics

mount:
  type: nfs
  export: /export/shared
  root: metrics
```

**Visual Representation:**
```
[osfs] → [permfs] → [quotafs] → [logfs] → [metricsfs]
                                               ↑
                                          [NFS Export]
```

---

## Development Roadmap

### Milestone 1: Foundation (Core Engine)
- [ ] Composition spec parser
- [ ] Node registry
- [ ] Stack builder
- [ ] FUSE mount handler
- [ ] CLI tool
- [ ] Unit tests (>80% coverage)
- [ ] 5+ example compositions

**Estimated Scope:** Core functionality

---

### Milestone 2: API & Backend
- [ ] REST API server
- [ ] WebSocket support
- [ ] Composition persistence (BoltDB)
- [ ] WebDAV mount handler
- [ ] NFS mount handler
- [ ] API documentation (OpenAPI)
- [ ] Integration tests

**Estimated Scope:** Backend services

---

### Milestone 3: Web UI
- [ ] Svelte project setup
- [ ] Canvas drag-and-drop
- [ ] Node palette
- [ ] Configuration panels
- [ ] Composition save/load
- [ ] Dark theme
- [ ] Responsive design

**Estimated Scope:** Frontend application

---

### Milestone 4: Deployment
- [ ] Docker deployment
- [ ] Kubernetes deployment
- [ ] Code generation
- [ ] Cross-compilation
- [ ] Packaging (deb, rpm)
- [ ] Deployment wizard UI

**Estimated Scope:** Production deployment

---

### Milestone 5: Extensions
- [ ] Plugin system
- [ ] Plugin SDK
- [ ] Example plugins
- [ ] Plugin marketplace
- [ ] Security scanning
- [ ] Documentation

**Estimated Scope:** Extensibility

---

### Milestone 6: Polish
- [ ] Template library
- [ ] Git integration
- [ ] Monitoring dashboard
- [ ] Testing tools
- [ ] Performance optimization
- [ ] Documentation site

**Estimated Scope:** Production-ready features

---

## Technology Stack

### Backend
- **Language:** Go 1.23+
- **Frameworks:**
  - net/http (REST API)
  - gorilla/websocket (WebSocket)
  - spf13/cobra (CLI)
  - spf13/viper (Configuration)
- **Storage:** BoltDB (composition persistence)
- **Mounting:**
  - github.com/hanwen/go-fuse/v2 (FUSE)
  - golang.org/x/net/webdav (WebDAV)
  - github.com/willscott/go-nfs (NFS)
- **Testing:**
  - testing (standard library)
  - testify (assertions)

### Frontend
- **Framework:** Svelte + SvelteKit
- **Styling:** Tailwind CSS
- **Node Editor:** Custom SVG or Rete.js
- **Build:** Vite
- **HTTP:** fetch API
- **WebSocket:** native WebSocket

### DevOps
- **Container:** Docker
- **Orchestration:** Kubernetes (optional)
- **CI/CD:** GitHub Actions
- **Packaging:** goreleaser

---

## Repository Structure

```
fscomposer/
├── README.md              # This file
├── LICENSE                # MIT License
├── go.mod                 # Go module
├── go.sum
├── .github/
│   └── workflows/
│       └── ci.yml         # CI/CD
│
├── cmd/
│   ├── fscomposer/        # CLI tool
│   │   └── main.go
│   └── fscomposesd/       # Daemon server
│       └── main.go
│
├── engine/                # Composition engine
│   ├── spec.go
│   ├── parser.go
│   ├── validator.go
│   └── builder.go
│
├── registry/              # Node registry
│   ├── registry.go
│   ├── nodes.go
│   └── schema.go
│
├── mount/                 # Mount handlers
│   ├── fuse.go
│   ├── webdav.go
│   ├── nfs.go
│   └── api.go
│
├── api/                   # REST API
│   ├── server.go
│   ├── handlers.go
│   └── websocket.go
│
├── storage/               # Persistence
│   └── compositions.go
│
├── plugin/                # Plugin system
│   ├── interface.go
│   ├── loader.go
│   └── registry.go
│
├── deploy/                # Deployment generators
│   ├── docker/
│   ├── kubernetes/
│   └── codegen/
│
├── frontend/              # Web UI
│   ├── package.json
│   ├── svelte.config.js
│   ├── vite.config.ts
│   └── src/
│       ├── routes/
│       ├── lib/
│       └── app.css
│
├── examples/              # Example compositions
│   ├── encrypted-s3.yaml
│   ├── tiered-storage.yaml
│   └── team-storage.yaml
│
├── docs/                  # Documentation
│   ├── architecture.md
│   ├── api.md
│   ├── plugins.md
│   └── deployment.md
│
└── tests/                 # Integration tests
    ├── engine_test.go
    ├── api_test.go
    └── e2e/
```

---

## Getting Started

### Prerequisites
- Go 1.23+
- Node.js 18+ (for frontend)
- Docker (optional, for deployment)
- FUSE support (libfuse on Linux, macFUSE on macOS)

### Installation

**Clone repository:**
```bash
git clone https://github.com/absfs/fscomposer.git
cd fscomposer
```

**Build CLI:**
```bash
go build -o fscomposer ./cmd/fscomposer
```

**Build server:**
```bash
go build -o fscomposesd ./cmd/fscomposesd
```

**Build frontend:**
```bash
cd frontend
npm install
npm run build
```

### Quick Start

**1. Create composition spec:**
```bash
cat > my-stack.yaml <<EOF
version: "1.0"
name: "simple-cache"
nodes:
  - id: backend
    type: osfs
    config:
      root: /tmp/data
  - id: cache
    type: cachefs
    config:
      size: 1073741824
      policy: LRU
connections:
  - from: backend
    to: cache
mount:
  type: fuse
  path: /mnt/composed
  root: cache
EOF
```

**2. Validate:**
```bash
./fscomposer validate --spec my-stack.yaml
```

**3. Mount:**
```bash
./fscomposer mount --spec my-stack.yaml
```

**4. Use filesystem:**
```bash
echo "Hello World" > /mnt/composed/test.txt
cat /mnt/composed/test.txt
```

**5. Unmount:**
```bash
umount /mnt/composed
```

---

## Contributing

Contributions welcome! Areas we need help:

1. **Core Engine:** Additional node types, validation improvements
2. **UI/UX:** Visual design, usability testing
3. **Plugins:** Community wrappers and backends
4. **Documentation:** Tutorials, examples, translations
5. **Testing:** Integration tests, performance benchmarks

### Development Workflow

1. Fork repository
2. Create feature branch: `git checkout -b feature/my-feature`
3. Make changes and test
4. Commit: `git commit -m "Add my feature"`
5. Push: `git push origin feature/my-feature`
6. Open pull request

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./tests/...

# Run frontend tests
cd frontend && npm test
```

---

## Design Notes & Decisions

### Why Hybrid Plugin System?

We chose the hybrid approach (Option 3) because:

1. **Core wrappers always available** - Fast startup, no external dependencies
2. **Extensible** - Community can add wrappers without forking
3. **Moderate binary size** - ~50MB with core wrappers
4. **Platform support** - Go plugins work on Linux/macOS (Windows via CGO)

**Trade-off:** Plugin versioning complexity, but we mitigate with:
- Semantic versioning enforcement
- Plugin compatibility matrix
- Automatic compatibility checks

### Why Svelte over React?

- **Smaller bundle size:** ~10KB vs ~40KB (React)
- **Faster runtime:** No virtual DOM
- **Better DX:** Less boilerplate
- **Compile-time optimization**

**Trade-off:** Smaller ecosystem, but we only need basic components.

### Why BoltDB for persistence?

- **Embedded:** No external database
- **Zero config:** Just a file
- **ACID transactions**
- **Proven:** Used in production (etcd, Consul)

**Trade-off:** Single-writer, but we only need simple CRUD.

### Why YAML for composition specs?

- **Human-readable:** Easy to edit by hand
- **Comments:** Can document inline
- **Git-friendly:** Diffable, mergeable

**Trade-off:** JSON also supported for programmatic generation.

---

## Related Projects

### absfs Ecosystem
- [absfs](https://github.com/absfs/absfs) - Core filesystem abstraction
- [unionfs](https://github.com/absfs/unionfs) - Multi-layer composition
- [cachefs](https://github.com/absfs/cachefs) - Caching wrapper
- [encryptfs](https://github.com/absfs/encryptfs) - Encryption wrapper
- [retryfs](https://github.com/absfs/retryfs) - Retry logic
- [metricsfs](https://github.com/absfs/metricsfs) - Observability
- [switchfs](https://github.com/absfs/switchfs) - Path routing
- [permfs](https://github.com/absfs/permfs) - Access control

### Similar Projects
- [rclone](https://rclone.org/) - CLI cloud storage sync (similar backends, no composition)
- [juicefs](https://juicefs.com/) - Distributed filesystem (monolithic, not composable)
- [s3fs-fuse](https://github.com/s3fs-fuse/s3fs-fuse) - S3 FUSE mount (single backend)
- [go-billy](https://github.com/go-git/go-billy) - VFS abstraction (no composition tools)

**Unique Value:** Visual composition + deployment + extensibility

---

## License

MIT License - see LICENSE file for details.

Copyright (c) 2024 The AbsFS Contributors

---

## Contact & Support

- **GitHub:** https://github.com/absfs/fscomposer
- **Issues:** https://github.com/absfs/fscomposer/issues
- **Discussions:** https://github.com/absfs/fscomposer/discussions

---

## Acknowledgments

This project builds on the excellent work of:
- The absfs ecosystem maintainers
- FUSE library authors
- Svelte and SvelteKit teams
- Go community

---

**Status:** 🚧 In Development - Contributions Welcome!
