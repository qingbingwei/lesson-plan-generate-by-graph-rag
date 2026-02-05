<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { updateProfile, changePassword } from '@/api/auth';
import {
  UserCircleIcon,
  KeyIcon,
  BellIcon,
} from '@heroicons/vue/24/outline';

const authStore = useAuthStore();

const activeTab = ref('profile');
const loading = ref(false);
const message = ref({ type: '', text: '' });

// 个人信息表单
const profileForm = ref({
  name: '',
  email: '',
  avatar: '',
});

// 密码表单
const passwordForm = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
});

// 通知设置
const notifications = ref({
  email: true,
  browser: false,
  lessonReminder: true,
  weeklyReport: false,
});

// 加载用户信息
function loadUserInfo() {
  if (authStore.user) {
    profileForm.value = {
      name: authStore.user.profile?.name || authStore.user.username,
      email: authStore.user.email,
      avatar: authStore.user.profile?.avatar || '',
    };
  }
}

// 保存个人信息
async function saveProfile() {
  loading.value = true;
  message.value = { type: '', text: '' };
  
  try {
    await updateProfile(profileForm.value);
    await authStore.fetchUser();
    message.value = { type: 'success', text: '个人信息已更新' };
  } catch {
    message.value = { type: 'error', text: '保存失败，请重试' };
  } finally {
    loading.value = false;
  }
}

// 修改密码
async function savePassword() {
  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    message.value = { type: 'error', text: '两次输入的密码不一致' };
    return;
  }
  
  if (passwordForm.value.newPassword.length < 6) {
    message.value = { type: 'error', text: '密码长度至少为6位' };
    return;
  }
  
  loading.value = true;
  message.value = { type: '', text: '' };
  
  try {
    await changePassword({
      oldPassword: passwordForm.value.oldPassword,
      newPassword: passwordForm.value.newPassword
    });
    message.value = { type: 'success', text: '密码已修改' };
    passwordForm.value = {
      oldPassword: '',
      newPassword: '',
      confirmPassword: '',
    };
  } catch {
    message.value = { type: 'error', text: '修改失败，请检查原密码是否正确' };
  } finally {
    loading.value = false;
  }
}

// 保存通知设置
async function saveNotifications() {
  loading.value = true;
  message.value = { type: '', text: '' };
  
  try {
    // 保存到本地存储
    localStorage.setItem('notifications', JSON.stringify(notifications.value));
    message.value = { type: 'success', text: '通知设置已保存' };
  } catch {
    message.value = { type: 'error', text: '保存失败' };
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  loadUserInfo();
  // 加载通知设置
  const stored = localStorage.getItem('notifications');
  if (stored) {
    notifications.value = JSON.parse(stored);
  }
});
</script>

<template>
  <div class="max-w-3xl mx-auto space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-bold text-gray-900">个人设置</h1>
      <p class="mt-1 text-sm text-gray-500">
        管理您的账户信息和偏好设置
      </p>
    </div>

    <!-- Tabs -->
    <div class="border-b border-gray-200">
      <nav class="-mb-px flex space-x-8">
        <button
          type="button"
          class="py-4 px-1 border-b-2 font-medium text-sm transition-colors"
          :class="[
            activeTab === 'profile'
              ? 'border-primary-500 text-primary-600'
              : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
          ]"
          @click="activeTab = 'profile'"
        >
          <UserCircleIcon class="h-5 w-5 inline-block mr-1" />
          个人信息
        </button>
        <button
          type="button"
          class="py-4 px-1 border-b-2 font-medium text-sm transition-colors"
          :class="[
            activeTab === 'password'
              ? 'border-primary-500 text-primary-600'
              : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
          ]"
          @click="activeTab = 'password'"
        >
          <KeyIcon class="h-5 w-5 inline-block mr-1" />
          修改密码
        </button>
        <button
          type="button"
          class="py-4 px-1 border-b-2 font-medium text-sm transition-colors"
          :class="[
            activeTab === 'notifications'
              ? 'border-primary-500 text-primary-600'
              : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
          ]"
          @click="activeTab = 'notifications'"
        >
          <BellIcon class="h-5 w-5 inline-block mr-1" />
          通知设置
        </button>
      </nav>
    </div>

    <!-- Message -->
    <div
      v-if="message.text"
      class="p-4 rounded-lg"
      :class="[
        message.type === 'success'
          ? 'bg-green-50 text-green-700'
          : 'bg-red-50 text-red-700',
      ]"
    >
      {{ message.text }}
    </div>

    <!-- Profile Tab -->
    <div v-if="activeTab === 'profile'" class="card">
      <div class="card-body space-y-6">
        <div class="flex items-center gap-4">
          <div
            class="h-20 w-20 rounded-full bg-primary-100 flex items-center justify-center text-primary-600 text-2xl font-bold"
          >
            {{ profileForm.name?.charAt(0) || 'U' }}
          </div>
          <div>
            <button type="button" class="btn-outline btn-sm">
              上传头像
            </button>
            <p class="text-xs text-gray-500 mt-1">
              支持 JPG, PNG 格式，最大 2MB
            </p>
          </div>
        </div>

        <div class="grid grid-cols-1 gap-4">
          <div>
            <label class="label">姓名</label>
            <input v-model="profileForm.name" type="text" class="input" />
          </div>
          <div>
            <label class="label">邮箱</label>
            <input
              v-model="profileForm.email"
              type="email"
              class="input"
              disabled
            />
            <p class="text-xs text-gray-500 mt-1">邮箱地址不可修改</p>
          </div>
        </div>

        <div class="flex justify-end">
          <button
            type="button"
            class="btn-primary"
            :disabled="loading"
            @click="saveProfile"
          >
            {{ loading ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Password Tab -->
    <div v-if="activeTab === 'password'" class="card">
      <div class="card-body space-y-4">
        <div>
          <label class="label">原密码</label>
          <input
            v-model="passwordForm.oldPassword"
            type="password"
            class="input"
            placeholder="请输入原密码"
          />
        </div>
        <div>
          <label class="label">新密码</label>
          <input
            v-model="passwordForm.newPassword"
            type="password"
            class="input"
            placeholder="请输入新密码"
          />
        </div>
        <div>
          <label class="label">确认新密码</label>
          <input
            v-model="passwordForm.confirmPassword"
            type="password"
            class="input"
            placeholder="请再次输入新密码"
          />
        </div>

        <div class="flex justify-end">
          <button
            type="button"
            class="btn-primary"
            :disabled="loading"
            @click="savePassword"
          >
            {{ loading ? '保存中...' : '修改密码' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Notifications Tab -->
    <div v-if="activeTab === 'notifications'" class="card">
      <div class="card-body space-y-4">
        <div class="flex items-center justify-between py-3 border-b border-gray-100">
          <div>
            <h4 class="font-medium text-gray-900">邮件通知</h4>
            <p class="text-sm text-gray-500">接收邮件通知</p>
          </div>
          <label class="relative inline-flex items-center cursor-pointer">
            <input
              v-model="notifications.email"
              type="checkbox"
              class="sr-only peer"
            />
            <div
              class="w-11 h-6 bg-gray-200 rounded-full peer peer-checked:bg-primary-600 peer-focus:ring-4 peer-focus:ring-primary-300 after:content-[''] after:absolute after:top-0.5 after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:after:translate-x-full"
            ></div>
          </label>
        </div>

        <div class="flex items-center justify-between py-3 border-b border-gray-100">
          <div>
            <h4 class="font-medium text-gray-900">浏览器通知</h4>
            <p class="text-sm text-gray-500">接收桌面推送通知</p>
          </div>
          <label class="relative inline-flex items-center cursor-pointer">
            <input
              v-model="notifications.browser"
              type="checkbox"
              class="sr-only peer"
            />
            <div
              class="w-11 h-6 bg-gray-200 rounded-full peer peer-checked:bg-primary-600 peer-focus:ring-4 peer-focus:ring-primary-300 after:content-[''] after:absolute after:top-0.5 after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:after:translate-x-full"
            ></div>
          </label>
        </div>

        <div class="flex items-center justify-between py-3 border-b border-gray-100">
          <div>
            <h4 class="font-medium text-gray-900">教案提醒</h4>
            <p class="text-sm text-gray-500">教案生成完成时通知</p>
          </div>
          <label class="relative inline-flex items-center cursor-pointer">
            <input
              v-model="notifications.lessonReminder"
              type="checkbox"
              class="sr-only peer"
            />
            <div
              class="w-11 h-6 bg-gray-200 rounded-full peer peer-checked:bg-primary-600 peer-focus:ring-4 peer-focus:ring-primary-300 after:content-[''] after:absolute after:top-0.5 after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:after:translate-x-full"
            ></div>
          </label>
        </div>

        <div class="flex items-center justify-between py-3">
          <div>
            <h4 class="font-medium text-gray-900">周报</h4>
            <p class="text-sm text-gray-500">每周发送使用情况汇总</p>
          </div>
          <label class="relative inline-flex items-center cursor-pointer">
            <input
              v-model="notifications.weeklyReport"
              type="checkbox"
              class="sr-only peer"
            />
            <div
              class="w-11 h-6 bg-gray-200 rounded-full peer peer-checked:bg-primary-600 peer-focus:ring-4 peer-focus:ring-primary-300 after:content-[''] after:absolute after:top-0.5 after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:after:translate-x-full"
            ></div>
          </label>
        </div>

        <div class="flex justify-end pt-4">
          <button
            type="button"
            class="btn-primary"
            :disabled="loading"
            @click="saveNotifications"
          >
            {{ loading ? '保存中...' : '保存设置' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
