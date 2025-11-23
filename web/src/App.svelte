<script>
  import Canvas from './lib/Canvas.svelte';
  import NodePalette from './lib/NodePalette.svelte';
  import ConfigPanel from './lib/ConfigPanel.svelte';
  import { composition, selectedNode } from './lib/stores.js';
  import { onMount } from 'svelte';

  let ws;
  let nodes = [];
  let loading = true;

  onMount(async () => {
    // Load available node types
    const res = await fetch('http://localhost:8080/api/nodes');
    nodes = await res.json();
    loading = false;

    // Connect WebSocket
    ws = new WebSocket('ws://localhost:8080/api/ws');
    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      console.log('WebSocket message:', msg);
    };
  });

  async function saveComposition() {
    const res = await fetch('http://localhost:8080/api/compositions', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify($composition)
    });

    if (res.ok) {
      alert('Composition saved!');
    }
  }

  async function validateComposition() {
    const id = $composition.name;
    const res = await fetch(`http://localhost:8080/api/compositions/${id}/validate`, {
      method: 'POST'
    });

    const result = await res.json();
    if (result.valid) {
      alert('Composition is valid!');
    } else {
      alert(`Validation error: ${result.error}`);
    }
  }

  async function buildComposition() {
    const id = $composition.name;
    const res = await fetch(`http://localhost:8080/api/compositions/${id}/build`, {
      method: 'POST'
    });

    const result = await res.json();
    if (result.success) {
      alert(`Build successful!\n\nTests:\n${result.tests.join('\n')}`);
    } else {
      alert(`Build failed: ${result.error}`);
    }
  }
</script>

<div class="app">
  <header>
    <h1>FS Composer</h1>
    <div class="toolbar">
      <button on:click={saveComposition}>Save</button>
      <button on:click={validateComposition}>Validate</button>
      <button on:click={buildComposition}>Build & Test</button>
    </div>
  </header>

  <div class="workspace">
    <NodePalette {nodes} {loading} />
    <Canvas />
    {#if $selectedNode}
      <ConfigPanel />
    {/if}
  </div>
</div>

<style>
  :global(body) {
    margin: 0;
    background: #0a0a0a;
    color: #e0e0e0;
    font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    overflow: hidden;
  }

  .app {
    height: 100vh;
    display: flex;
    flex-direction: column;
  }

  header {
    background: #000;
    border-bottom: 1px solid #1a1a1a;
    padding: 1rem 2rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  h1 {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 600;
    background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .toolbar {
    display: flex;
    gap: 0.75rem;
  }

  button {
    background: #1a1a1a;
    border: 1px solid #2a2a2a;
    color: #e0e0e0;
    padding: 0.5rem 1.25rem;
    border-radius: 0.375rem;
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    transition: all 0.2s;
  }

  button:hover {
    background: #2a2a2a;
    border-color: #3a3a3a;
  }

  button:active {
    transform: scale(0.98);
  }

  .workspace {
    flex: 1;
    display: flex;
    position: relative;
    overflow: hidden;
  }
</style>
