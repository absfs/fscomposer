import { writable } from 'svelte/store';

// Main composition store
export const composition = writable({
  version: '1.0',
  name: 'untitled',
  description: '',
  nodes: [],
  connections: [],
  mount: {
    type: 'fuse',
    root: '',
    path: '/mnt/fscomposer'
  }
});

// Selected node for configuration
export const selectedNode = writable(null);

// Canvas state
export const canvasState = writable({
  scale: 1,
  offsetX: 0,
  offsetY: 0,
  dragging: null
});

// Node types cache
export const nodeTypes = writable([]);
