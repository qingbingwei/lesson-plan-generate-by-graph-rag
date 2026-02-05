<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue';
import * as d3 from 'd3';
import api from '@/api/index';
import type { KnowledgeNode, KnowledgeLink, KnowledgePoint, KnowledgeNodeType, ApiResponse } from '@/types';
import { MagnifyingGlassIcon } from '@heroicons/vue/24/outline';

// 后端返回的图谱数据结构
interface GraphApiResponse {
  nodes: Array<{
    id: string;
    label: string;
    type: string;
    subject: string;
    grade: string;
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

// 筛选条件
const filters = ref({
  subject: '',  // 空字符串表示所有学科
  grade: '',    // 空字符串表示所有年级
  topic: '',
});

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
      { params: { subject, grade } }
    );
    
    const data = response.data.data;
    
    console.log('Knowledge graph API response:', data);
    console.log('Nodes count:', data.nodes?.length);
    console.log('Edges count:', data.edges?.length);
    
    // 转换节点数据 - 确保 type 是有效的 KnowledgeNodeType
    graphData.value.nodes = data.nodes.map(node => ({
      id: node.id,
      name: node.label,
      type: node.type as KnowledgeNodeType,
      properties: {
        subject: node.subject,
        grade: node.grade,
      },
    }));
    
    // 转换边数据为 d3 需要的格式
    graphData.value.links = data.edges.map(edge => ({
      source: edge.source,
      target: edge.target,
      type: edge.type,
      properties: {
        weight: edge.weight,
      },
    }));
    
    // 提取知识点列表用于侧边栏显示
    knowledgePoints.value = data.nodes.map(node => ({
      id: node.id,
      name: node.label,
      description: `${node.type} - ${node.grade || '通用'}`,
      difficulty: 'medium',
      grade: node.grade || '',
      importance: 3,
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
  // 高亮相关节点 (可以后续扩展)
}

// 渲染图谱
function renderGraph() {
  console.log('renderGraph called, nodes:', graphData.value.nodes.length, 'svgRef:', !!svgRef.value);
  
  if (!svgRef.value || graphData.value.nodes.length === 0) {
    console.log('renderGraph skipped - no svg or no nodes');
    return;
  }

  const svg = d3.select(svgRef.value);
  svg.selectAll('*').remove();

  // 获取 SVG 容器尺寸，确保有最小值
  const width = Math.max(svgRef.value.clientWidth, 600);
  const height = Math.max(svgRef.value.clientHeight, 400);
  
  console.log('SVG dimensions:', width, 'x', height);
  
  // 设置 SVG viewBox 确保可见
  svg.attr('viewBox', `0 0 ${width} ${height}`);

  const g = svg.append('g');

  // 缩放
  const zoom = d3.zoom<SVGSVGElement, unknown>()
    .scaleExtent([0.1, 4])
    .on('zoom', (event) => {
      g.attr('transform', event.transform);
    });

  svg.call(zoom);

  // 深拷贝节点和边数据，避免 D3 修改响应式数据
  const nodes = graphData.value.nodes.map(n => ({ ...n }));
  const links = graphData.value.links.map(l => ({ ...l }));

  // 力导向图
  const simulation = d3.forceSimulation(nodes as d3.SimulationNodeDatum[])
    .force('link', d3.forceLink(links)
      .id((d: unknown) => (d as KnowledgeNode).id)
      .distance(100))
    .force('charge', d3.forceManyBody().strength(-300))
    .force('center', d3.forceCenter(width / 2, height / 2))
    .force('collision', d3.forceCollide().radius(50));

  // 连线
  const link = g.append('g')
    .selectAll('line')
    .data(links)
    .join('line')
    .attr('stroke', '#cbd5e1')
    .attr('stroke-width', 2)
    .attr('stroke-opacity', 0.6);
  
  // 节点
  const node = g.append('g')
    .selectAll('g')
    .data(nodes)
    .join('g')
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
    .attr('r', 25)
    .attr('fill', (d) => getNodeColor(d.type))
    .attr('stroke', '#fff')
    .attr('stroke-width', 2);

  // 节点文字
  node.append('text')
    .text((d) => d.name.substring(0, 4))
    .attr('text-anchor', 'middle')
    .attr('dy', '0.35em')
    .attr('fill', '#fff')
    .attr('font-size', '12px')
    .attr('font-weight', 'bold');

  // 更新位置
  simulation.on('tick', () => {
    link
      .attr('x1', (d) => (d.source as d3.SimulationNodeDatum).x!)
      .attr('y1', (d) => (d.source as d3.SimulationNodeDatum).y!)
      .attr('x2', (d) => (d.target as d3.SimulationNodeDatum).x!)
      .attr('y2', (d) => (d.target as d3.SimulationNodeDatum).y!);

    node.attr('transform', (d) => 
      `translate(${(d as d3.SimulationNodeDatum).x},${(d as d3.SimulationNodeDatum).y})`
    );
  });
}

// 获取节点颜色
function getNodeColor(type: string): string {
  const colors: Record<string, string> = {
    Subject: '#3b82f6',
    Chapter: '#8b5cf6',
    KnowledgePoint: '#10b981',
    Skill: '#f59e0b',
    Resource: '#ec4899',
    Lesson: '#6366f1',
  };
  return colors[type] || '#6b7280';
}

// 处理窗口大小变化
function handleResize() {
  renderGraph();
}

onMounted(() => {
  loadKnowledgeGraph();
  window.addEventListener('resize', handleResize);
});

onUnmounted(() => {
  window.removeEventListener('resize', handleResize);
});
</script>

<template>
  <div class="h-[calc(100vh-12rem)] flex gap-6">
    <!-- 左侧面板 -->
    <div class="w-80 flex-shrink-0 flex flex-col space-y-4">
      <!-- 筛选 -->
      <div class="card">
        <div class="card-body space-y-3">
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
          <div>
            <label class="label">关键词</label>
            <div class="relative">
              <MagnifyingGlassIcon class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
              <input
                v-model="filters.topic"
                type="text"
                class="input pl-9"
                placeholder="搜索..."
                @keyup.enter="loadKnowledgeGraph"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- 知识点列表 -->
      <div class="card flex-1 overflow-hidden flex flex-col">
        <div class="card-header">
          <h3 class="font-medium">知识点</h3>
        </div>
        <div class="flex-1 overflow-auto">
          <div v-if="loading" class="p-4 text-center">
            <div class="loading" />
          </div>
          <ul v-else class="divide-y divide-gray-100">
            <li
              v-for="point in knowledgePoints"
              :key="point.id"
              class="p-3 hover:bg-gray-50 cursor-pointer transition-colors"
              :class="{ 'bg-primary-50': selectedNode?.id === point.id }"
              @click="selectKnowledgePoint(point)"
            >
              <h4 class="font-medium text-sm text-gray-900">{{ point.name }}</h4>
              <p class="text-xs text-gray-500 mt-1 line-clamp-2">{{ point.description }}</p>
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
    <div class="flex-1 card overflow-hidden flex flex-col">
      <div class="card-header flex items-center justify-between flex-shrink-0">
        <h3 class="font-medium">知识图谱</h3>
        <div class="flex gap-2">
          <span class="flex items-center gap-1 text-xs">
            <span class="w-3 h-3 rounded-full bg-blue-500"></span>
            学科
          </span>
          <span class="flex items-center gap-1 text-xs">
            <span class="w-3 h-3 rounded-full bg-purple-500"></span>
            章节
          </span>
          <span class="flex items-center gap-1 text-xs">
            <span class="w-3 h-3 rounded-full bg-green-500"></span>
            知识点
          </span>
          <span class="flex items-center gap-1 text-xs">
            <span class="w-3 h-3 rounded-full bg-yellow-500"></span>
            技能
          </span>
        </div>
      </div>
      <div class="flex-1 relative min-h-0">
        <svg
          ref="svgRef"
          class="w-full h-full absolute inset-0"
          :class="{ 'opacity-50': loading }"
          style="min-height: 400px;"
        />
        <div
          v-if="graphData.nodes.length === 0 && !loading"
          class="absolute inset-0 flex items-center justify-center"
        >
          <div class="text-center text-gray-500">
            <p>暂无知识图谱数据</p>
            <p class="text-sm mt-1">请尝试调整筛选条件</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
