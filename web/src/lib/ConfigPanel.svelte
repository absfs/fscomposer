<script>
  import { composition, selectedNode } from './stores.js';

  let config = {};
  let fields = [];

  $: if ($selectedNode) {
    config = { ...$selectedNode.config };
    // Get fields from node type
    loadNodeFields($selectedNode.type);
  }

  async function loadNodeFields(nodeType) {
    try {
      const res = await fetch(`http://localhost:8080/api/nodes/${nodeType}/fields`);
      if (res.ok) {
        fields = await res.json();
      } else {
        // Fallback: generate basic fields from config
        fields = Object.keys(config).map(key => ({
          name: key,
          type: typeof config[key] === 'boolean' ? 'bool' : 'string',
          description: '',
          required: false
        }));
      }
    } catch (error) {
      console.error('Failed to load fields:', error);
      fields = [];
    }
  }

  function handleSave() {
    composition.update(c => {
      const node = c.nodes.find(n => n.id === $selectedNode.id);
      if (node) {
        node.config = { ...config };
        // Update selectedNode to trigger reactivity
        selectedNode.update(n => ({ ...n, config: { ...config } }));
      }
      return c;
    });
  }

  function handleDelete() {
    if (!confirm(`Delete node "${$selectedNode.name}"?`)) return;

    composition.update(c => ({
      ...c,
      nodes: c.nodes.filter(n => n.id !== $selectedNode.id),
      connections: c.connections.filter(
        conn => conn.from !== $selectedNode.id && conn.to !== $selectedNode.id
      )
    }));

    selectedNode.set(null);
  }

  function handleClose() {
    selectedNode.set(null);
  }

  function getFieldComponent(field) {
    switch (field.type) {
      case 'bool':
        return 'checkbox';
      case 'int':
      case 'number':
        return 'number';
      case 'select':
      case 'enum':
        return 'select';
      default:
        return 'text';
    }
  }
</script>

{#if $selectedNode}
  <aside class="config-panel">
    <div class="header">
      <div class="title-section">
        <h2>Configure Node</h2>
        <button class="close-btn" on:click={handleClose} aria-label="Close">
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
            <path
              d="M15 5L5 15M5 5L15 15"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
            />
          </svg>
        </button>
      </div>
    </div>

    <div class="content">
      <div class="node-info">
        <div class="node-badge" style="background: {$selectedNode.category === 'backend' ? '#6366f1' : '#8b5cf6'}">
          {$selectedNode.category === 'backend' ? '💾' : '🔄'}
        </div>
        <div class="node-details">
          <div class="node-name">{$selectedNode.name}</div>
          <div class="node-type">{$selectedNode.type}</div>
        </div>
      </div>

      <div class="section">
        <h3>Basic Information</h3>
        <div class="form-group">
          <label for="node-id">Node ID</label>
          <input
            type="text"
            id="node-id"
            value={$selectedNode.id}
            disabled
            class="disabled"
          />
        </div>

        <div class="form-group">
          <label for="node-name">Display Name</label>
          <input
            type="text"
            id="node-name"
            bind:value={$selectedNode.name}
            on:change={handleSave}
          />
        </div>
      </div>

      {#if fields.length > 0}
        <div class="section">
          <h3>Configuration</h3>
          {#each fields as field}
            <div class="form-group">
              <label for={field.name}>
                {field.name}
                {#if field.required}
                  <span class="required">*</span>
                {/if}
              </label>
              {#if field.description}
                <p class="field-description">{field.description}</p>
              {/if}

              {#if getFieldComponent(field) === 'checkbox'}
                <label class="checkbox-label">
                  <input
                    type="checkbox"
                    id={field.name}
                    bind:checked={config[field.name]}
                  />
                  <span class="checkbox-text">
                    {field.label || 'Enable'}
                  </span>
                </label>
              {:else if getFieldComponent(field) === 'select'}
                <select
                  id={field.name}
                  bind:value={config[field.name]}
                >
                  <option value="">Select...</option>
                  {#each field.options || [] as option}
                    <option value={option.value || option}>
                      {option.label || option}
                    </option>
                  {/each}
                </select>
              {:else if getFieldComponent(field) === 'number'}
                <input
                  type="number"
                  id={field.name}
                  bind:value={config[field.name]}
                  min={field.min}
                  max={field.max}
                  step={field.step || 1}
                />
              {:else}
                <input
                  type="text"
                  id={field.name}
                  bind:value={config[field.name]}
                  placeholder={field.placeholder || ''}
                />
              {/if}
            </div>
          {/each}
        </div>
      {/if}

      <div class="section">
        <h3>Position</h3>
        <div class="form-row">
          <div class="form-group">
            <label for="pos-x">X</label>
            <input
              type="number"
              id="pos-x"
              bind:value={$selectedNode.x}
              on:change={handleSave}
            />
          </div>
          <div class="form-group">
            <label for="pos-y">Y</label>
            <input
              type="number"
              id="pos-y"
              bind:value={$selectedNode.y}
              on:change={handleSave}
            />
          </div>
        </div>
      </div>
    </div>

    <div class="footer">
      <button class="btn btn-primary" on:click={handleSave}>
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
          <path
            d="M13.5 4L6 11.5L2.5 8"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
        Save Changes
      </button>
      <button class="btn btn-danger" on:click={handleDelete}>
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
          <path
            d="M2 4H14M6 4V2.5C6 2.22386 6.22386 2 6.5 2H9.5C9.77614 2 10 2.22386 10 2.5V4M12.5 4V13.5C12.5 13.7761 12.2761 14 12 14H4C3.72386 14 3.5 13.7761 3.5 13.5V4"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
        Delete Node
      </button>
    </div>
  </aside>
{/if}

<style>
  .config-panel {
    width: 320px;
    background: #0a0a0a;
    border-left: 1px solid #2a2a2a;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .header {
    padding: 1.5rem 1rem;
    border-bottom: 1px solid #1a1a1a;
  }

  .title-section {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  h2 {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: #e0e0e0;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: #6b7280;
    padding: 0.25rem;
    cursor: pointer;
    border-radius: 0.25rem;
    transition: all 0.2s;
  }

  .close-btn:hover {
    background: #1a1a1a;
    color: #e0e0e0;
  }

  .content {
    flex: 1;
    overflow-y: auto;
    padding: 1rem;
  }

  .node-info {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1rem;
    background: #121212;
    border: 1px solid #2a2a2a;
    border-radius: 0.5rem;
    margin-bottom: 1.5rem;
  }

  .node-badge {
    width: 48px;
    height: 48px;
    border-radius: 0.5rem;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.5rem;
    flex-shrink: 0;
  }

  .node-details {
    flex: 1;
    min-width: 0;
  }

  .node-name {
    font-size: 0.875rem;
    font-weight: 600;
    color: #e0e0e0;
    margin-bottom: 0.25rem;
  }

  .node-type {
    font-size: 0.75rem;
    color: #6b7280;
    font-family: 'Monaco', 'Courier New', monospace;
  }

  .section {
    margin-bottom: 1.5rem;
  }

  h3 {
    margin: 0 0 1rem 0;
    font-size: 0.875rem;
    font-weight: 600;
    color: #9ca3af;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .form-group {
    margin-bottom: 1rem;
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.75rem;
  }

  label {
    display: block;
    font-size: 0.875rem;
    font-weight: 500;
    color: #e0e0e0;
    margin-bottom: 0.5rem;
  }

  .required {
    color: #ef4444;
  }

  .field-description {
    margin: -0.25rem 0 0.5rem 0;
    font-size: 0.75rem;
    color: #6b7280;
    line-height: 1.4;
  }

  input[type="text"],
  input[type="number"],
  select {
    width: 100%;
    background: #121212;
    border: 1px solid #2a2a2a;
    color: #e0e0e0;
    padding: 0.5rem 0.75rem;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    transition: all 0.2s;
  }

  input[type="text"]:focus,
  input[type="number"]:focus,
  select:focus {
    outline: none;
    border-color: #6366f1;
    box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.1);
  }

  input[type="text"].disabled,
  input[type="number"].disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  input::placeholder {
    color: #6b7280;
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    padding: 0.5rem;
    border-radius: 0.375rem;
    transition: background 0.2s;
  }

  .checkbox-label:hover {
    background: #121212;
  }

  input[type="checkbox"] {
    width: 18px;
    height: 18px;
    cursor: pointer;
    accent-color: #6366f1;
  }

  .checkbox-text {
    color: #e0e0e0;
    font-size: 0.875rem;
  }

  select {
    cursor: pointer;
  }

  .footer {
    padding: 1rem;
    border-top: 1px solid #1a1a1a;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.625rem 1rem;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    border: 1px solid;
    transition: all 0.2s;
  }

  .btn svg {
    flex-shrink: 0;
  }

  .btn-primary {
    background: #6366f1;
    border-color: #6366f1;
    color: #fff;
  }

  .btn-primary:hover {
    background: #5558e3;
    border-color: #5558e3;
  }

  .btn-primary:active {
    transform: scale(0.98);
  }

  .btn-danger {
    background: transparent;
    border-color: #2a2a2a;
    color: #ef4444;
  }

  .btn-danger:hover {
    background: rgba(239, 68, 68, 0.1);
    border-color: #ef4444;
  }

  .btn-danger:active {
    transform: scale(0.98);
  }

  /* Scrollbar styling */
  .content::-webkit-scrollbar {
    width: 6px;
  }

  .content::-webkit-scrollbar-track {
    background: transparent;
  }

  .content::-webkit-scrollbar-thumb {
    background: #2a2a2a;
    border-radius: 3px;
  }

  .content::-webkit-scrollbar-thumb:hover {
    background: #3a3a3a;
  }
</style>
