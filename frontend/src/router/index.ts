import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { guest: true },
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/Register.vue'),
    meta: { guest: true },
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '首页' },
      },
      {
        path: 'generate',
        name: 'Generate',
        component: () => import('@/views/Generate.vue'),
        meta: { title: '生成教案' },
      },
      {
        path: 'lessons',
        name: 'Lessons',
        component: () => import('@/views/Lessons.vue'),
        meta: { title: '我的教案' },
      },
      {
        path: 'lessons/:id',
        name: 'LessonDetail',
        component: () => import('@/views/LessonDetail.vue'),
        meta: { title: '教案详情' },
      },
      {
        path: 'lessons/:id/edit',
        name: 'LessonEdit',
        component: () => import('@/views/LessonEdit.vue'),
        meta: { title: '编辑教案' },
      },
      {
        path: 'knowledge',
        name: 'Knowledge',
        component: () => import('@/views/Knowledge.vue'),
        meta: { title: '知识图谱' },
      },
      {
        path: 'knowledge/upload',
        name: 'KnowledgeUpload',
        component: () => import('@/views/KnowledgeUpload.vue'),
        meta: { title: '知识库管理' },
      },
      {
        path: 'favorites',
        name: 'Favorites',
        component: () => import('@/views/Favorites.vue'),
        meta: { title: '我的收藏' },
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/Profile.vue'),
        meta: { title: '个人中心' },
      },
      {
        path: 'token-usage',
        name: 'TokenUsage',
        component: () => import('@/views/TokenUsage.vue'),
        meta: { title: 'Token与密钥' },
      },

      {
        path: 'help-docs',
        name: 'HelpDocs',
        component: () => import('@/views/HelpDocs.vue'),
        meta: { title: '帮助文档' },
      },
      {
        path: 'ai-assistant',
        name: 'AiAssistant',
        component: () => import('@/views/AiAssistant.vue'),
        meta: { title: '智能问答' },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue'),
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior() {
    return { top: 0 };
  },
});

router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore();

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'Login', query: { redirect: to.fullPath } });
    return;
  }

  if (to.meta.guest && authStore.isAuthenticated) {
    next({ name: 'Dashboard' });
    return;
  }

  next();
});

router.afterEach((to) => {
  const title = to.meta.title as string;
  document.title = title ? `${title} - 智能教案生成系统` : '智能教案生成系统';
});

export default router;
