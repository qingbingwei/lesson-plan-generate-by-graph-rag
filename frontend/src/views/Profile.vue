<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { useAuthStore } from '@/stores/auth';
import { changePassword, updateProfile } from '@/api/auth';
import { User, Lock, Bell } from '@element-plus/icons-vue';

const authStore = useAuthStore();

const activeTab = ref('profile');
const savingProfile = ref(false);
const savingPassword = ref(false);
const savingNotifications = ref(false);

const profileForm = ref({
  name: '',
  email: '',
  avatar: '',
});

const passwordForm = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
});

const notifications = ref({
  email: true,
  browser: false,
  lessonReminder: true,
  weeklyReport: false,
});

function loadUserInfo() {
  if (!authStore.user) {
    return;
  }

  profileForm.value = {
    name: authStore.user.profile?.name || authStore.user.username,
    email: authStore.user.email,
    avatar: authStore.user.profile?.avatar || '',
  };
}

async function saveProfile() {
  if (!profileForm.value.name.trim()) {
    ElMessage.warning('请输入姓名');
    return;
  }

  savingProfile.value = true;
  try {
    await updateProfile(profileForm.value as any);
    await authStore.fetchUser();
    ElMessage.success('个人信息已更新');
  } catch {
    ElMessage.error('保存失败，请重试');
  } finally {
    savingProfile.value = false;
  }
}

async function savePassword() {
  if (!passwordForm.value.oldPassword) {
    ElMessage.warning('请输入原密码');
    return;
  }

  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    ElMessage.error('两次输入的密码不一致');
    return;
  }

  if (passwordForm.value.newPassword.length < 6) {
    ElMessage.error('密码长度至少为 6 位');
    return;
  }

  savingPassword.value = true;
  try {
    await changePassword({
      oldPassword: passwordForm.value.oldPassword,
      newPassword: passwordForm.value.newPassword,
    });

    passwordForm.value = {
      oldPassword: '',
      newPassword: '',
      confirmPassword: '',
    };

    ElMessage.success('密码已修改');
  } catch {
    ElMessage.error('修改失败，请检查原密码是否正确');
  } finally {
    savingPassword.value = false;
  }
}

async function saveNotifications() {
  savingNotifications.value = true;
  try {
    localStorage.setItem('notifications', JSON.stringify(notifications.value));
    ElMessage.success('通知设置已保存');
  } catch {
    ElMessage.error('保存失败，请重试');
  } finally {
    savingNotifications.value = false;
  }
}

onMounted(() => {
  loadUserInfo();

  const stored = localStorage.getItem('notifications');
  if (stored) {
    notifications.value = JSON.parse(stored);
  }
});
</script>

<template>
  <div class="page-container max-w-4xl mx-auto">
    <div class="page-header">
      <h1 class="page-title">个人设置</h1>
      <p class="page-subtitle">管理您的账户信息和偏好设置</p>
    </div>

    <el-card class="surface-card" shadow="never">
      <el-tabs v-model="activeTab">
        <el-tab-pane name="profile">
          <template #label>
            <span class="inline-flex items-center gap-1">
              <el-icon><User /></el-icon>
              <span>个人信息</span>
            </span>
          </template>

          <div class="space-y-5">
            <el-card class="surface-card" shadow="never">
              <div class="flex flex-col sm:flex-row sm:items-center gap-4">
                <el-avatar
                  :size="72"
                  class="app-avatar"
                >
                  {{ profileForm.name?.charAt(0) || 'U' }}
                </el-avatar>
                <div class="min-w-0 flex-1">
                  <div class="text-sm app-text-secondary">头像地址（可选）</div>
                  <el-input v-model="profileForm.avatar" placeholder="请输入头像 URL" />
                </div>
              </div>
            </el-card>

            <el-form :model="profileForm" label-position="top" class="max-w-xl">
              <el-form-item label="姓名" required>
                <el-input v-model="profileForm.name" placeholder="请输入姓名" />
              </el-form-item>

              <el-form-item label="邮箱">
                <el-input v-model="profileForm.email" disabled />
                <div class="text-xs app-text-muted mt-1">邮箱地址不可修改</div>
              </el-form-item>
            </el-form>

            <div class="flex justify-end">
              <el-button type="primary" :loading="savingProfile" @click="saveProfile">保存信息</el-button>
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane name="password">
          <template #label>
            <span class="inline-flex items-center gap-1">
              <el-icon><Lock /></el-icon>
              <span>修改密码</span>
            </span>
          </template>

          <el-form :model="passwordForm" label-position="top" class="max-w-xl">
            <el-form-item label="原密码" required>
              <el-input v-model="passwordForm.oldPassword" type="password" show-password placeholder="请输入原密码" />
            </el-form-item>

            <el-form-item label="新密码" required>
              <el-input v-model="passwordForm.newPassword" type="password" show-password placeholder="请输入新密码" />
            </el-form-item>

            <el-form-item label="确认新密码" required>
              <el-input
                v-model="passwordForm.confirmPassword"
                type="password"
                show-password
                placeholder="请再次输入新密码"
              />
            </el-form-item>
          </el-form>

          <div class="flex justify-end">
            <el-button type="primary" :loading="savingPassword" @click="savePassword">修改密码</el-button>
          </div>
        </el-tab-pane>

        <el-tab-pane name="notifications">
          <template #label>
            <span class="inline-flex items-center gap-1">
              <el-icon><Bell /></el-icon>
              <span>通知设置</span>
            </span>
          </template>

          <el-card class="surface-card" shadow="never">
            <el-form label-position="left" label-width="140px">
              <el-form-item label="邮件通知">
                <div class="flex items-center justify-between w-full gap-4">
                  <span class="app-text-muted">接收邮件通知</span>
                  <el-switch v-model="notifications.email" />
                </div>
              </el-form-item>

              <el-form-item label="浏览器通知">
                <div class="flex items-center justify-between w-full gap-4">
                  <span class="app-text-muted">接收桌面推送通知</span>
                  <el-switch v-model="notifications.browser" />
                </div>
              </el-form-item>

              <el-form-item label="教案提醒">
                <div class="flex items-center justify-between w-full gap-4">
                  <span class="app-text-muted">教案生成完成时通知</span>
                  <el-switch v-model="notifications.lessonReminder" />
                </div>
              </el-form-item>

              <el-form-item label="每周报告">
                <div class="flex items-center justify-between w-full gap-4">
                  <span class="app-text-muted">每周发送使用情况汇总</span>
                  <el-switch v-model="notifications.weeklyReport" />
                </div>
              </el-form-item>
            </el-form>

            <div class="flex justify-end">
              <el-button type="primary" :loading="savingNotifications" @click="saveNotifications">保存设置</el-button>
            </div>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>
