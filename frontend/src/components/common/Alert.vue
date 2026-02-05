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

const typeClasses = computed(() => {
  const classes: Record<string, string> = {
    info: 'bg-blue-50 text-blue-800 border-blue-200',
    success: 'bg-green-50 text-green-800 border-green-200',
    warning: 'bg-yellow-50 text-yellow-800 border-yellow-200',
    error: 'bg-red-50 text-red-800 border-red-200',
  };
  return classes[props.type];
});
</script>

<template>
  <div
    class="p-4 rounded-lg border"
    :class="typeClasses"
    role="alert"
  >
    <div class="flex">
      <div class="flex-1">
        <h3 v-if="title" class="text-sm font-medium mb-1">{{ title }}</h3>
        <div class="text-sm">
          <slot />
        </div>
      </div>
      <button
        v-if="closable"
        type="button"
        class="ml-4 inline-flex text-current opacity-60 hover:opacity-100 transition-opacity"
        @click="emit('close')"
      >
        <span class="sr-only">关闭</span>
        <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
          <path
            fill-rule="evenodd"
            d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
            clip-rule="evenodd"
          />
        </svg>
      </button>
    </div>
  </div>
</template>
