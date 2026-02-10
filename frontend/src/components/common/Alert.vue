<script setup lang="ts">
import { computed } from 'vue';

const props = withDefaults(
  defineProps<{
    type?: 'info' | 'success' | 'warning' | 'error';
    title?: string;
    closable?: boolean;
  }>(),
  {
    type: 'info',
    closable: false,
  }
);

const emit = defineEmits<{
  close: [];
}>();

const elType = computed(() => {
  const map: Record<string, 'success' | 'warning' | 'info' | 'error'> = {
    info: 'info',
    success: 'success',
    warning: 'warning',
    error: 'error',
  };
  return map[props.type];
});
</script>

<template>
  <el-alert
    :title="title"
    :type="elType"
    :closable="closable"
    show-icon
    class="w-full"
    @close="emit('close')"
  >
    <template v-if="$slots.default" #default>
      <slot />
    </template>
  </el-alert>
</template>
