<script setup lang="ts">
import { ref, computed } from 'vue';
import { RouterLink, RouterView, useRoute } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import {
  Bars3Icon,
  XMarkIcon,
  HomeIcon,
  DocumentTextIcon,
  SparklesIcon,
  AcademicCapIcon,
  HeartIcon,
  UserCircleIcon,
  ArrowRightOnRectangleIcon,
  CloudArrowUpIcon,
} from '@heroicons/vue/24/outline';

const route = useRoute();
const authStore = useAuthStore();

const isSidebarOpen = ref(false);
const isUserMenuOpen = ref(false);

const navigation = [
  { name: '首页', href: '/dashboard', icon: HomeIcon },
  { name: '生成教案', href: '/generate', icon: SparklesIcon },
  { name: '我的教案', href: '/lessons', icon: DocumentTextIcon },
  { name: '知识图谱', href: '/knowledge', icon: AcademicCapIcon },
  { name: '知识库管理', href: '/knowledge/upload', icon: CloudArrowUpIcon },
  { name: '我的收藏', href: '/favorites', icon: HeartIcon },
];

const currentRoute = computed(() => route.path);

function isActive(href: string) {
  return currentRoute.value === href || currentRoute.value.startsWith(href + '/');
}

function handleLogout() {
  authStore.logout();
  window.location.href = '/login';
}
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Mobile sidebar backdrop -->
    <transition name="fade">
      <div
        v-if="isSidebarOpen"
        class="fixed inset-0 z-40 bg-gray-900/50 lg:hidden"
        @click="isSidebarOpen = false"
      />
    </transition>

    <!-- Mobile sidebar -->
    <transition name="slide-right">
      <div
        v-if="isSidebarOpen"
        class="fixed inset-y-0 left-0 z-50 w-64 bg-white shadow-xl lg:hidden"
      >
        <div class="flex h-16 items-center justify-between px-4 border-b">
          <span class="text-xl font-bold text-gradient">智能教案</span>
          <button
            type="button"
            class="btn-icon"
            @click="isSidebarOpen = false"
          >
            <XMarkIcon class="h-6 w-6" />
          </button>
        </div>
        <nav class="p-4 space-y-1">
          <RouterLink
            v-for="item in navigation"
            :key="item.name"
            :to="item.href"
            class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors"
            :class="[
              isActive(item.href)
                ? 'bg-primary-50 text-primary-700'
                : 'text-gray-700 hover:bg-gray-100'
            ]"
            @click="isSidebarOpen = false"
          >
            <component :is="item.icon" class="h-5 w-5" />
            {{ item.name }}
          </RouterLink>
        </nav>
      </div>
    </transition>

    <!-- Desktop sidebar -->
    <div class="hidden lg:fixed lg:inset-y-0 lg:flex lg:w-64 lg:flex-col">
      <div class="flex flex-1 flex-col bg-white border-r border-gray-200">
        <div class="flex h-16 items-center px-6 border-b">
          <span class="text-xl font-bold text-gradient">智能教案生成系统</span>
        </div>
        <nav class="flex-1 p-4 space-y-1">
          <RouterLink
            v-for="item in navigation"
            :key="item.name"
            :to="item.href"
            class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors"
            :class="[
              isActive(item.href)
                ? 'bg-primary-50 text-primary-700'
                : 'text-gray-700 hover:bg-gray-100'
            ]"
          >
            <component :is="item.icon" class="h-5 w-5" />
            {{ item.name }}
          </RouterLink>
        </nav>
        <div class="p-4 border-t">
          <div class="flex items-center gap-3 px-3 py-2">
            <div class="flex-shrink-0">
              <div class="h-8 w-8 rounded-full bg-primary-100 flex items-center justify-center">
                <span class="text-sm font-medium text-primary-700">
                  {{ authStore.userName.charAt(0).toUpperCase() }}
                </span>
              </div>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-gray-900 truncate">
                {{ authStore.userName }}
              </p>
              <p class="text-xs text-gray-500 truncate">
                {{ authStore.user?.email }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Main content -->
    <div class="lg:pl-64">
      <!-- Top navbar -->
      <header class="sticky top-0 z-30 flex h-16 items-center gap-4 border-b bg-white px-4 lg:px-8">
        <button
          type="button"
          class="btn-icon lg:hidden"
          @click="isSidebarOpen = true"
        >
          <Bars3Icon class="h-6 w-6" />
        </button>

        <div class="flex-1" />

        <!-- User menu -->
        <div class="relative">
          <button
            type="button"
            class="flex items-center gap-2 rounded-full p-1 hover:bg-gray-100"
            @click="isUserMenuOpen = !isUserMenuOpen"
          >
            <div class="h-8 w-8 rounded-full bg-primary-100 flex items-center justify-center">
              <span class="text-sm font-medium text-primary-700">
                {{ authStore.userName.charAt(0).toUpperCase() }}
              </span>
            </div>
          </button>

          <transition name="scale">
            <div
              v-if="isUserMenuOpen"
              class="absolute right-0 mt-2 w-48 rounded-lg bg-white shadow-lg ring-1 ring-black ring-opacity-5"
            >
              <div class="py-1">
                <RouterLink
                  to="/profile"
                  class="flex items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  @click="isUserMenuOpen = false"
                >
                  <UserCircleIcon class="h-5 w-5" />
                  个人中心
                </RouterLink>
                <button
                  type="button"
                  class="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  @click="handleLogout"
                >
                  <ArrowRightOnRectangleIcon class="h-5 w-5" />
                  退出登录
                </button>
              </div>
            </div>
          </transition>
        </div>
      </header>

      <!-- Page content -->
      <main class="p-4 lg:p-8">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<style scoped>
.slide-right-enter-active,
.slide-right-leave-active {
  transition: transform 0.3s ease;
}

.slide-right-enter-from,
.slide-right-leave-to {
  transform: translateX(-100%);
}
</style>
