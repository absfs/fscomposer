<script>
  export let nodes = [];
  export let loading = false;

  let searchQuery = '';
  let expandedCategories = {
    backend: true,
    wrapper: true
  };

  $: filteredNodes = nodes.filter(node => {
    if (!searchQuery) return true;
    const query = searchQuery.toLowerCase();
    return (
      node.name.toLowerCase().includes(query) ||
      node.description.toLowerCase().includes(query) ||
      node.category.toLowerCase().includes(query)
    );
  });

  $: groupedNodes = filteredNodes.reduce((groups, node) => {
    const category = node.category || 'other';
    if (!groups[category]) {
      groups[category] = [];
    }
    groups[category].push(node);
    return groups;
  }, {});

  function handleDragStart(e, node) {
    e.dataTransfer.effectAllowed = 'copy';
    e.dataTransfer.setData('nodeType', JSON.stringify(node));
  }

  function toggleCategory(category) {
    expandedCategories[category] = !expandedCategories[category];
  }

  function getCategoryIcon(category) {
    switch (category) {
      case 'backend': return '💾';
      case 'wrapper': return '🔄';
      default: return '📦';
    }
  }

  function getCategoryColor(category) {
    switch (category) {
      case 'backend': return '#6366f1';
      case 'wrapper': return '#8b5cf6';
      default: return '#64748b';
    }
  }
</script>

<aside class="palette">
  <div class="header">
    <h2>Node Palette</h2>
  </div>

  <div class="search">
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
      <path
        d="M7 12C9.76142 12 12 9.76142 12 7C12 4.23858 9.76142 2 7 2C4.23858 2 2 4.23858 2 7C2 9.76142 4.23858 12 7 12Z"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
      <path
        d="M14 14L10.5 10.5"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
    </svg>
    <input
      type="text"
      placeholder="Search nodes..."
      bind:value={searchQuery}
    />
  </div>

  {#if loading}
    <div class="loading">
      <div class="spinner"></div>
      <p>Loading nodes...</p>
    </div>
  {:else if Object.keys(groupedNodes).length === 0}
    <div class="empty">
      <p>No nodes found</p>
    </div>
  {:else}
    <div class="categories">
      {#each Object.entries(groupedNodes) as [category, categoryNodes]}
        <div class="category">
          <button
            class="category-header"
            on:click={() => toggleCategory(category)}
            style="--category-color: {getCategoryColor(category)}"
          >
            <span class="icon">{getCategoryIcon(category)}</span>
            <span class="name">{category}</span>
            <span class="count">{categoryNodes.length}</span>
            <svg
              class="chevron"
              class:expanded={expandedCategories[category]}
              width="16"
              height="16"
              viewBox="0 0 16 16"
              fill="none"
            >
              <path
                d="M4 6L8 10L12 6"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </button>

          {#if expandedCategories[category]}
            <div class="nodes">
              {#each categoryNodes as node}
                <div
                  class="node"
                  draggable="true"
                  on:dragstart={(e) => handleDragStart(e, node)}
                  style="--node-color: {getCategoryColor(category)}"
                >
                  <div class="node-header">
                    <div
                      class="node-icon"
                      style="background: {getCategoryColor(category)}"
                    >
                      {getCategoryIcon(category)}
                    </div>
                    <div class="node-info">
                      <div class="node-name">{node.name}</div>
                      <div class="node-type">{node.type}</div>
                    </div>
                  </div>
                  {#if node.description}
                    <div class="node-description">{node.description}</div>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</aside>

<style>
  .palette {
    width: 280px;
    background: #0a0a0a;
    border-right: 1px solid #2a2a2a;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .header {
    padding: 1.5rem 1rem 1rem;
    border-bottom: 1px solid #1a1a1a;
  }

  h2 {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: #e0e0e0;
  }

  .search {
    padding: 1rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    border-bottom: 1px solid #1a1a1a;
  }

  .search svg {
    color: #6b7280;
    flex-shrink: 0;
  }

  .search input {
    flex: 1;
    background: transparent;
    border: none;
    color: #e0e0e0;
    font-size: 0.875rem;
    outline: none;
  }

  .search input::placeholder {
    color: #6b7280;
  }

  .loading,
  .empty {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 2rem;
    color: #6b7280;
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid #1a1a1a;
    border-top-color: #6366f1;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: 1rem;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .categories {
    flex: 1;
    overflow-y: auto;
    padding: 0.5rem 0;
  }

  .category {
    margin-bottom: 0.5rem;
  }

  .category-header {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    background: transparent;
    border: none;
    color: #e0e0e0;
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    transition: background 0.2s;
  }

  .category-header:hover {
    background: #121212;
  }

  .category-header .icon {
    font-size: 1rem;
  }

  .category-header .name {
    flex: 1;
    text-align: left;
    text-transform: capitalize;
  }

  .category-header .count {
    color: #6b7280;
    font-size: 0.75rem;
    background: #1a1a1a;
    padding: 0.125rem 0.5rem;
    border-radius: 0.75rem;
  }

  .category-header .chevron {
    color: #6b7280;
    transition: transform 0.2s;
  }

  .category-header .chevron.expanded {
    transform: rotate(180deg);
  }

  .nodes {
    padding: 0.25rem 0.5rem;
  }

  .node {
    padding: 0.75rem;
    margin-bottom: 0.5rem;
    background: #121212;
    border: 1px solid #2a2a2a;
    border-radius: 0.5rem;
    cursor: grab;
    transition: all 0.2s;
  }

  .node:hover {
    background: #1a1a1a;
    border-color: var(--node-color);
    box-shadow: 0 0 0 1px var(--node-color);
  }

  .node:active {
    cursor: grabbing;
    transform: scale(0.98);
  }

  .node-header {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
  }

  .node-icon {
    width: 32px;
    height: 32px;
    border-radius: 0.375rem;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1rem;
    flex-shrink: 0;
  }

  .node-info {
    flex: 1;
    min-width: 0;
  }

  .node-name {
    font-size: 0.875rem;
    font-weight: 500;
    color: #e0e0e0;
    margin-bottom: 0.125rem;
  }

  .node-type {
    font-size: 0.75rem;
    color: #6b7280;
    font-family: 'Monaco', 'Courier New', monospace;
  }

  .node-description {
    margin-top: 0.5rem;
    font-size: 0.75rem;
    color: #9ca3af;
    line-height: 1.4;
  }

  /* Scrollbar styling */
  .categories::-webkit-scrollbar {
    width: 6px;
  }

  .categories::-webkit-scrollbar-track {
    background: transparent;
  }

  .categories::-webkit-scrollbar-thumb {
    background: #2a2a2a;
    border-radius: 3px;
  }

  .categories::-webkit-scrollbar-thumb:hover {
    background: #3a3a3a;
  }
</style>
