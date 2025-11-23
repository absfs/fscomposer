<script>
  import { composition, selectedNode, canvasState } from './stores.js';
  import { onMount } from 'svelte';

  let canvas;
  let ctx;
  let isDragging = false;
  let dragStartX = 0;
  let dragStartY = 0;
  let draggedNode = null;
  let dragOffsetGridX = 0;
  let dragOffsetGridY = 0;

  $: scale = $canvasState.scale;
  $: offsetX = $canvasState.offsetX;
  $: offsetY = $canvasState.offsetY;

  // Isometric grid constants
  const TILE_WIDTH = 60;
  const TILE_HEIGHT = 30;
  const CUBE_SIZE = 60;

  onMount(() => {
    ctx = canvas.getContext('2d');
    resizeCanvas();
    window.addEventListener('resize', resizeCanvas);
    draw();

    return () => {
      window.removeEventListener('resize', resizeCanvas);
    };
  });

  function resizeCanvas() {
    if (!canvas) return;
    canvas.width = canvas.offsetWidth;
    canvas.height = canvas.offsetHeight;
    draw();
  }

  $: if (ctx && ($composition || $canvasState)) {
    draw();
  }

  // Convert grid coordinates to screen coordinates
  function gridToScreen(gridX, gridY) {
    const screenX = (gridX - gridY) * TILE_WIDTH / 2;
    const screenY = (gridX + gridY) * TILE_HEIGHT / 2;
    return {
      x: screenX * scale + canvas.width / 2 + offsetX,
      y: screenY * scale + 100 + offsetY
    };
  }

  // Convert screen coordinates to grid coordinates
  function screenToGrid(screenX, screenY) {
    const centerX = canvas.width / 2 + offsetX;
    const centerY = 100 + offsetY;
    const x = (screenX - centerX) / scale;
    const y = (screenY - centerY) / scale;
    const gridX = Math.round((x / TILE_WIDTH + y / TILE_HEIGHT));
    const gridY = Math.round((y / TILE_HEIGHT - x / TILE_WIDTH));
    return { gridX, gridY };
  }

  function draw() {
    if (!ctx) return;

    const width = canvas.width;
    const height = canvas.height;

    // Background gradient
    const gradient = ctx.createLinearGradient(0, 0, 0, height);
    gradient.addColorStop(0, '#0a0a0a');
    gradient.addColorStop(1, '#121212');
    ctx.fillStyle = gradient;
    ctx.fillRect(0, 0, width, height);

    // Draw isometric grid covering entire canvas
    drawIsometricGrid();

    // Draw connections
    drawConnections();

    // Draw nodes sorted by depth
    drawNodes();
  }

  function drawIsometricGrid() {
    const width = canvas.width;
    const height = canvas.height;

    ctx.strokeStyle = '#1a1a1a';
    ctx.lineWidth = 1;

    // Calculate grid bounds to cover entire canvas
    const centerX = width / 2 + offsetX;
    const centerY = 100 + offsetY;

    // Determine how many grid lines we need
    const gridRange = Math.max(width, height) / (TILE_WIDTH * scale) + 20;

    ctx.save();

    // Draw diagonal lines forming isometric diamond pattern
    for (let i = -gridRange; i <= gridRange; i++) {
      // Lines going down-right (30 degrees)
      ctx.beginPath();
      const x1 = centerX + i * TILE_WIDTH / 2 * scale;
      const y1 = centerY + i * TILE_HEIGHT / 2 * scale;
      const x2 = centerX + (i - gridRange * 2) * TILE_WIDTH / 2 * scale;
      const y2 = centerY + (i + gridRange * 2) * TILE_HEIGHT / 2 * scale;
      ctx.moveTo(x1, y1);
      ctx.lineTo(x2, y2);
      ctx.stroke();

      // Lines going down-left (-30 degrees)
      ctx.beginPath();
      const x3 = centerX - i * TILE_WIDTH / 2 * scale;
      const y3 = centerY + i * TILE_HEIGHT / 2 * scale;
      const x4 = centerX - (i - gridRange * 2) * TILE_WIDTH / 2 * scale;
      const y4 = centerY + (i + gridRange * 2) * TILE_HEIGHT / 2 * scale;
      ctx.moveTo(x3, y3);
      ctx.lineTo(x4, y4);
      ctx.stroke();
    }

    ctx.restore();
  }

  function drawConnections() {
    if (!$composition.connections) return;

    $composition.connections.forEach(conn => {
      const fromNode = $composition.nodes.find(n => n.id === conn.from);
      const toNode = $composition.nodes.find(n => n.id === conn.to);

      if (!fromNode || !toNode) return;

      const fromPos = gridToScreen(fromNode.gridX || 0, fromNode.gridY || 0);
      const toPos = gridToScreen(toNode.gridX || 0, toNode.gridY || 0);

      // Calculate cube centers
      const fromCenterX = fromPos.x;
      const fromCenterY = fromPos.y + TILE_HEIGHT * scale;
      const toCenterX = toPos.x;
      const toCenterY = toPos.y + TILE_HEIGHT * scale;

      // Draw glowing line
      ctx.save();

      // Outer glow
      ctx.strokeStyle = 'rgba(99, 102, 241, 0.2)';
      ctx.lineWidth = 8;
      ctx.beginPath();
      ctx.moveTo(fromCenterX, fromCenterY);
      ctx.lineTo(toCenterX, toCenterY);
      ctx.stroke();

      // Inner line
      ctx.strokeStyle = '#6366f1';
      ctx.lineWidth = 2;
      ctx.beginPath();
      ctx.moveTo(fromCenterX, fromCenterY);
      ctx.lineTo(toCenterX, toCenterY);
      ctx.stroke();

      ctx.restore();
    });
  }

  function drawNodes() {
    if (!$composition.nodes) return;

    // Sort nodes by depth (back to front)
    const sortedNodes = [...$composition.nodes].sort((a, b) => {
      const depthA = (a.gridY || 0) + (a.gridX || 0);
      const depthB = (b.gridY || 0) + (b.gridX || 0);
      return depthA - depthB;
    });

    sortedNodes.forEach(node => {
      const gridX = node.gridX || 0;
      const gridY = node.gridY || 0;
      const pos = gridToScreen(gridX, gridY);
      const isSelected = $selectedNode && $selectedNode.id === node.id;
      const color = node.category === 'backend' ? '#6366f1' : '#8b5cf6';

      drawIsometricCube(pos.x, pos.y, CUBE_SIZE * scale, color, isSelected, node.name);
    });
  }

  function drawIsometricCube(x, y, size, color, isSelected, label) {
    ctx.save();

    const halfSize = size / 2;
    const quarterSize = size / 4;

    // Add glow if selected
    if (isSelected) {
      ctx.shadowColor = color;
      ctx.shadowBlur = 20;
    }

    // Top face (lightest)
    ctx.fillStyle = lightenColor(color, 30);
    ctx.beginPath();
    ctx.moveTo(x, y);
    ctx.lineTo(x + halfSize, y + quarterSize);
    ctx.lineTo(x, y + halfSize);
    ctx.lineTo(x - halfSize, y + quarterSize);
    ctx.closePath();
    ctx.fill();

    // Remove shadow for other faces
    ctx.shadowBlur = 0;

    // Left face (medium dark)
    ctx.fillStyle = darkenColor(color, 30);
    ctx.beginPath();
    ctx.moveTo(x - halfSize, y + quarterSize);
    ctx.lineTo(x, y + halfSize);
    ctx.lineTo(x, y + size);
    ctx.lineTo(x - halfSize, y + size - quarterSize);
    ctx.closePath();
    ctx.fill();

    // Right face (darkest)
    ctx.fillStyle = darkenColor(color, 10);
    ctx.beginPath();
    ctx.moveTo(x + halfSize, y + quarterSize);
    ctx.lineTo(x, y + halfSize);
    ctx.lineTo(x, y + size);
    ctx.lineTo(x + halfSize, y + size - quarterSize);
    ctx.closePath();
    ctx.fill();

    // Draw edges for better definition
    ctx.strokeStyle = 'rgba(0, 0, 0, 0.3)';
    ctx.lineWidth = 1;

    // Top edges
    ctx.beginPath();
    ctx.moveTo(x, y);
    ctx.lineTo(x + halfSize, y + quarterSize);
    ctx.lineTo(x, y + halfSize);
    ctx.lineTo(x - halfSize, y + quarterSize);
    ctx.closePath();
    ctx.stroke();

    // Label
    ctx.fillStyle = '#e0e0e0';
    ctx.font = `${Math.max(10, 12 * scale)}px Inter, sans-serif`;
    ctx.textAlign = 'center';
    ctx.fillText(label, x, y + size + 20 * scale);

    ctx.restore();
  }

  function lightenColor(color, percent) {
    const num = parseInt(color.replace('#', ''), 16);
    const amt = Math.round(2.55 * percent);
    const R = Math.min(255, (num >> 16) + amt);
    const G = Math.min(255, ((num >> 8) & 0x00FF) + amt);
    const B = Math.min(255, (num & 0x0000FF) + amt);
    return `rgb(${R}, ${G}, ${B})`;
  }

  function darkenColor(color, percent) {
    const num = parseInt(color.replace('#', ''), 16);
    const amt = Math.round(2.55 * percent);
    const R = Math.max(0, (num >> 16) - amt);
    const G = Math.max(0, ((num >> 8) & 0x00FF) - amt);
    const B = Math.max(0, (num & 0x0000FF) - amt);
    return `rgb(${R}, ${G}, ${B})`;
  }

  function handleMouseDown(e) {
    const rect = canvas.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;

    // Check if clicking on a node
    const clickedNode = findNodeAtPosition(mouseX, mouseY);

    if (clickedNode) {
      draggedNode = clickedNode;
      dragOffsetGridX = clickedNode.gridX || 0;
      dragOffsetGridY = clickedNode.gridY || 0;
      selectedNode.set(clickedNode);
    } else {
      isDragging = true;
      dragStartX = mouseX - offsetX;
      dragStartY = mouseY - offsetY;
      selectedNode.set(null);
    }
  }

  function handleMouseMove(e) {
    const rect = canvas.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;

    if (draggedNode) {
      // Snap to grid when dragging
      const gridPos = screenToGrid(mouseX, mouseY);
      composition.update(c => {
        const node = c.nodes.find(n => n.id === draggedNode.id);
        if (node) {
          node.gridX = gridPos.gridX;
          node.gridY = gridPos.gridY;
          // Keep old x/y for backward compatibility
          const screenPos = gridToScreen(gridPos.gridX, gridPos.gridY);
          node.x = screenPos.x;
          node.y = screenPos.y;
        }
        return c;
      });
      canvas.style.cursor = 'grabbing';
    } else if (isDragging) {
      canvasState.update(s => ({
        ...s,
        offsetX: mouseX - dragStartX,
        offsetY: mouseY - dragStartY
      }));
      canvas.style.cursor = 'grabbing';
    } else {
      // Check if hovering over node
      const hoveredNode = findNodeAtPosition(mouseX, mouseY);
      canvas.style.cursor = hoveredNode ? 'pointer' : 'default';
    }
  }

  function handleMouseUp() {
    isDragging = false;
    draggedNode = null;
    canvas.style.cursor = 'default';
  }

  function handleWheel(e) {
    e.preventDefault();
    const rect = canvas.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;

    const delta = e.deltaY > 0 ? 0.9 : 1.1;
    const newScale = Math.max(0.5, Math.min(2, scale * delta));

    // Zoom towards mouse position
    const worldPosOld = {
      x: (mouseX - offsetX - canvas.width / 2) / scale,
      y: (mouseY - offsetY - 100) / scale
    };
    const newOffsetX = mouseX - canvas.width / 2 - worldPosOld.x * newScale;
    const newOffsetY = mouseY - 100 - worldPosOld.y * newScale;

    canvasState.update(s => ({
      ...s,
      scale: newScale,
      offsetX: newOffsetX,
      offsetY: newOffsetY
    }));
  }

  function findNodeAtPosition(x, y) {
    if (!$composition.nodes) return null;

    // Check in reverse order (front nodes first based on depth)
    const sortedNodes = [...$composition.nodes].sort((a, b) => {
      const depthA = (a.gridY || 0) + (a.gridX || 0);
      const depthB = (b.gridY || 0) + (b.gridX || 0);
      return depthB - depthA; // Reverse for front to back
    });

    for (const node of sortedNodes) {
      const gridX = node.gridX || 0;
      const gridY = node.gridY || 0;
      const pos = gridToScreen(gridX, gridY);
      const size = CUBE_SIZE * scale;
      const halfSize = size / 2;

      // Isometric hit detection (diamond shape for top)
      const dx = x - pos.x;
      const dy = y - pos.y;

      // Check if point is inside the isometric cube's bounding area
      if (Math.abs(dx) / halfSize + Math.abs(dy - halfSize / 2) / (halfSize / 2) <= 1) {
        return node;
      }

      // Also check the vertical faces
      if (dy > 0 && dy < size && Math.abs(dx) < halfSize) {
        return node;
      }
    }
    return null;
  }

  // Handle drop from palette
  function handleDrop(e) {
    e.preventDefault();
    const nodeType = e.dataTransfer.getData('nodeType');
    if (!nodeType) return;

    const rect = canvas.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;

    // Snap to grid
    const gridPos = screenToGrid(mouseX, mouseY);
    const screenPos = gridToScreen(gridPos.gridX, gridPos.gridY);

    const nodeData = JSON.parse(nodeType);

    // Auto-layout based on connections
    let finalGridX = gridPos.gridX;
    let finalGridY = gridPos.gridY;

    // If this is a wrapper (has inputs), position inputs above it
    if (nodeData.category === 'wrapper') {
      // This will be the receiver at bottom-center
      finalGridX = gridPos.gridX;
      finalGridY = gridPos.gridY;
    }

    const newNode = {
      id: `node_${Date.now()}`,
      name: nodeData.name,
      type: nodeData.type,
      category: nodeData.category,
      gridX: finalGridX,
      gridY: finalGridY,
      x: screenPos.x, // Keep for backward compatibility
      y: screenPos.y,
      config: {}
    };

    composition.update(c => ({
      ...c,
      nodes: [...c.nodes, newNode]
    }));
  }

  function handleDragOver(e) {
    e.preventDefault();
  }
</script>

<canvas
  bind:this={canvas}
  on:mousedown={handleMouseDown}
  on:mousemove={handleMouseMove}
  on:mouseup={handleMouseUp}
  on:mouseleave={handleMouseUp}
  on:wheel={handleWheel}
  on:drop={handleDrop}
  on:dragover={handleDragOver}
/>

<style>
  canvas {
    flex: 1;
    cursor: default;
  }
</style>
