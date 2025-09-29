import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useApi } from '@/services/ApiService'

export interface AxiosErrorLike extends Error {
  isAxiosError?: boolean
  response?: {
    status?: number
    statusText?: string
    data?: any
  }
}
export const useNotificationStore = defineStore('notification', () => {
  type NotificationType = 'info' | 'success' | 'warning' | 'error'

  interface Notification {
    id: number
    message: string
    type: NotificationType
    progress: number
  }

  const notifications = ref<Notification[]>([])
  const nextId = ref(1)
  const api = useApi()

  // Network/backend status
  const isOffline = ref(!navigator.onLine)
  const backendDown = ref(false)
  let intervalId: ReturnType<typeof setInterval> | null = null

  function show(message: string | Error, type: NotificationType = 'info', timeout = 3000) {
    let text = ''

    if (message instanceof Error) {
      text = `Fehler: ${message.message}`
    } else {
      text = message
    }

    const id = nextId.value++
    notifications.value.push({ id, message: text, type, progress: 100 })

    if (timeout > 0) {
      const start = Date.now()
      const interval = 50
      const timer = setInterval(() => {
        const elapsed = Date.now() - start
        const remaining = timeout - elapsed
        if (remaining <= 0) {
          notifications.value = notifications.value.filter((n) => n.id !== id)
          clearInterval(timer)
        } else {
          const index = notifications.value.findIndex((n) => n.id === id)
          if (index !== -1) {
            notifications.value[index] = {
              ...notifications.value[index],
              progress: (remaining / timeout) * 100,
            }
          }
        }
      }, interval)
    }
  }

  function extractAxiosErrorMessage(
    error: AxiosErrorLike,
    context?: string,
    showDetailed = true,
  ): string {
    if (!error.isAxiosError) {
      return `Unbekannter Fehler${context ? ` bei ${context}` : ''}`
    }

    if (showDetailed) {
      const serverMsg = error.response?.data?.message
      const statusText = error.response?.statusText
      const baseMsg = serverMsg || statusText || error.message || 'Unbekannter Fehler'

      return context ? `${context}: ${baseMsg}` : baseMsg
    }

    return context
      ? `Fehler ist aufgetreten${context ? ` bei ${context}` : ''}`
      : 'Ein Fehler ist aufgetreten'
  }

  function showError(
    error: Error | AxiosErrorLike,
    context?: string,
    showDetailed = true,
    timeout = 5000,
  ) {
    const message =
      error instanceof Error && (error as AxiosErrorLike).isAxiosError
        ? extractAxiosErrorMessage(error as AxiosErrorLike, context, showDetailed)
        : `Fehler${context ? ` bei ${context}` : ''}: ${error.message || error}`

    show(message, 'error', timeout)
  }

  function showWarn(message: string, timeout = 4000) {
    show(message, 'warning', timeout)
  }

  function dismiss(id: number) {
    notifications.value = notifications.value.filter((n) => n.id !== id)
  }

  async function updateOnlineStatus() {
    const wasOffline = isOffline.value
    isOffline.value = !navigator.onLine

    if (wasOffline && !isOffline.value) {
      show('Sie sind wieder online!', 'success', 5000)
      await checkBackendHealth()
    }
  }

  async function checkBackendHealth() {
    if (isOffline.value) {
      backendDown.value = false
      return
    }

    try {
      const res = await api.serviceHealthCheck()
      const wasDown = backendDown.value
      backendDown.value = !(res.status >= 200 && res.status < 300)

      if (wasDown && !backendDown.value) {
        show('Der Backend-Dienst ist wieder verfügbar.', 'success', 5000)
      }
    } catch (err: unknown) {
      const wasDown = backendDown.value
      backendDown.value = true
      if (!wasDown) {
        if (err instanceof Error) {
          show('Der Backend-Dienst ist nicht erreichbar.', 'error', 5000)
        }
      }
    }
  }

  function startMonitoring(intervalMs = 10000) {
    window.addEventListener('online', updateOnlineStatus)
    window.addEventListener('offline', updateOnlineStatus)

    updateOnlineStatus()
    checkBackendHealth()

    intervalId = setInterval(checkBackendHealth, intervalMs)
  }

  function stopMonitoring() {
    window.removeEventListener('online', updateOnlineStatus)
    window.removeEventListener('offline', updateOnlineStatus)
    if (intervalId) clearInterval(intervalId)
  }

  return {
    notifications,
    show,
    showError,
    showWarn,
    dismiss,
    isOffline,
    backendDown,
    startMonitoring,
    stopMonitoring,
  }
})
