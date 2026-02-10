<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import * as d3 from 'd3';
import { useDark } from '@vueuse/core';
import { ElMessage } from 'element-plus';
import {
  Search,
  Operation,
  Plus,
  Minus,
  RefreshRight,
  ArrowLeftBold,
  ArrowRightBold,
} from '@element-plus/icons-vue';
import api from '@/api/index';
import { useDebounceFn } from '@/composables';
import type { ApiResponse, KnowledgeLink, KnowledgeNode, KnowledgeNodeType, KnowledgePoint } from '@/types';

type GraphScope = 'matched' | 'one_hop' | 'two_hop';

interface GraphApiResponse {
  nodes: Array<{
    id: string;
    label: string;
    type: string;
    subject: string;
    grade: string;
    difficulty: string;
    importance: number;
    description?: string;
    keywords?: string[];
  }>;
  edges: Array<{
    source: string;
    target: string;
    type: string;
    weight: number;
  }>;
}

const svgRef = ref<SVGSVGElement | null>(null);
const windowWidth = ref(window.innerWidth);
const isMobile = computed(() => windowWidth.value < 1024);
const showSidebar = ref(true);
const loading = ref(false);
const selectedNode = ref<KnowledgePoint | null>(null);
const highlightedNodeId = ref<string | null>(null);

const isDark = useDark({
  selector: 'html',
  attribute: 'class',
  valueDark: 'dark',
  valueLight: '',
});

const filters = ref({
  subject: '全部',
  grade: '全部',
  topic: '',
  scope: 'one_hop' as GraphScope,
  limit: 50,
});

const limitOptions = [20, 50, 100, 200, 500];
const subjects = ['全部', '语文', '数学', '英语', '物理', '化学', '生物', '历史', '地理', '政治', '科学', '信息技术', '音乐', '美术', '体育'];
const grades = ['全部', '一年级', '二年级', '三年级', '四年级', '五年级', '六年级', '七年级', '八年级', '九年级', '高一', '高二', '高三'];

const scopeOptions = [
  { label: '仅命中', value: 'matched' },
  { label: '命中 + 1跳', value: 'one_hop' },
  { label: '命中 + 2跳', value: 'two_hop' },
] as const;

const graphData = ref<{
  nodes: KnowledgeNode[];
  links: KnowledgeLink[];
}>({
  nodes: [],
  links: [],
});

const knowledgePoints = ref<KnowledgePoint[]>([]);
const debouncedSearch = useDebounceFn(() => {
  loadKnowledgeGraph();
}, 400);

let currentZoom: d3.ZoomBehavior<SVGSVGElement, unknown> | null = null;
let currentSvg: d3.Selection<SVGSVGElement, unknown, null, undefined> | null = null;
let graphRequestSeq = 0;

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

function selectKnowledgePoint(point: KnowledgePoint) {
  selectedNode.value = point;
  highlightedNodeId.value = point.id;
}

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
    currentSvg.transition().duration(500).call(currentZoom.transform, d3.zoomIdentity);
  }
}


function matchesKeyword(node: GraphApiResponse['nodes'][number], keyword: string): boolean {
  const normalizedKeyword = keyword.trim().toLowerCase();
  if (!normalizedKeyword) {
    return true;
  }

  const keywordTexts = [
    node.label,
    node.description || '',
    ...(node.keywords || []).map((item) => String(item)),
  ];

  return keywordTexts.some((text) => text.toLowerCase().includes(normalizedKeyword));
}

function filterGraphByTopicScope(
  nodes: GraphApiResponse['nodes'],
  edges: GraphApiResponse['edges'],
  topic: string,
  scope: GraphScope,
): { nodes: GraphApiResponse['nodes']; edges: GraphApiResponse['edges'] } {
  const normalizedTopic = topic.trim().toLowerCase();
  if (!normalizedTopic) {
    return { nodes, edges };
  }

  const matchedIds = new Set(
    nodes
      .filter((node) => matchesKeyword(node, normalizedTopic))
      .map((node) => node.id)
  );

  if (matchedIds.size === 0) {
    return { nodes: [], edges: [] };
  }

  const maxDepth = scope === 'matched' ? 0 : scope === 'two_hop' ? 2 : 1;
  const keptIds = new Set(matchedIds);

  if (maxDepth > 0) {
    const adjacency = new Map<string, Set<string>>();
    edges.forEach((edge) => {
      if (!adjacency.has(edge.source)) adjacency.set(edge.source, new Set());
      if (!adjacency.has(edge.target)) adjacency.set(edge.target, new Set());
      adjacency.get(edge.source)!.add(edge.target);
      adjacency.get(edge.target)!.add(edge.source);
    });

    const queue: Array<{ id: string; depth: number }> = Array.from(matchedIds).map((id) => ({ id, depth: 0 }));

    while (queue.length > 0) {
      const current = queue.shift();
      if (!current || current.depth >= maxDepth) {
        continue;
      }

      const neighbors = adjacency.get(current.id);
      if (!neighbors) {
        continue;
      }

      neighbors.forEach((neighborId) => {
        if (keptIds.has(neighborId)) {
          return;
        }

        keptIds.add(neighborId);
        queue.push({ id: neighborId, depth: current.depth + 1 });
      });
    }
  }

  return {
    nodes: nodes.filter((node) => keptIds.has(node.id)),
    edges: edges.filter((edge) => keptIds.has(edge.source) && keptIds.has(edge.target)),
  };
}

async function loadKnowledgeGraph() {
  const requestSeq = ++graphRequestSeq;
  loading.value = true;
  graphData.value = { nodes: [], links: [] };
  knowledgePoints.value = [];
  selectedNode.value = null;

  if (svgRef.value) {
    d3.select(svgRef.value).selectAll('*').remove();
  }

  try {
    const subject = filters.value.subject === '全部' ? '' : filters.value.subject;
    const grade = filters.value.grade === '全部' ? '' : filters.value.grade;
    const topic = filters.value.topic.trim();

    const response = await api.get<ApiResponse<GraphApiResponse>>('/knowledge/graph', {
      params: {
        subject,
        grade,
        topic: topic || undefined,
        scope: topic ? filters.value.scope : undefined,
        limit: filters.value.limit,
      },
    });

    if (requestSeq !== graphRequestSeq) {
      return;
    }

    const data = response.data.data;
    const apiNodes = data?.nodes || [];
    const apiEdges = data?.edges || [];
    const filteredGraph = filterGraphByTopicScope(apiNodes, apiEdges, topic, filters.value.scope);

    graphData.value.nodes = filteredGraph.nodes.map((node) => ({
      id: node.id,
      name: node.label,
      type: node.type as KnowledgeNodeType,
      properties: {
        subject: node.subject,
        grade: node.grade,
      },
    }));

    const nodeIdSet = new Set(filteredGraph.nodes.map((node) => node.id));
    graphData.value.links = filteredGraph.edges
      .filter((edge) => nodeIdSet.has(edge.source) && nodeIdSet.has(edge.target))
      .map((edge) => ({
        source: edge.source,
        target: edge.target,
        type: edge.type,
        properties: {
          weight: edge.weight,
        },
      }));

    knowledgePoints.value = filteredGraph.nodes.map((node) => ({
      id: node.id,
      name: node.label,
      description: `${getNodeTypeName(node.type)} - ${node.grade || '通用'}`,
      difficulty: node.difficulty || 'medium',
      grade: node.grade || '',
      importance: Math.round((node.importance || 0.5) * 5),
      content: '',
    }));

    await nextTick();
    renderGraph();
  } catch (error) {
    if (requestSeq === graphRequestSeq) {
      console.error('Failed to load knowledge graph:', error);
      ElMessage.error('知识图谱加载失败，请稍后重试');
    }
  } finally {
    if (requestSeq === graphRequestSeq) {
      loading.value = false;
    }
  }
}

function renderGraph() {
  if (!svgRef.value || graphData.value.nodes.length === 0) {
    return;
  }

  const svg = d3.select(svgRef.value);
  svg.selectAll('*').remove();

  const width = Math.max(svgRef.value.clientWidth, 600);
  const height = Math.max(svgRef.value.clientHeight, 420);
  const theme = getGraphTheme();

  svg.attr('width', width).attr('height', height);
  svg.style('overflow', 'visible');
  svg.style('background', theme.canvasBackground);

  const g = svg.append('g');

  const tooltip = d3
    .select(svgRef.value.parentElement!)
    .selectAll<HTMLDivElement, unknown>('.graph-tooltip')
    .data([0])
    .join('div')
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

  const zoom = d3
    .zoom<SVGSVGElement, unknown>()
    .scaleExtent([0.05, 6])
    .on('zoom', (event) => {
      g.attr('transform', event.transform);
    });

  svg.call(zoom);
  currentZoom = zoom;
  currentSvg = svg;

  svg.style('cursor', 'grab');
  svg.on('mousedown.cursor', () => svg.style('cursor', 'grabbing')).on('mouseup.cursor', () => svg.style('cursor', 'grab'));

  svg.on('click', (event) => {
    if (event.target === svgRef.value) {
      highlightedNodeId.value = null;
      resetHighlight();
    }
  });

  const nodes = graphData.value.nodes.map((node) => ({ ...node }));
  const links = graphData.value.links.map((link) => ({ ...link }));

  const adjacencyMap = new Map<string, Set<string>>();
  links.forEach((link) => {
    const src = typeof link.source === 'string' ? link.source : (link.source as any).id;
    const tgt = typeof link.target === 'string' ? link.target : (link.target as any).id;
    if (!adjacencyMap.has(src)) adjacencyMap.set(src, new Set());
    if (!adjacencyMap.has(tgt)) adjacencyMap.set(tgt, new Set());
    adjacencyMap.get(src)!.add(tgt);
    adjacencyMap.get(tgt)!.add(src);
  });

  const nodeCount = nodes.length;
  const chargeStrength = nodeCount > 200 ? -50 : nodeCount > 80 ? -100 : nodeCount > 30 ? -200 : -300;
  const linkDistance = nodeCount > 200 ? 30 : nodeCount > 80 ? 50 : nodeCount > 30 ? 70 : 100;
  const collisionRadius = nodeCount > 200 ? 12 : nodeCount > 80 ? 18 : nodeCount > 30 ? 25 : 35;
  const nodeRadius = nodeCount > 200 ? 8 : nodeCount > 80 ? 12 : nodeCount > 30 ? 18 : 25;
  const fontSize = nodeCount > 200 ? '6px' : nodeCount > 80 ? '8px' : nodeCount > 30 ? '10px' : '12px';
  const labelLength = nodeCount > 80 ? 4 : 6;

  const simulation = d3
    .forceSimulation(nodes as d3.SimulationNodeDatum[])
    .force(
      'link',
      d3
        .forceLink(links)
        .id((d: unknown) => (d as KnowledgeNode).id)
        .distance(linkDistance)
    )
    .force('charge', d3.forceManyBody().strength(chargeStrength))
    .force('center', d3.forceCenter(width / 2, height / 2))
    .force('collision', d3.forceCollide().radius(collisionRadius))
    .force('x', d3.forceX(width / 2).strength(0.05))
    .force('y', d3.forceY(height / 2).strength(0.05));

  const link = g
    .append('g')
    .selectAll('line')
    .data(links)
    .join('line')
    .attr('stroke', theme.linkColor)
    .attr('stroke-width', 2)
    .attr('stroke-opacity', 0.6);

  const node = g
    .append('g')
    .selectAll('g')
    .data(nodes)
    .join('g')
    .style('cursor', 'pointer')
    .call(
      d3
        .drag<SVGGElement, KnowledgeNode>()
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
        }) as any
    );

  node
    .append('circle')
    .attr('r', nodeRadius)
    .attr('fill', (d) => getNodeColor(d.type))
    .attr('stroke', theme.nodeStroke)
    .attr('stroke-width', nodeRadius > 15 ? 2 : 1);

  node
    .append('text')
    .text((d) => (d.name.length > labelLength ? `${d.name.substring(0, labelLength)}…` : d.name))
    .attr('text-anchor', 'middle')
    .attr('dy', '0.35em')
    .attr('fill', '#fff')
    .attr('font-size', fontSize)
    .attr('font-weight', 'bold');

  node
    .on('mouseenter', (_event, d) => {
      const typeName = getNodeTypeName(d.type);
      tooltip
        .style('display', 'block')
        .html(`<div class="font-semibold">${d.name}</div><div style="color:${theme.tooltipMeta};margin-top:2px">${typeName}</div>${d.properties?.subject ? `<div style="color:${theme.tooltipMeta}">${d.properties.subject} ${(d.properties?.grade as string) || ''}</div>` : ''}`);
    })
    .on('mousemove', (event) => {
      const rect = svgRef.value!.parentElement!.getBoundingClientRect();
      tooltip.style('left', `${event.clientX - rect.left + 12}px`).style('top', `${event.clientY - rect.top - 10}px`);
    })
    .on('mouseleave', () => {
      tooltip.style('display', 'none');
    });

  node.on('click', (event, d) => {
    event.stopPropagation();
    const clickedId = d.id;
    highlightedNodeId.value = clickedId;

    const point = knowledgePoints.value.find((item) => item.id === clickedId);
    if (point) {
      selectedNode.value = point;
    }

    const neighbors = adjacencyMap.get(clickedId) || new Set();

    node
      .select('circle')
      .attr('opacity', (n) => {
        if (n.id === clickedId) return 1;
        if (neighbors.has(n.id)) return 1;
        return 0.15;
      })
      .attr('stroke', (n) => (n.id === clickedId ? theme.nodeHighlight : theme.nodeStroke))
      .attr('stroke-width', (n) => (n.id === clickedId ? 3 : nodeRadius > 15 ? 2 : 1));

    node.select('text').attr('opacity', (n) => {
      if (n.id === clickedId) return 1;
      if (neighbors.has(n.id)) return 1;
      return 0.15;
    });

    link
      .attr('stroke-opacity', (item) => {
        const src = typeof item.source === 'string' ? item.source : (item.source as any).id;
        const tgt = typeof item.target === 'string' ? item.target : (item.target as any).id;
        return src === clickedId || tgt === clickedId ? 0.8 : 0.05;
      })
      .attr('stroke', (item) => {
        const src = typeof item.source === 'string' ? item.source : (item.source as any).id;
        const tgt = typeof item.target === 'string' ? item.target : (item.target as any).id;
        return src === clickedId || tgt === clickedId ? theme.linkHighlight : theme.linkColor;
      })
      .attr('stroke-width', (item) => {
        const src = typeof item.source === 'string' ? item.source : (item.source as any).id;
        const tgt = typeof item.target === 'string' ? item.target : (item.target as any).id;
        return src === clickedId || tgt === clickedId ? 3 : 2;
      });
  });

  function resetHighlight() {
    node
      .select('circle')
      .attr('opacity', 1)
      .attr('stroke', theme.nodeStroke)
      .attr('stroke-width', nodeRadius > 15 ? 2 : 1);

    node.select('text').attr('opacity', 1);

    link.attr('stroke-opacity', 0.6).attr('stroke', theme.linkColor).attr('stroke-width', 2);
  }

  function fitToView(animated = true) {
    let minX = Infinity;
    let minY = Infinity;
    let maxX = -Infinity;
    let maxY = -Infinity;

    nodes.forEach((node) => {
      const item = node as d3.SimulationNodeDatum;
      if (item.x != null && item.x < minX) minX = item.x;
      if (item.y != null && item.y < minY) minY = item.y;
      if (item.x != null && item.x > maxX) maxX = item.x;
      if (item.y != null && item.y > maxY) maxY = item.y;
    });

    if (!isFinite(minX)) {
      return;
    }

    const graphWidth = maxX - minX || 1;
    const graphHeight = maxY - minY || 1;
    const padding = 80;
    const scale = Math.min((width - padding * 2) / graphWidth, (height - padding * 2) / graphHeight, 2);
    const centerX = (minX + maxX) / 2;
    const centerY = (minY + maxY) / 2;

    const transform = d3.zoomIdentity.translate(width / 2, height / 2).scale(scale).translate(-centerX, -centerY);

    if (animated) {
      svg.transition().duration(500).call(zoom.transform, transform);
    } else {
      svg.call(zoom.transform, transform);
    }
  }

  let tickCount = 0;
  simulation.on('tick', () => {
    tickCount += 1;

    link
      .attr('x1', (d) => (d.source as d3.SimulationNodeDatum).x!)
      .attr('y1', (d) => (d.source as d3.SimulationNodeDatum).y!)
      .attr('x2', (d) => (d.target as d3.SimulationNodeDatum).x!)
      .attr('y2', (d) => (d.target as d3.SimulationNodeDatum).y!);

    node.attr('transform', (d) => `translate(${(d as d3.SimulationNodeDatum).x},${(d as d3.SimulationNodeDatum).y})`);

    if (tickCount === 3) {
      fitToView(false);
    }
  });

  simulation.on('end', () => {
    fitToView(true);
  });
}

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
  <div class="knowledge-page flex flex-col lg:flex-row gap-4 lg:gap-6" style="height: calc(100vh - 10rem)">
    <div class="lg:w-84 flex-shrink-0 flex flex-col gap-4" :class="{ hidden: !showSidebar && isMobile }">
      <el-card class="surface-card" shadow="never">
        <template #header>
          <div class="font-semibold">筛选条件</div>
        </template>

        <el-form label-position="top" class="space-y-2">
          <el-row :gutter="12">
            <el-col :xs="12" :lg="24">
              <el-form-item label="学科">
                <el-select v-model="filters.subject" @change="loadKnowledgeGraph">
                  <el-option v-for="subject in subjects" :key="subject" :label="subject" :value="subject" />
                </el-select>
              </el-form-item>
            </el-col>

            <el-col :xs="12" :lg="24">
              <el-form-item label="年级">
                <el-select v-model="filters.grade" @change="loadKnowledgeGraph">
                  <el-option v-for="grade in grades" :key="grade" :label="grade" :value="grade" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="12">
            <el-col :xs="12" :lg="24">
              <el-form-item label="节点数量">
                <el-select v-model="filters.limit" @change="loadKnowledgeGraph">
                  <el-option v-for="limit in limitOptions" :key="limit" :label="`${limit} 个`" :value="limit" />
                </el-select>
              </el-form-item>
            </el-col>

            <el-col :xs="12" :lg="24">
              <el-form-item label="关键词">
                <el-input
                  v-model="filters.topic"
                  clearable
                  placeholder="输入关键词搜索"
                  :prefix-icon="Search"
                  @input="debouncedSearch"
                  @clear="loadKnowledgeGraph"
                  @keyup.enter="loadKnowledgeGraph"
                />
              </el-form-item>
            </el-col>

            <el-col :xs="24" :lg="24">
              <el-form-item label="展示范围">
                <el-radio-group
                  v-model="filters.scope"
                  size="small"
                  :disabled="!filters.topic.trim()"
                  @change="loadKnowledgeGraph"
                >
                  <el-radio-button
                    v-for="scope in scopeOptions"
                    :key="scope.value"
                    :label="scope.value"
                  >
                    {{ scope.label }}
                  </el-radio-button>
                </el-radio-group>
              </el-form-item>
            </el-col>
          </el-row>

          <el-button type="primary" plain :icon="Operation" @click="loadKnowledgeGraph">刷新图谱</el-button>
        </el-form>
      </el-card>

      <el-card class="surface-card flex-1 min-h-0" shadow="never">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="font-semibold">知识点列表</span>
            <el-tag size="small" effect="plain">{{ knowledgePoints.length }} 个</el-tag>
          </div>
        </template>

        <el-skeleton v-if="loading" :rows="6" animated />

        <el-empty v-else-if="knowledgePoints.length === 0" description="暂无知识点数据" />

        <el-scrollbar v-else class="knowledge-list-scroll">
          <div class="space-y-2 pr-1">
            <el-card
              v-for="point in knowledgePoints"
              :key="point.id"
              class="knowledge-point-card cursor-pointer"
              shadow="hover"
              :class="{ 'is-active': selectedNode?.id === point.id }"
              @click="selectKnowledgePoint(point)"
            >
              <div class="font-medium app-text-primary line-clamp-1">{{ point.name }}</div>
              <div class="text-xs app-text-muted mt-1 line-clamp-2">{{ point.description }}</div>
              <div class="flex items-center gap-2 mt-2">
                <el-tag size="small" type="info" effect="plain">{{ point.difficulty }}</el-tag>
                <el-tag size="small" type="success" effect="plain">重要度 {{ point.importance }}</el-tag>
              </div>
            </el-card>
          </div>
        </el-scrollbar>
      </el-card>
    </div>

    <el-card class="surface-card flex-1 min-h-[360px]" shadow="never">
      <template #header>
        <div class="flex items-center justify-between gap-2 flex-wrap">
          <div class="flex items-center gap-2">
            <el-button
              v-if="isMobile"
              circle
              :icon="showSidebar ? ArrowLeftBold : ArrowRightBold"
              @click="showSidebar = !showSidebar"
            />
            <span class="font-semibold">知识图谱</span>
          </div>

          <div class="flex items-center gap-2 flex-wrap">
            <div class="graph-legend hidden md:flex items-center gap-2 flex-wrap">
              <span class="legend-item"><span class="legend-dot" style="background:#3b82f6" />学科</span>
              <span class="legend-item"><span class="legend-dot" style="background:#8b5cf6" />章节</span>
              <span class="legend-item"><span class="legend-dot" style="background:#10b981" />知识点</span>
              <span class="legend-item"><span class="legend-dot" style="background:#f59e0b" />技能</span>
              <span class="legend-item"><span class="legend-dot" style="background:#06b6d4" />概念</span>
              <span class="legend-item"><span class="legend-dot" style="background:#ec4899" />原理</span>
            </div>

            <el-button-group>
              <el-button :icon="Plus" @click="zoomIn" />
              <el-button :icon="RefreshRight" @click="zoomReset" />
              <el-button :icon="Minus" @click="zoomOut" />
            </el-button-group>
          </div>
        </div>
      </template>

      <div class="graph-wrapper relative h-full min-h-[420px]">
        <svg ref="svgRef" class="graph-canvas w-full h-full absolute inset-0" :class="{ 'opacity-50': loading }" />

        <div v-if="graphData.nodes.length === 0 && !loading" class="absolute inset-0 flex items-center justify-center">
          <el-empty description="暂无知识图谱数据，请先上传文档或调整筛选条件" />
        </div>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.knowledge-list-scroll {
  height: calc(100vh - 24rem);
  min-height: 220px;
}

.knowledge-point-card :deep(.el-card__body) {
  padding: 12px;
}

.knowledge-point-card.is-active {
  border-color: var(--el-color-primary);
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--el-color-primary) 45%, transparent);
}

.graph-wrapper {
  border-radius: 10px;
  overflow: hidden;
}

.graph-canvas {
  background: #f8fafc;
}

:global(.dark) .graph-canvas {
  background: #0f172a;
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #64748b;
}

:global(.dark) .legend-item {
  color: #cbd5e1;
}

.legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 9999px;
}
</style>
