import { ref, onMounted, onUnmounted } from 'vue';

/**
 * 响应式媒体查询
 */
export function useMediaQuery(query: string) {
  const matches = ref(false);
  
  let mediaQuery: MediaQueryList;
  
  function updateMatches() {
    matches.value = mediaQuery.matches;
  }
  
  onMounted(() => {
    mediaQuery = window.matchMedia(query);
    matches.value = mediaQuery.matches;
    mediaQuery.addEventListener('change', updateMatches);
  });
  
  onUnmounted(() => {
    mediaQuery?.removeEventListener('change', updateMatches);
  });
  
  return matches;
}

/**
 * 是否为移动设备
 */
export function useMobile() {
  return useMediaQuery('(max-width: 768px)');
}

/**
 * 暗色模式
 */
export function useDarkMode() {
  const isDark = ref(false);
  
  function toggle() {
    isDark.value = !isDark.value;
    updateTheme();
  }
  
  function updateTheme() {
    if (isDark.value) {
      document.documentElement.classList.add('dark');
      localStorage.setItem('theme', 'dark');
    } else {
      document.documentElement.classList.remove('dark');
      localStorage.setItem('theme', 'light');
    }
  }
  
  onMounted(() => {
    const stored = localStorage.getItem('theme');
    if (stored === 'dark') {
      isDark.value = true;
    } else if (!stored) {
      isDark.value = window.matchMedia('(prefers-color-scheme: dark)').matches;
    }
    updateTheme();
  });
  
  return { isDark, toggle };
}

/**
 * 本地存储
 */
export function useLocalStorage<T>(key: string, defaultValue: T) {
  const storedValue = localStorage.getItem(key);
  const data = ref<T>(storedValue ? JSON.parse(storedValue) : defaultValue);
  
  function set(value: T) {
    data.value = value;
    localStorage.setItem(key, JSON.stringify(value));
  }
  
  function remove() {
    data.value = defaultValue;
    localStorage.removeItem(key);
  }
  
  return { data, set, remove };
}

/**
 * 窗口大小
 */
export function useWindowSize() {
  const width = ref(window.innerWidth);
  const height = ref(window.innerHeight);
  
  function update() {
    width.value = window.innerWidth;
    height.value = window.innerHeight;
  }
  
  onMounted(() => {
    window.addEventListener('resize', update);
  });
  
  onUnmounted(() => {
    window.removeEventListener('resize', update);
  });
  
  return { width, height };
}

/**
 * 滚动位置
 */
export function useScroll() {
  const x = ref(0);
  const y = ref(0);
  const isScrolling = ref(false);
  
  let timeout: ReturnType<typeof setTimeout>;
  
  function update() {
    x.value = window.scrollX;
    y.value = window.scrollY;
    isScrolling.value = true;
    
    clearTimeout(timeout);
    timeout = setTimeout(() => {
      isScrolling.value = false;
    }, 150);
  }
  
  onMounted(() => {
    window.addEventListener('scroll', update, { passive: true });
  });
  
  onUnmounted(() => {
    window.removeEventListener('scroll', update);
  });
  
  return { x, y, isScrolling };
}

/**
 * 键盘事件
 */
export function useKeyboard(key: string, callback: () => void) {
  function handler(event: KeyboardEvent) {
    if (event.key === key) {
      callback();
    }
  }
  
  onMounted(() => {
    window.addEventListener('keydown', handler);
  });
  
  onUnmounted(() => {
    window.removeEventListener('keydown', handler);
  });
}

/**
 * 点击外部
 */
export function useClickOutside(
  elementRef: { value: HTMLElement | null },
  callback: () => void
) {
  function handler(event: MouseEvent) {
    if (elementRef.value && !elementRef.value.contains(event.target as Node)) {
      callback();
    }
  }
  
  onMounted(() => {
    document.addEventListener('click', handler);
  });
  
  onUnmounted(() => {
    document.removeEventListener('click', handler);
  });
}

/**
 * 在线状态
 */
export function useOnline() {
  const isOnline = ref(navigator.onLine);
  
  function updateOnline() {
    isOnline.value = true;
  }
  
  function updateOffline() {
    isOnline.value = false;
  }
  
  onMounted(() => {
    window.addEventListener('online', updateOnline);
    window.addEventListener('offline', updateOffline);
  });
  
  onUnmounted(() => {
    window.removeEventListener('online', updateOnline);
    window.removeEventListener('offline', updateOffline);
  });
  
  return isOnline;
}

/**
 * 文档标题
 */
export function useTitle(title: string) {
  const originalTitle = document.title;
  
  onMounted(() => {
    document.title = title;
  });
  
  onUnmounted(() => {
    document.title = originalTitle;
  });
}
