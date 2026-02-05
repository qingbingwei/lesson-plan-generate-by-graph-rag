<script setup lang="ts">
import { computed } from 'vue';

const props = withDefaults(
  defineProps<{
    modelValue: boolean;
    title?: string;
    size?: 'sm' | 'md' | 'lg' | 'xl';
    closable?: boolean;
  }>(),
  {
    size: 'md',
    closable: true,
  }
);

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
}>();

const sizeClasses = computed(() => {
  const classes: Record<string, string> = {
    sm: 'max-w-sm',
    md: 'max-w-md',
    lg: 'max-w-lg',
    xl: 'max-w-xl',
  };
  return classes[props.size];
});

function close() {
  if (props.closable) {
    emit('update:modelValue', false);
  }
}

function handleBackdropClick(event: MouseEvent) {
  if (event.target === event.currentTarget) {
    close();
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-200"
      leave-active-class="transition-opacity duration-200"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
    >
      <div
        v-if="modelValue"
        class="fixed inset-0 z-50 overflow-y-auto"
        @click="handleBackdropClick"
      >
        <div class="flex min-h-full items-center justify-center p-4">
          <div
            class="fixed inset-0 bg-black/50 transition-opacity"
            aria-hidden="true"
          />

          <Transition
            enter-active-class="transition-all duration-200"
            leave-active-class="transition-all duration-200"
            enter-from-class="opacity-0 scale-95"
            leave-to-class="opacity-0 scale-95"
          >
            <div
              v-if="modelValue"
              class="relative w-full rounded-lg bg-white shadow-xl"
              :class="sizeClasses"
            >
              <!-- Header -->
              <div
                v-if="title || closable"
                class="flex items-center justify-between border-b border-gray-200 px-6 py-4"
              >
                <h3 v-if="title" class="text-lg font-medium text-gray-900">
                  {{ title }}
                </h3>
                <button
                  v-if="closable"
                  type="button"
                  class="ml-auto text-gray-400 hover:text-gray-500 transition-colors"
                  @click="close"
                >
                  <span class="sr-only">关闭</span>
                  <svg
                    class="h-6 w-6"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke-width="1.5"
                    stroke="currentColor"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </button>
              </div>

              <!-- Body -->
              <div class="px-6 py-4">
                <slot />
              </div>

              <!-- Footer -->
              <div
                v-if="$slots.footer"
                class="border-t border-gray-200 px-6 py-4"
              >
                <slot name="footer" />
              </div>
            </div>
          </Transition>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
