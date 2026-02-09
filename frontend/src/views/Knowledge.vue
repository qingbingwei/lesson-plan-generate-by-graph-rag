<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, computed, watch } from 'vue';
import * as d3 from 'd3';
import api from '@/api/index';
import type { KnowledgeNode, KnowledgeLink, KnowledgePoint, KnowledgeNodeType, ApiResponse } from '@/types';
import { MagnifyingGlassIcon } from '@heroicons/vue/24/outline';
import { useDebounceFn } from '@/composables';
import { useDark } from '@vueuse/core';

// 移动端检测
const windowWidth = ref(window.innerWidth);
const isMobile = computed(() => windowWidth.value < 1024);
const showSidebar = ref(true);

// 搜索防抖
const debouncedSearch = useDebounceFn(() => {
  loadKnowledgeGraph();
}, 400);

// 后端返回的图谱数据结构
interface GraphApiResponse {
  nodes: Array<{
    id: string;
    label: string;
    type: string;
    subject: string;
    grade: string;
    difficulty: string;
    importance: number;
  }>;
  edges: Array<{
    source: string;
    target: string;
    type: string;
    weight: number;
  }>;
}

const svgRef = ref<SVGSVGElement | null>(null);
const loading = ref(false);
const selectedNode = ref<KnowledgePoint | null>(null);
const highlightedNodeId = ref<string | null>(null);
const isDark = useDark({
  selector: 'html',
  attribute: 'class',
  valueDark: 'dark',
  valueLight: '',
});

// 保存 zoom 实例和 svg 引用供缩放按钮使用
let currentZoom: d3.ZoomBehavior<SVGSVGElement, unknown> | null = null;
let currentSvg: d3.Selection<SVGSVGElement, unknown, null, undefined> | null = null;

// 筛选条件
const filters = ref({
  subject: '',  // 空字符串表示所有学科
  grade: '',    // 空字符串表示所有年级
  topic: '',
  limit: 50,    // 展示节点数量
});

const limitOptions = [20, 50, 100, 200, 500];

const subjects = ['全部', '语文', '数学', '英语', '物理', '化学', '生物', '历史', '地理', '政治', '科学', '信息技术', '音乐', '美术', '体育'];
const grades = [
  '全部',
  '一年级', '二年级', '三年级', '四年级', '五年级', '六年级',
  '七年级', '八年级', '九年级',
  '高一', '高二', '高三'
];

// 图谱数据
const graphData = ref<{
  nodes: KnowledgeNode[];
  links: KnowledgeLink[];
}>({
  nodes: [],
  links: [],
});

// 知识点列表 - 从图谱节点中提取
const knowledgePoints = ref<KnowledgePoint[]>([]);

// 加载知识图谱
async function loadKnowledgeGraph() {
  loading.value = true;
  
  // 清空旧数据，避免新旧数据混合显示
  graphData.value = { nodes: [], links: [] };
  knowledgePoints.value = [];
  selectedNode.value = null;
  
  // 立即清空 SVG 内容
  if (svgRef.value) {
    d3.select(svgRef.value).selectAll('*').remove();
  }
  
  try {
    const subject = filters.value.subject === '全部' ? '' : filters.value.subject;
    const grade = filters.value.grade === '全部' ? '' : filters.value.grade;
    
    const response = await api.get<ApiResponse<GraphApiResponse>>(
      '/knowledge/graph',
      { params: { subject, grade, limit: filters.value.limit } }
    );
    
    const data = response.data.data;
    const apiNodes = data?.nodes || [];
    const apiEdges = data?.edges || [];
    
    console.log('Knowledge graph API response:', data);
    console.log('Nodes count:', apiNodes.length);
    console.log('Edges count:', apiEdges.length);
    if (apiNodes.length > 0) {
      console.log('First node:', JSON.stringify(apiNodes[0]));
    }
    
    // 转换节点数据 - 确保 type 是有效的 KnowledgeNodeType
    graphData.value.nodes = apiNodes.map(node => ({
      id: node.id,
      name: node.label,
      type: node.type as KnowledgeNodeType,
      properties: {
        subject: node.subject,
        grade: node.grade,
      },
    }));
    
    // 转换边数据为 d3 需要的格式，过滤掉引用不存在节点的边
    const nodeIdSet = new Set(apiNodes.map(n => n.id));
    graphData.value.links = apiEdges
      .filter(edge => nodeIdSet.has(edge.source) && nodeIdSet.has(edge.target))
      .map(edge => ({
        source: edge.source,
        target: edge.target,
        type: edge.type,
        properties: {
          weight: edge.weight,
        },
      }));
    
    // 提取知识点列表用于侧边栏显示
    knowledgePoints.value = apiNodes.map(node => ({
      id: node.id,
      name: node.label,
      description: `${node.type} - ${node.grade || '通用'}`,
      difficulty: node.difficulty || 'medium',
      grade: node.grade || '',
      importance: Math.round((node.importance || 0.5) * 5),
      content: '',
    }));
    
    // 等待 DOM 更新后再渲染图谱
    await nextTick();
    renderGraph();
  } catch (error) {
    console.error('Failed to load knowledge graph:', error);
  } finally {
    loading.value = false;
  }
}

// 选择知识点
function selectKnowledgePoint(point: KnowledgePoint) {
  selectedNode.value = point;
  highlightedNodeId.value = point.id;
}

// 缩放控制
function zoomIn() {
  if (currentSvg && currentZoom) {
    currentSvg.transition().duration(300).call(currentZoom.scaleBy, 1.5);
  }
}
function zoomOut() {
  if (currentSvg && currentZoom) {
    currentSvg.transition().duration(300).call(currentZoom.scaleBy, 0.67);
  }
}
function zoomReset() {
  if (currentSvg && currentZoom) {
    currentSvg.transition().duration(500).call(
      currentZoom.transform, d3.zoomIdentity
    );
  }
}

function getGraphTheme() {
  if (isDark.value) {
    return {
      canvasBackground: '#0f172a',
      tooltipBackground: 'rgba(15, 23, 42, 0.95)',
      tooltipText: '#e2e8f0',
      tooltipMeta: '#94a3b8',
      linkColor: '#475569',
      linkHighlight: '#60a5fa',
      nodeStroke: '#0f172a',
      nodeHighlight: '#fbbf24',
    };
  }

  return {
    canvasBackground: '#f8fafc',
    tooltipBackground: 'rgba(15, 23, 42, 0.9)',
    tooltipText: '#ffffff',
    tooltipMeta: '#94a3b8',
    linkColor: '#cbd5e1',
    linkHighlight: '#3b82f6',
    nodeStroke: '#ffffff',
    nodeHighlight: '#fbbf24',
  };
}

// 渲染图谱
function renderGraph() {
  if (!svgRef.value || graphData.value.nodes.length === 0) return;

  const svg = d3.select(svgRef.value);
  svg.selectAll('*').remove();

  const width = Math.max(svgRef.value.clientWidth, 600);
  const height = Math.max(svgRef.value.clientHeight, 400);
  const theme = getGraphTheme();
  
  svg.attr('width', width).attr('height', height);
  svg.style('overflow', 'visible');
  svg.style('background', theme.canvasBackground);

  const g = svg.append('g');

  // Tooltip
  const tooltip = d3.select(svgRef.value.parentElement!)
    .selectAll<HTMLDivElement, unknown>('.graph-tooltip').data([0]).join('div')
    .attr('class', 'graph-tooltip')
    .style('position', 'absolute')
    .style('display', 'none')
    .style('background', theme.tooltipBackground)
    .style('color', theme.tooltipText)
    .style('padding', '8px 12px')
    .style('border-radius', '6px')
    .style('font-size', '12px')
    .style('pointer-events', 'none')
    .style('z-index', '50')
    .style('max-width', '220px')
    .style('box-shadow', '0 4px 12px rgba(0,0,0,0.3)');

  // 缩放 + 平移
  const zoom = d3.zoom<SVGSVGElement, unknown>()
    .scaleExtent([0.05, 6])
    .on('zoom', (event) => {
      g.attr('transform', event.transform);
    });

  svg.call(zoom);
  currentZoom = zoom;
  currentSvg = svg;

  svg.style('cursor', 'grab');
  svg.on('mousedown.cursor', () => svg.style('cursor', 'grabbing'))
     .on('mouseup.cursor', () => svg.style('cursor', 'grab'));

  // 点击空白取消高亮
  svg.on('click', (event) => {
    if (event.target === svgRef.value) {
      highlightedNodeId.value = null;
      resetHighlight();
    }
  });

  // 深拷贝节点和边数据
  const nodes = graphData.value.nodes.map(n => ({ ...n }));
  const links = graphData.value.links.map(l => ({ ...l }));

  // 构建邻接表用于高亮
  const adjacencyMap = new Map<string, Set<string>>();
  links.forEach(l => {
    const src = typeof l.source === 'string' ? l.source : (l.source as any).id;
    const tgt = typeof l.target === 'string' ? l.target : (l.target as any).id;
    if (!adjacencyMap.has(src)) adjacencyMap.set(src, new Set());
    if (!adjacencyMap.has(tgt)) adjacencyMap.set(tgt, new Set());
    adjacencyMap.get(src)!.add(tgt);
    adjacencyMap.get(tgt)!.add(src);
  });

  // 根据节点数量动态调整力导向参数
  const nodeCount = nodes.length;
  const chargeStrength = nodeCount > 200 ? -50 : nodeCount > 80 ? -100 : nodeCount > 30 ? -200 : -300;
  const linkDistance = nodeCount > 200 ? 30 : nodeCount > 80 ? 50 : nodeCount > 30 ? 70 : 100;
  const collisionRadius = nodeCount > 200 ? 12 : nodeCount > 80 ? 18 : nodeCount > 30 ? 25 : 35;
  const nodeRadius = nodeCount > 200 ? 8 : nodeCount > 80 ? 12 : nodeCount > 30 ? 18 : 25;
  const fontSize = nodeCount > 200 ? '6px' : nodeCount > 80 ? '8px' : nodeCount > 30 ? '10px' : '12px';
  const labelLength = nodeCount > 80 ? 4 : 6;

  // 力导向图
  const simulation = d3.forceSimulation(nodes as d3.SimulationNodeDatum[])
    .force('link', d3.forceLink(links)
      .id((d: unknown) => (d as KnowledgeNode).id)
      .distance(linkDistance))
    .force('charge', d3.forceManyBody().strength(chargeStrength))
    .force('center', d3.forceCenter(width / 2, height / 2))
    .force('collision', d3.forceCollide().radius(collisionRadius))
    .force('x', d3.forceX(width / 2).strength(0.05))
    .force('y', d3.forceY(height / 2).strength(0.05));

  // 连线
  const link = g.append('g')
    .selectAll('line')
    .data(links)
    .join('line')
    .attr('stroke', theme.linkColor)
    .attr('stroke-width', 2)
    .attr('stroke-opacity', 0.6);
  
  // 节点
  const node = g.append('g')
    .selectAll('g')
    .data(nodes)
    .join('g')
    .style('cursor', 'pointer')
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    .call(d3.drag<SVGGElement, KnowledgeNode>()
      .on('start', (event, d) => {
        if (!event.active) simulation.alphaTarget(0.3).restart();
        (d as d3.SimulationNodeDatum).fx = (d as d3.SimulationNodeDatum).x;
        (d as d3.SimulationNodeDatum).fy = (d as d3.SimulationNodeDatum).y;
      })
      .on('drag', (event, d) => {
        (d as d3.SimulationNodeDatum).fx = event.x;
        (d as d3.SimulationNodeDatum).fy = event.y;
      })
      .on('end', (event, d) => {
        if (!event.active) simulation.alphaTarget(0);
        (d as d3.SimulationNodeDatum).fx = null;
        (d as d3.SimulationNodeDatum).fy = null;
      }) as any);

  // 节点圆形
  node.append('circle')
    .attr('r', nodeRadius)
    .attr('fill', (d) => getNodeColor(d.type))
    .attr('stroke', theme.nodeStroke)
    .attr('stroke-width', nodeRadius > 15 ? 2 : 1);

  // 节点文字
  node.append('text')
    .text((d) => d.name.length > labelLength ? d.name.substring(0, labelLength) + '…' : d.name)
    .attr('text-anchor', 'middle')
    .attr('dy', '0.35em')
    .attr('fill', '#fff')
    .attr('font-size', fontSize)
    .attr('font-weight', 'bold');

  // ---- 交互：Tooltip ----
  node.on('mouseenter', (_event, d) => {
    const typeName = getNodeTypeName(d.type);
    tooltip
      .style('display', 'block')
      .html(`<div class="font-semibold">${d.name}</div>
             <div style="color:${theme.tooltipMeta};margin-top:2px">${typeName}</div>
             ${d.properties?.subject ? `<div style="color:${theme.tooltipMeta}">${d.properties.subject} ${d.properties.grade || ''}</div>` : ''}`);
  })
  .on('mousemove', (event) => {
    const rect = svgRef.value!.parentElement!.getBoundingClientRect();
    tooltip
      .style('left', `${event.clientX - rect.left + 12}px`)
      .style('top', `${event.clientY - rect.top - 10}px`);
  })
  .on('mouseleave', () => {
    tooltip.style('display', 'none');
  });

  // ---- 交互：点击高亮 ----
  node.on('click', (event, d) => {
    event.stopPropagation();
    const clickedId = d.id;
    highlightedNodeId.value = clickedId;
    
    // 选中侧边栏
    const point = knowledgePoints.value.find(p => p.id === clickedId);
    if (point) selectedNode.value = point;

    const neighbors = adjacencyMap.get(clickedId) || new Set();

    // 节点高亮
    node.select('circle')
      .attr('opacity', (n) => {
        if (n.id === clickedId) return 1;
        if (neighbors.has(n.id)) return 1;
        return 0.15;
      })
      .attr('stroke', (n) => n.id === clickedId ? theme.nodeHighlight : theme.nodeStroke)
      .attr('stroke-width', (n) => n.id === clickedId ? 3 : (nodeRadius > 15 ? 2 : 1));

    node.select('text')
      .attr('opacity', (n) => {
        if (n.id === clickedId) return 1;
        if (neighbors.has(n.id)) return 1;
        return 0.15;
      });

    // 边高亮
    link
      .attr('stroke-opacity', (l) => {
        const src = typeof l.source === 'string' ? l.source : (l.source as any).id;
        const tgt = typeof l.target === 'string' ? l.target : (l.target as any).id;
        return (src === clickedId || tgt === clickedId) ? 0.8 : 0.05;
      })
      .attr('stroke', (l) => {
        const src = typeof l.source === 'string' ? l.source : (l.source as any).id;
        const tgt = typeof l.target === 'string' ? l.target : (l.target as any).id;
        return (src === clickedId || tgt === clickedId) ? theme.linkHighlight : theme.linkColor;
      })
      .attr('stroke-width', (l) => {
        const src = typeof l.source === 'string' ? l.source : (l.source as any).id;
        const tgt = typeof l.target === 'string' ? l.target : (l.target as any).id;
        return (src === clickedId || tgt === clickedId) ? 3 : 2;
      });
  });

  // 取消高亮
  function resetHighlight() {
    node.select('circle')
      .attr('opacity', 1)
      .attr('stroke', theme.nodeStroke)
      .attr('stroke-width', nodeRadius > 15 ? 2 : 1);
    node.select('text').attr('opacity', 1);
    link
      .attr('stroke-opacity', 0.6)
      .attr('stroke', theme.linkColor)
      .attr('stroke-width', 2);
  }

  // 自动缩放函数
  function fitToView(animated = true) {
    let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
    nodes.forEach(n => {
      const nd = n as d3.SimulationNodeDatum;
      if (nd.x != null && nd.x < minX) minX = nd.x;
      if (nd.y != null && nd.y < minY) minY = nd.y;
      if (nd.x != null && nd.x > maxX) maxX = nd.x;
      if (nd.y != null && nd.y > maxY) maxY = nd.y;
    });
    if (!isFinite(minX)) return;

    const graphWidth = maxX - minX || 1;
    const graphHeight = maxY - minY || 1;
    const padding = 80;
    const scale = Math.min(
      (width - padding * 2) / graphWidth,
      (height - padding * 2) / graphHeight,
      2
    );
    const centerX = (minX + maxX) / 2;
    const centerY = (minY + maxY) / 2;

    const transform = d3.zoomIdentity
      .translate(width / 2, height / 2)
      .scale(scale)
      .translate(-centerX, -centerY);

    if (animated) {
      svg.transition().duration(500).call(zoom.transform, transform);
    } else {
      svg.call(zoom.transform, transform);
    }
  }

  // 更新位置
  let tickCount = 0;
  simulation.on('tick', () => {
    tickCount++;
    link
      .attr('x1', (d) => (d.source as d3.SimulationNodeDatum).x!)
      .attr('y1', (d) => (d.source as d3.SimulationNodeDatum).y!)
      .attr('x2', (d) => (d.target as d3.SimulationNodeDatum).x!)
      .attr('y2', (d) => (d.target as d3.SimulationNodeDatum).y!);

    node.attr('transform', (d) => 
      `translate(${(d as d3.SimulationNodeDatum).x},${(d as d3.SimulationNodeDatum).y})`
    );

    if (tickCount === 3) {
      fitToView(false);
    }
  });

  simulation.on('end', () => {
    fitToView(true);
  });
}

// 获取节点颜色
function getNodeColor(type: string): string {
  const colors: Record<string, string> = {
    Subject: '#3b82f6',
    Chapter: '#8b5cf6',
    KnowledgePoint: '#10b981',
    Skill: '#f59e0b',
    Concept: '#06b6d4',
    Principle: '#ec4899',
    Formula: '#f97316',
    Example: '#64748b',
    Resource: '#ec4899',
    Lesson: '#6366f1',
  };
  return colors[type] || '#6b7280';
}

// 获取节点类型中文名
function getNodeTypeName(type: string): string {
  const names: Record<string, string> = {
    Subject: '学科',
    Chapter: '章节',
    KnowledgePoint: '知识点',
    Skill: '技能',
    Concept: '概念',
    Principle: '原理',
    Formula: '公式',
    Example: '示例',
    Resource: '资源',
    Lesson: '课程',
  };
  return names[type] || type || '其他';
}

// 处理窗口大小变化
function handleResize() {
  windowWidth.value = window.innerWidth;
  renderGraph();
}

onMounted(() => {
  loadKnowledgeGraph();
  window.addEventListener('resize', handleResize);
});

watch(isDark, () => {
  renderGraph();
});

onUnmounted(() => {
  window.removeEventListener('resize', handleResize);
});
</script>

<template>
  <div class="knowledge-page flex flex-col lg:flex-row gap-4 lg:gap-6" style="height: calc(100vh - 10rem);">
    <!-- 左侧面板 - 移动端可折叠 -->
    <div class="lg:w-80 flex-shrink-0 flex flex-col space-y-4" :class="{ 'hidden': !showSidebar && isMobile }">
      <!-- 筛选 -->
      <div class="card">
        <div class="card-body space-y-3">
          <div class="grid grid-cols-2 lg:grid-cols-1 gap-3">
            <div>
              <label class="label">学科</label>
              <select v-model="filters.subject" class="select" @change="loadKnowledgeGraph">
                <option v-for="s in subjects" :key="s" :value="s">{{ s }}</option>
              </select>
            </div>
            <div>
              <label class="label">年级</label>
              <select v-model="filters.grade" class="select" @change="loadKnowledgeGraph">
                <option v-for="g in grades" :key="g" :value="g">{{ g }}</option>
              </select>
            </div>
          </div>
          <div class="grid grid-cols-2 lg:grid-cols-1 gap-3">
            <div>
              <label class="label">节点数量</label>
              <select v-model.number="filters.limit" class="select" @change="loadKnowledgeGraph">
                <option v-for="n in limitOptions" :key="n" :value="n">{{ n }} 个</option>
              </select>
            </div>
            <div>
              <label class="label">关键词</label>
              <div class="relative">
                <MagnifyingGlassIcon class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400 dark:text-gray-500" />
                <input
                  v-model="filters.topic"
                  type="text"
                  class="input pl-9"
                  placeholder="搜索..."
                  @input="debouncedSearch"
                  @keyup.enter="loadKnowledgeGraph"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 知识点列表 - 移动端最多显示200px高度 -->
      <div class="card flex-1 overflow-hidden flex flex-col max-h-48 lg:max-h-none">
        <div class="card-header">
          <h3 class="font-medium">知识点</h3>
        </div>
        <div class="flex-1 overflow-auto scrollbar-thin">
          <div v-if="loading" class="p-4 text-center">
            <div class="loading" />
          </div>
          <ul v-else class="divide-y divide-gray-100 dark:divide-gray-700">
            <li
              v-for="point in knowledgePoints"
              :key="point.id"
              class="p-3 hover:bg-gray-50 dark:hover:bg-gray-700/40 cursor-pointer transition-colors"
              :class="{ 'bg-primary-50 dark:bg-primary-900/30': selectedNode?.id === point.id }"
              @click="selectKnowledgePoint(point)"
            >
              <h4 class="font-medium text-sm text-gray-900 dark:text-gray-100">{{ point.name }}</h4>
              <p class="text-xs text-gray-500 dark:text-gray-400 mt-1 line-clamp-2">{{ point.description }}</p>
              <div class="flex gap-2 mt-2">
                <span class="badge-secondary text-xs">{{ point.difficulty }}</span>
                <span class="badge-secondary text-xs">重要度: {{ point.importance }}</span>
              </div>
            </li>
          </ul>
        </div>
      </div>
    </div>

    <!-- 右侧图谱 -->
    <div class="flex-1 card overflow-hidden flex flex-col min-h-[300px] lg:min-h-0">
      <div class="card-header flex items-center justify-between flex-shrink-0 gap-2">
        <div class="flex items-center gap-2">
          <!-- 移动端侧边栏切换按钮 -->
          <button
            type="button"
            class="lg:hidden btn-icon p-1"
            @click="showSidebar = !showSidebar"
            :title="showSidebar ? '隐藏筛选' : '显示筛选'"
          >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="h-5 w-5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M10.5 6h9.75M10.5 6a1.5 1.5 0 11-3 0m3 0a1.5 1.5 0 10-3 0M3.75 6H7.5m3 12h9.75m-9.75 0a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m-3.75 0H7.5m9-6h3.75m-3.75 0a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m-9.75 0h9.75" />
            </svg>
          </button>
          <h3 class="font-medium">知识图谱</h3>
        </div>
        <div class="flex items-center gap-2 lg:gap-3">
          <!-- 图例 - 移动端隐藏文字 -->
          <div class="graph-legend hidden sm:flex gap-2 flex-wrap">
            <span class="flex items-center gap-1 text-xs">
              <span class="w-2.5 h-2.5 rounded-full bg-blue-500"></span>学科
            </span>
            <span class="flex items-center gap-1 text-xs">
              <span class="w-2.5 h-2.5 rounded-full bg-purple-500"></span>章节
            </span>
            <span class="flex items-center gap-1 text-xs">
              <span class="w-2.5 h-2.5 rounded-full bg-emerald-500"></span>知识点
            </span>
            <span class="flex items-center gap-1 text-xs">
              <span class="w-2.5 h-2.5 rounded-full bg-yellow-500"></span>技能
            </span>
            <span class="flex items-center gap-1 text-xs">
              <span class="w-2.5 h-2.5 rounded-full bg-cyan-500"></span>概念
            </span>
            <span class="flex items-center gap-1 text-xs">
              <span class="w-2.5 h-2.5 rounded-full bg-pink-500"></span>原理
            </span>
            <span class="flex items-center gap-1 text-xs">
              <span class="w-2.5 h-2.5 rounded-full" style="background:#64748b"></span>示例
            </span>
            <span class="flex items-center gap-1 text-xs">
              <span class="w-2.5 h-2.5 rounded-full bg-orange-500"></span>公式
            </span>
            <span class="flex items-center gap-1 text-xs">
              <span class="w-2.5 h-2.5 rounded-full bg-indigo-500"></span>课程
            </span>
            <span class="flex items-center gap-1 text-xs">
              <span class="w-2.5 h-2.5 rounded-full bg-gray-400"></span>其他
            </span>
          </div>
          <!-- 缩放控制 -->
          <div class="graph-zoom-controls flex items-center border border-gray-200 dark:border-gray-600 rounded-md divide-x divide-gray-200 dark:divide-gray-600">
            <button type="button" class="px-2 py-1 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700/80" title="放大" @click="zoomIn">+</button>
            <button type="button" class="px-2 py-1 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700/80" title="重置" @click="zoomReset">⟳</button>
            <button type="button" class="px-2 py-1 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700/80" title="缩小" @click="zoomOut">−</button>
          </div>
        </div>
      </div>
      <div class="flex-1 relative min-h-0">
        <svg
          ref="svgRef"
          class="graph-canvas w-full h-full absolute inset-0"
          :class="{ 'opacity-50': loading }"
          style="min-height: 300px;"
        />
        <div
          v-if="graphData.nodes.length === 0 && !loading"
          class="absolute inset-0 flex items-center justify-center"
        >
          <div class="graph-empty text-center text-gray-500 dark:text-gray-400">
            <p>暂无知识图谱数据</p>
            <p class="text-sm mt-1">请先上传文档或调整筛选条件</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.graph-canvas {
  background: #f8fafc;
}

:global(.dark) .graph-canvas {
  background: #0f172a;
}

.graph-legend span {
  color: #4b5563;
}

:global(.dark) .graph-legend span {
  color: #d1d5db;
}

.graph-zoom-controls {
  border-color: #e5e7eb;
  background: rgba(255, 255, 255, 0.85);
  color: #374151;
}

.graph-zoom-controls :deep(button) {
  color: inherit;
}

.graph-zoom-controls :deep(button) {
  transition: background-color 0.2s;
}

.graph-zoom-controls :deep(button:hover) {
  background: #f3f4f6;
}

:global(.dark) .graph-zoom-controls {
  border-color: #4b5563;
  background: rgba(31, 41, 55, 0.85);
  color: #e5e7eb;
}

:global(.dark) .graph-zoom-controls :deep(button:hover) {
  background: #374151;
}

:global(.dark) .graph-empty {
  color: #9ca3af;
}
</style>
