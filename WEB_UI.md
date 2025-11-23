# FSComposer Web UI

Visual filesystem composition studio with an isometric 2.5D interface for building complex filesystem stacks.

## Quick Start

### 1. Start the API Server

```bash
# Build the server
go build -o fscomposer-server ./cmd/fscomposer-server

# Run the server
./fscomposer-server
```

The server will start on `http://localhost:8080`

### 2. Access the Web UI

Open your browser to: `http://localhost:8080`

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      Web Browser                         │
│  ┌─────────────────────────────────────────────────┐    │
│  │              Svelte Frontend                     │    │
│  │  - Isometric Canvas (2.5D blocks)               │    │
│  │  - Node Palette (draggable types)               │    │
│  │  - Config Panel (settings)                      │    │
│  └─────────────────────────────────────────────────┘    │
└──────────────────────┬──────────────────────────────────┘
                       │ HTTP/WebSocket
┌──────────────────────▼──────────────────────────────────┐
│                 Go API Server (:8080)                    │
│  ┌──────────────────────────────────────────────────┐   │
│  │  REST API                                        │   │
│  │  - GET  /api/nodes                               │   │
│  │  - POST /api/compositions                        │   │
│  │  - POST /api/compositions/{id}/validate          │   │
│  │  - POST /api/compositions/{id}/build             │   │
│  └──────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────┐   │
│  │  WebSocket (/api/ws)                             │   │
│  │  - Live composition updates                      │   │
│  └──────────────────────────────────────────────────┘   │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│              FSComposer Engine                           │
│  - Parser    (YAML ↔ Spec)                              │
│  - Validator (cycles, types, config)                    │
│  - Builder   (instantiate filesystem stack)             │
│  - Registry  (node type catalog)                        │
└─────────────────────────────────────────────────────────┘
```

## Features

### Visual Composition

- **Isometric 2.5D Canvas**: Drag and arrange filesystem nodes as 3D blocks
- **Color-Coded Nodes**:
  - Blue (#6366f1) for backends (osfs, memfs, s3fs)
  - Purple (#8b5cf6) for wrappers (cachefs, encryptfs, metricsfs)
- **Connection Lines**: Visual connections between nodes with glowing effects
- **Pan & Zoom**: Navigate large compositions easily

### Node Palette

- **Searchable**: Filter nodes by name or description
- **Categorized**: Grouped by backends and wrappers
- **Drag & Drop**: Drag from palette to canvas to add nodes
- **Live Loading**: Fetches available node types from API

### Configuration Panel

- **Dynamic Forms**: Auto-generated based on node schema
- **Field Types**:
  - Text inputs for strings
  - Number inputs for integers
  - Checkboxes for booleans
  - Dropdowns for select fields
- **Position Control**: Adjust X/Y coordinates
- **Save/Delete**: Update or remove nodes

### Operations

- **Save**: Store composition to server
- **Validate**: Check for cycles, type errors, config issues
- **Build & Test**: Instantiate filesystem and run basic tests

### Real-Time Updates

- WebSocket connection for live collaboration
- Instant updates when compositions change

## API Endpoints

### Nodes

```
GET /api/nodes
```

Returns all available node types with schemas.

**Response:**
```json
[
  {
    "type": "cachefs",
    "description": "Caching filesystem wrapper",
    "category": "wrapper",
    "fields": [
      {
        "name": "maxBytes",
        "type": "int",
        "default": 1073741824,
        "description": "Maximum cache size in bytes"
      }
    ]
  }
]
```

### Compositions

```
POST /api/compositions
```

Create a new composition.

**Request:**
```json
{
  "version": "1.0",
  "name": "my-fs",
  "description": "Encrypted cached filesystem",
  "nodes": [
    { "id": "storage", "type": "osfs", "config": { "root": "/tmp/data" } },
    { "id": "cache", "type": "cachefs", "config": { "maxBytes": 10485760 } }
  ],
  "connections": [
    { "from": "storage", "to": "cache" }
  ],
  "mount": {
    "type": "fuse",
    "root": "cache",
    "path": "/mnt/myfs"
  }
}
```

### Validation

```
POST /api/compositions/{id}/validate
```

Validate a composition for errors.

**Response:**
```json
{
  "valid": true
}
```

Or:

```json
{
  "valid": false,
  "error": "cycle detected: storage → cache → storage"
}
```

### Build & Test

```
POST /api/compositions/{id}/build
```

Build and test a filesystem composition.

**Response:**
```json
{
  "success": true,
  "tests": [
    "✓ Create file",
    "✓ Write data",
    "✓ Read data",
    "✓ Verify data"
  ]
}
```

## Development

### Frontend Development

```bash
cd web

# Install dependencies
npm install

# Run dev server with hot reload
npm run dev

# Build for production
npm run build
```

### Backend Development

```bash
# Run with auto-reload (using air or similar)
air -c .air.toml

# Or manual rebuild
go build -o fscomposer-server ./cmd/fscomposer-server
./fscomposer-server
```

## Design System

### Colors

- **Backgrounds**:
  - `#0a0a0a` - Darkest
  - `#121212` - Dark
  - `#1a1a1a` - Medium dark
- **Text**:
  - `#e0e0e0` - Primary
  - `#a0a0a0` - Secondary
  - `#6b7280` - Tertiary
- **Accents**:
  - `#6366f1` - Blue (backends)
  - `#8b5cf6` - Purple (wrappers)
  - `#10b981` - Green (success)
  - `#ef4444` - Red (error)
  - `#f59e0b` - Amber (warning)
- **Borders**: `#2a2a2a`

### Typography

- **Font**: Inter, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif
- **Sizes**:
  - Header: 1.5rem (24px)
  - Body: 0.875rem (14px)
  - Small: 0.75rem (12px)

## Example Compositions

### Simple Cache

```yaml
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
      maxBytes: 10485760
      policy: LRU
connections:
  - from: backend
    to: cache
mount:
  type: fuse
  root: cache
  path: /mnt/cached
```

### Encrypted Cache with Metrics

```yaml
version: "1.0"
name: "encrypted-cache-metrics"
nodes:
  - id: storage
    type: osfs
    config:
      root: /tmp/data
  - id: encrypt
    type: encryptfs
    config:
      password: "your-password"
      cipher: AES-256-GCM
  - id: cache
    type: cachefs
    config:
      maxBytes: 10485760
      policy: LRU
  - id: metrics
    type: metricsfs
connections:
  - from: storage
    to: encrypt
  - from: encrypt
    to: cache
  - from: cache
    to: metrics
mount:
  type: fuse
  root: metrics
  path: /mnt/encrypted
```

## Troubleshooting

### CORS Errors

If you see CORS errors in the browser console, ensure the API server is running and accessible at `http://localhost:8080`.

### WebSocket Connection Failed

The WebSocket connection requires the API server to be running. Check that the server started successfully.

### Node Types Not Showing

If the node palette is empty, check the browser console for API errors. The frontend fetches node types from `GET /api/nodes`.

### Build Errors

Common build errors:
- **Validation error**: Check for cycles in connections, missing required configs
- **Node type not found**: Ensure all referenced node types are registered
- **Config error**: Verify config values match field types (int vs string, etc.)

## Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

Requires support for:
- ES6 modules
- CSS Grid
- Canvas 2D
- WebSocket
- Fetch API

## Performance

- **Canvas Rendering**: Uses requestAnimationFrame for smooth 60fps
- **WebSocket**: Minimal overhead for live updates
- **Bundle Size**: ~55KB gzipped (Svelte frontend)
- **API Response**: < 10ms for most operations

## Security Notes

- **CORS**: Currently allows all origins for development
- **WebSocket**: No authentication (add auth for production)
- **Input Validation**: Server validates all composition specs
- **XSS**: Svelte auto-escapes all user content

**For production**, add:
- Authentication/authorization
- HTTPS/WSS
- Rate limiting
- CORS restrictions
- Input sanitization

## Future Enhancements

- [ ] Composition templates library
- [ ] Export to Docker Compose / Kubernetes
- [ ] Undo/redo functionality
- [ ] Collaborative editing
- [ ] Visual diff for composition changes
- [ ] Performance profiling dashboard
- [ ] Node testing in UI
- [ ] Drag connections between nodes
- [ ] Minimap for large compositions
- [ ] Dark/light theme toggle

## Contributing

See the main README.md for contribution guidelines.

## License

MIT
