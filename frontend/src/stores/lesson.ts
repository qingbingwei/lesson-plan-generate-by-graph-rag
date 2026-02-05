import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { Lesson, PaginationParams } from '@/types';
import * as lessonApi from '@/api/lesson';

export const useLessonStore = defineStore('lesson', () => {
  // 状态
  const lessons = ref<Lesson[]>([]);
  const currentLesson = ref<Lesson | null>(null);
  const total = ref(0);
  const page = ref(1);
  const pageSize = ref(10);
  const loading = ref(false);
  const error = ref<string | null>(null);

  // 筛选条件
  const filters = ref({
    subject: '',
    grade: '',
    status: '',
    keyword: '',
  });

  // 计算属性
  const totalPages = computed(() => Math.ceil(total.value / pageSize.value));
  const hasMore = computed(() => page.value < totalPages.value);

  // 获取教案列表
  async function fetchLessons(params?: PaginationParams) {
    loading.value = true;
    error.value = null;
    
    try {
      const response = await lessonApi.getLessons({
        page: params?.page || page.value,
        pageSize: params?.pageSize || pageSize.value,
        ...filters.value,
      });
      
      lessons.value = response.items;
      total.value = response.total;
      page.value = response.page;
      pageSize.value = response.pageSize;
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取教案列表失败';
    } finally {
      loading.value = false;
    }
  }

  // 加载更多
  async function loadMore() {
    if (!hasMore.value || loading.value) return;
    
    loading.value = true;
    
    try {
      const response = await lessonApi.getLessons({
        page: page.value + 1,
        pageSize: pageSize.value,
        ...filters.value,
      });
      
      lessons.value = [...lessons.value, ...response.items];
      page.value = response.page;
    } catch (err) {
      error.value = err instanceof Error ? err.message : '加载更多失败';
    } finally {
      loading.value = false;
    }
  }

  // 获取教案详情
  async function fetchLesson(id: string) {
    loading.value = true;
    error.value = null;
    
    try {
      currentLesson.value = await lessonApi.getLesson(id);
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取教案详情失败';
    } finally {
      loading.value = false;
    }
  }

  // 创建教案
  async function createLesson(data: Partial<Lesson>) {
    loading.value = true;
    error.value = null;
    
    try {
      const lesson = await lessonApi.createLesson(data);
      lessons.value.unshift(lesson);
      total.value++;
      return lesson;
    } catch (err) {
      error.value = err instanceof Error ? err.message : '创建教案失败';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  // 更新教案
  async function updateLesson(id: string, data: Partial<Lesson>) {
    loading.value = true;
    error.value = null;
    
    try {
      const lesson = await lessonApi.updateLesson(id, data);
      
      // 更新列表中的教案
      const index = lessons.value.findIndex(l => l.id === id);
      if (index !== -1) {
        lessons.value[index] = lesson;
      }
      
      // 更新当前教案
      if (currentLesson.value?.id === id) {
        currentLesson.value = lesson;
      }
      
      return lesson;
    } catch (err) {
      error.value = err instanceof Error ? err.message : '更新教案失败';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  // 删除教案
  async function deleteLesson(id: string) {
    loading.value = true;
    error.value = null;
    
    try {
      await lessonApi.deleteLesson(id);
      
      // 从列表中移除
      lessons.value = lessons.value.filter(l => l.id !== id);
      total.value--;
      
      // 清除当前教案
      if (currentLesson.value?.id === id) {
        currentLesson.value = null;
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : '删除教案失败';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  // 发布教案
  async function publishLesson(id: string) {
    try {
      await lessonApi.publishLesson(id);
      
      // 发布后重新获取教案详情以更新状态
      await fetchLesson(id);
      
      // 更新列表中的状态
      const index = lessons.value.findIndex(l => l.id === id);
      if (index !== -1) {
        lessons.value[index].status = 'published';
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : '发布失败';
      throw err;
    }
  }

  // 设置筛选条件
  function setFilters(newFilters: Partial<typeof filters.value>) {
    filters.value = { ...filters.value, ...newFilters };
    page.value = 1;
  }

  // 重置筛选条件
  function resetFilters() {
    filters.value = {
      subject: '',
      grade: '',
      status: '',
      keyword: '',
    };
    page.value = 1;
  }

  // 清除当前教案
  function clearCurrentLesson() {
    currentLesson.value = null;
  }

  return {
    // 状态
    lessons,
    currentLesson,
    total,
    page,
    pageSize,
    loading,
    error,
    filters,
    
    // 计算属性
    totalPages,
    hasMore,
    
    // 方法
    fetchLessons,
    loadMore,
    fetchLesson,
    createLesson,
    updateLesson,
    deleteLesson,
    publishLesson,
    setFilters,
    resetFilters,
    clearCurrentLesson,
  };
});
