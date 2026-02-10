<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { RouterView, useRoute, useRouter } from 'vue-router';
import { useDark, useStorage } from '@vueuse/core';
import { useAuthStore } from '@/stores/auth';
import {
  Menu,
  House,
  MagicStick,
  Document,
  Share,
  Upload,
  Star,
  UserFilled,
  SwitchButton,
  Sunny,
  Moon,
  Fold,
  Expand,
  Key,
} from '@element-plus/icons-vue';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const isDark = useDark({
  selector: 'html',
  attribute: 'class',
  valueDark: 'dark',
  valueLight: '',
});

const mobileSidebarVisible = ref(false);
const desktopSidebarCollapsed = useStorage('lesson-plan:sidebar-collapsed', false);

const menuItems = [
  { name: '首页', path: '/dashboard', icon: House },
  { name: '生成教案', path: '/generate', icon: MagicStick },
  { name: '我的教案', path: '/lessons', icon: Document },
  { name: '知识图谱', path: '/knowledge', icon: Share },
  { name: '知识库管理', path: '/knowledge/upload', icon: Upload },
  { name: '我的收藏', path: '/favorites', icon: Star },
  { name: 'Token与密钥', path: '/token-usage', icon: Key },
];

const activeMenu = computed(() => {
  const currentPath = route.path;
  const exact = menuItems.find((item) => item.path === currentPath);
  if (exact) return exact.path;
  const matched = menuItems.find((item) => currentPath.startsWith(`${item.path}/`));
  return matched?.path || '/dashboard';
});

const currentTitle = computed(() => (route.meta.title as string) || '智能教案生成系统');

const userName = computed(() => authStore.userName || authStore.user?.username || '用户');
const userEmail = computed(() => {
  const directEmail = authStore.user?.email;
  const profileEmail = (authStore.user as { profile?: { email?: string } } | null)?.profile?.email;
  return directEmail || profileEmail || '';
});
const userInitial = computed(() => userName.value.charAt(0).toUpperCase());

function toggleDarkMode() {
  isDark.value = !isDark.value;
}

function toggleDesktopSidebar() {
  desktopSidebarCollapsed.value = !desktopSidebarCollapsed.value;
}

function handleMenuSelect(path: string) {
  router.push(path);
  mobileSidebarVisible.value = false;
}

function goProfile() {
  router.push('/profile');
}

function handleLogout() {
  authStore.logout();
  router.push('/login');
}

onMounted(() => {
  if (authStore.isAuthenticated) {
    authStore.fetchUser();
  }
});
</script>

<template>
  <div class="min-h-screen">
    <el-drawer
      v-model="mobileSidebarVisible"
      direction="ltr"
      :with-header="false"
      size="270px"
      class="lg:hidden"
    >
      <div class="h-full flex flex-col">
        <div class="mb-4 px-2">
          <div class="text-lg font-semibold app-text-primary">智能教案生成系统</div>
        </div>
        <el-menu
          :default-active="activeMenu"
          class="flex-1"
          @select="handleMenuSelect"
        >
          <el-menu-item v-for="item in menuItems" :key="item.path" :index="item.path">
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.name }}</span>
          </el-menu-item>
        </el-menu>
      </div>
    </el-drawer>

    <div
      class="hidden lg:block fixed inset-y-0 left-0 py-5 transition-all duration-300"
      :class="desktopSidebarCollapsed ? 'w-28 px-2' : 'w-72 px-4'"
    >
      <el-card
        class="surface-card h-full sidebar-card"
        shadow="never"
        :class="{ 'sidebar-card-collapsed': desktopSidebarCollapsed }"
      >
        <div class="h-full flex flex-col">
          <div
            class="mb-4 flex items-start"
            :class="desktopSidebarCollapsed ? 'justify-center' : 'justify-between px-1 gap-3'"
          >
            <div v-if="!desktopSidebarCollapsed" class="min-w-0">
              <div class="text-lg font-semibold app-text-primary truncate">智能教案生成系统</div>
            </div>

            <el-tooltip :content="desktopSidebarCollapsed ? '展开侧边栏' : '收起侧边栏'" placement="right">
              <el-button
                :icon="desktopSidebarCollapsed ? Expand : Fold"
                circle
                @click="toggleDesktopSidebar"
              />
            </el-tooltip>
          </div>

          <el-scrollbar class="flex-1">
            <el-menu
              :default-active="activeMenu"
              :collapse="desktopSidebarCollapsed"
              :collapse-transition="false"
              @select="handleMenuSelect"
            >
              <el-menu-item v-for="item in menuItems" :key="item.path" :index="item.path">
                <el-icon><component :is="item.icon" /></el-icon>
                <span>{{ item.name }}</span>
              </el-menu-item>
            </el-menu>
          </el-scrollbar>

          <el-divider class="my-3" />

          <div class="flex items-center px-2" :class="desktopSidebarCollapsed ? 'justify-center' : 'gap-3'">
            <el-avatar :size="38" class="app-avatar">
              {{ userInitial }}
            </el-avatar>
            <div v-if="!desktopSidebarCollapsed" class="min-w-0 flex-1">
              <div class="text-sm font-medium app-text-primary truncate">{{ userName }}</div>
              <div class="text-xs app-text-muted truncate">{{ userEmail || '未绑定邮箱' }}</div>
            </div>
          </div>
        </div>
      </el-card>
    </div>

    <div class="transition-[padding] duration-300" :class="desktopSidebarCollapsed ? 'lg:pl-28' : 'lg:pl-72'">
      <header class="sticky top-0 z-30 px-4 lg:px-8 pt-4">
        <el-card class="surface-card" shadow="never">
          <div class="flex items-center gap-3">
            <el-button class="lg:hidden" :icon="Menu" circle @click="mobileSidebarVisible = true" />

            <div class="min-w-0">
              <div class="text-sm font-semibold app-text-primary truncate">{{ currentTitle }}</div>
              <div class="text-xs app-text-muted">智能教案工作台</div>
            </div>

            <div class="flex-1" />

            <el-tooltip :content="isDark ? '切换到亮色模式' : '切换到暗色模式'" placement="bottom">
              <el-button circle :icon="isDark ? Sunny : Moon" @click="toggleDarkMode" />
            </el-tooltip>

            <el-dropdown trigger="click" placement="bottom-end">
              <span class="inline-flex cursor-pointer items-center">
                <el-avatar :size="34" class="app-avatar">
                  {{ userInitial }}
                </el-avatar>
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="goProfile">
                    <el-icon><UserFilled /></el-icon>
                    <span>个人中心</span>
                  </el-dropdown-item>
                  <el-dropdown-item divided @click="handleLogout">
                    <el-icon><SwitchButton /></el-icon>
                    <span>退出登录</span>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </el-card>
      </header>

      <main class="px-4 py-6 lg:px-8 lg:py-8">
        <div class="mx-auto max-w-[1400px]">
          <RouterView />
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.sidebar-card-collapsed :deep(.el-card__body) {
  padding: 12px 8px;
}

.sidebar-card-collapsed :deep(.el-menu--collapse) {
  width: 100%;
}

.sidebar-card-collapsed :deep(.el-menu-item),
.sidebar-card-collapsed :deep(.el-sub-menu__title) {
  justify-content: center;
}
</style>
