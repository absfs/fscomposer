<script>
  import { composition, selectedNode, canvasState } from './stores.js';
  import { onMount } from 'svelte';

  let canvas;
  let ctx;
  let isDragging = false;
  let dragStartX = 0;
  let dragStartY = 0;
  let draggedNode = null;
  let dragOffsetX = 0;
  let dragOffsetY = 0;

  $: scale = $canvasState.scale;
  $: offsetX = $canvasState.offsetX;
  $: offsetY = $canvasState.offsetY;

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

    // Draw isometric grid
    drawGrid();

    // Draw connections
    drawConnections();

    // Draw nodes
    drawNodes();
  }

  function drawGrid() {
    const width = canvas.width;
    const height = canvas.height;
    const gridSize = 40 * scale;
    const isoAngle = Math.PI / 6; // 30 degrees

    ctx.strokeStyle = '#1a1a1a';
    ctx.lineWidth = 1;

    // Horizontal lines
    for (let y = (offsetY % gridSize) - gridSize; y < height + gridSize; y += gridSize) {
      ctx.beginPath();
      ctx.moveTo(0, y);
      ctx.lineTo(width, y);
      ctx.stroke();
    }

    // Diagonal lines (isometric)
    ctx.strokeStyle = '#151515';
    for (let x = (offsetX % gridSize) - gridSize; x < width + gridSize; x += gridSize) {
      ctx.beginPath();
      ctx.moveTo(x, 0);
      ctx.lineTo(x + height / Math.tan(isoAngle), height);
      ctx.stroke();

      ctx.beginPath();
      ctx.moveTo(x, 0);
      ctx.lineTo(x - height / Math.tan(isoAngle), height);
      ctx.stroke();
    }
  }

  function drawConnections() {
    if (!$composition.connections) return;

    $composition.connections.forEach(conn => {
      const fromNode = $composition.nodes.find(n => n.id === conn.from);
      const toNode = $composition.nodes.find(n => n.id === conn.to);

      if (!fromNode || !toNode) return;

      const fromPos = toScreenCoords(fromNode.x, fromNode.y);
      const toPos = toScreenCoords(toNode.x, toNode.y);

      // Draw glowing line
      ctx.save();

      // Outer glow
      ctx.strokeStyle = 'rgba(99, 102, 241, 0.2)';
      ctx.lineWidth = 8;
      ctx.beginPath();
      ctx.moveTo(fromPos.x + 40, fromPos.y + 30);
      ctx.lineTo(toPos.x + 40, toPos.y + 30);
      ctx.stroke();

      // Inner line
      ctx.strokeStyle = '#6366f1';
      ctx.lineWidth = 2;
      ctx.beginPath();
      ctx.moveTo(fromPos.x + 40, fromPos.y + 30);
      ctx.lineTo(toPos.x + 40, toPos.y + 30);
      ctx.stroke();

      ctx.restore();
    });
  }

  function drawNodes() {
    if (!$composition.nodes) return;

    $composition.nodes.forEach(node => {
      const pos = toScreenCoords(node.x, node.y);
      const isSelected = $selectedNode && $selectedNode.id === node.id;
      const color = node.category === 'backend' ? '#6366f1' : '#8b5cf6';

      drawIsometricBlock(pos.x, pos.y, 80, 60, 40, color, isSelected, node.name);
    });
  }

  function drawIsometricBlock(x, y, width, height, depth, color, isSelected, label) {
    ctx.save();

    // Convert to isometric coordinates
    const isoX = x + (width / 2);
    const isoY = y;

    // Top face
    ctx.fillStyle = lightenColor(color, 20);
    if (isSelected) {
      ctx.shadowColor = color;
      ctx.shadowBlur = 20;
    }
    ctx.beginPath();
    ctx.moveTo(isoX, isoY);
    ctx.lineTo(isoX + width / 2, isoY + height / 4);
    ctx.lineTo(isoX, isoY + height / 2);
    ctx.lineTo(isoX - width / 2, isoY + height / 4);
    ctx.closePath();
    ctx.fill();

    // Left face
    ctx.shadowBlur = 0;
    ctx.fillStyle = darkenColor(color, 20);
    ctx.beginPath();
    ctx.moveTo(isoX - width / 2, isoY + height / 4);
    ctx.lineTo(isoX, isoY + height / 2);
    ctx.lineTo(isoX, isoY + height / 2 + depth);
    ctx.lineTo(isoX - width / 2, isoY + height / 4 + depth);
    ctx.closePath();
    ctx.fill();

    // Right face
    ctx.fillStyle = color;
    ctx.beginPath();
    ctx.moveTo(isoX + width / 2, isoY + height / 4);
    ctx.lineTo(isoX, isoY + height / 2);
    ctx.lineTo(isoX, isoY + height / 2 + depth);
    ctx.lineTo(isoX + width / 2, isoY + height / 4 + depth);
    ctx.closePath();
    ctx.fill();

    // Label
    ctx.fillStyle = '#e0e0e0';
    ctx.font = '12px Inter, sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText(label, isoX, isoY + height / 2 + depth + 20);

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

  function toScreenCoords(worldX, worldY) {
    return {
      x: worldX * scale + offsetX,
      y: worldY * scale + offsetY
    };
  }

  function toWorldCoords(screenX, screenY) {
    return {
      x: (screenX - offsetX) / scale,
      y: (screenY - offsetY) / scale
    };
  }

  function handleMouseDown(e) {
    const rect = canvas.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;

    // Check if clicking on a node
    const clickedNode = findNodeAtPosition(mouseX, mouseY);

    if (clickedNode) {
      draggedNode = clickedNode;
      const pos = toScreenCoords(clickedNode.x, clickedNode.y);
      dragOffsetX = mouseX - pos.x;
      dragOffsetY = mouseY - pos.y;
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
      const worldPos = toWorldCoords(mouseX - dragOffsetX, mouseY - dragOffsetY);
      composition.update(c => {
        const node = c.nodes.find(n => n.id === draggedNode.id);
        if (node) {
          node.x = worldPos.x;
          node.y = worldPos.y;
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
    const worldPos = toWorldCoords(mouseX, mouseY);
    const newOffsetX = mouseX - worldPos.x * newScale;
    const newOffsetY = mouseY - worldPos.y * newScale;

    canvasState.update(s => ({
      ...s,
      scale: newScale,
      offsetX: newOffsetX,
      offsetY: newOffsetY
    }));
  }

  function findNodeAtPosition(x, y) {
    if (!$composition.nodes) return null;

    // Check in reverse order (top nodes first)
    for (let i = $composition.nodes.length - 1; i >= 0; i--) {
      const node = $composition.nodes[i];
      const pos = toScreenCoords(node.x, node.y);

      // Simple bounding box check
      if (x >= pos.x && x <= pos.x + 80 &&
          y >= pos.y && y <= pos.y + 80) {
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
    const worldPos = toWorldCoords(mouseX, mouseY);

    const nodeData = JSON.parse(nodeType);
    const newNode = {
      id: `node_${Date.now()}`,
      name: nodeData.name,
      type: nodeData.type,
      category: nodeData.category,
      x: worldPos.x,
      y: worldPos.y,
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
