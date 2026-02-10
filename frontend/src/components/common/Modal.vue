<script setup lang="ts">
withDefaults(
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

const widthMap: Record<string, string> = {
  sm: '420px',
  md: '560px',
  lg: '760px',
  xl: '980px',
};
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    :width="widthMap[size]"
    :close-on-click-modal="closable"
    :show-close="closable"
    append-to-body
    @update:model-value="emit('update:modelValue', $event)"
  >
    <slot />
    <template v-if="$slots.footer" #footer>
      <slot name="footer" />
    </template>
  </el-dialog>
</template>
