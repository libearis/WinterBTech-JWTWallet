/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string
  readonly VITE_IDLE_TIMEOUT_MS?: string
  readonly VITE_REQUEST_TIMEOUT_MS?: string
  readonly VITE_MAX_RETRIES?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
