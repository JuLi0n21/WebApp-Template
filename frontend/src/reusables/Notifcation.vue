<template>
  <div
    class="z-100 fixed w-full notification-container"
    role="region"
    aria-live="polite"
    aria-atomic="true"
  >
    <transition-group name="fade" tag="div" class="flex flex-col space-y-2 w-full">
      <div
        v-for="note in notificationStore.notifications"
        :key="note.id"
        role="alert"
        tabindex="0"
        @click="notificationStore.dismiss(note.id)"
        aria-label="Benachrichtigung schließen"
        :class="[
          'notification flex items-center space-x-3 shadow-md p-3 rounded-md relative',
          note.type,
        ]"
      >
        <svg
          v-if="note.type === 'success'"
          class="flex-shrink-0 w-5 h-5 text-green-300"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          viewBox="0 0 24 24"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="M5 13l4 4L19 7" />
        </svg>
        <svg
          v-else-if="note.type === 'error'"
          class="flex-shrink-0 w-5 h-5 text-red-400"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          viewBox="0 0 24 24"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <line x1="18" y1="6" x2="6" y2="18" />
          <line x1="6" y1="6" x2="18" y2="18" />
        </svg>
        <svg
          v-else-if="note.type === 'warning'"
          class="flex-shrink-0 w-5 h-5 text-yellow-400"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          viewBox="0 0 24 24"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path
            d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"
          />
          <line x1="12" y1="9" x2="12" y2="13" />
          <line x1="12" y1="17" x2="12" y2="17" />
        </svg>
        <svg
          v-else
          class="flex-shrink-0 w-5 h-5 text-blue-400"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          viewBox="0 0 24 24"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <circle cx="12" cy="12" r="10" />
          <line x1="12" y1="16" x2="12" y2="12" />
          <line x1="12" y1="8" x2="12" y2="8" />
        </svg>

        <p class="flex-1 w-full font-semibold text-white text-sm break-words leading-snug">
          {{ note.message }}
        </p>
        <div class="progress-bar" :style="{ width: note.progress + '%' }" aria-hidden="true"></div>
      </div>
    </transition-group>
  </div>
</template>

<script setup lang="ts">
import { useNotificationStore } from '@/stores/notificationStore'

const notificationStore = useNotificationStore()
</script>

<style scoped>
.notification-container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-height: 80vh; /* optional: avoid too tall container on mobile */
  overflow-y: auto;
}

.notification {
  color: white;
  font-weight: 600;
  cursor: pointer;
  user-select: none;
  word-break: break-word;
}

.progress-bar {
  position: absolute;
  bottom: 0;
  left: 0;
  height: 3px;
  background-color: rgba(255, 255, 255, 0.7);
  transition: width 0.05s linear;
  border-radius: 0 0 4px 4px;
}

.notification.info {
  background-color: #2196f3;
}
.notification.success {
  background-color: #4caf50;
}
.notification.warning {
  background-color: #ff9800;
}
.notification.error {
  background-color: #f44336;
}

.fade-enter-active,
.fade-leave-active {
  transition:
    opacity 0.3s ease,
    transform 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
