/**
 * Composables — re-export from @vueuse/core to avoid custom reimplementations.
 * 
 * Previously this file contained 10 hand-rolled composables that duplicated
 * @vueuse/core functionality. They were all dead code (unused in the codebase).
 * Now we re-export the canonical vueuse equivalents so any future usage is
 * automatically powered by the well-tested library.
 */
export {
  useMediaQuery,
  useWindowSize,
  useScroll,
  useOnline,
  useDark as useDarkMode,
  useStorage as useLocalStorage,
  onClickOutside as useClickOutside,
  onKeyStroke as useKeyboard,
  useTitle,
  useDebounceFn,
} from '@vueuse/core';

import { useMediaQuery } from '@vueuse/core';

/** 是否为移动设备（<= 768px） */
export function useMobile() {
  return useMediaQuery('(max-width: 768px)');
}
