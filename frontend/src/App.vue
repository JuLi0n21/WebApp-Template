<script setup lang="ts">
import { RouterView } from 'vue-router'
import Notification from '@/reusables/Notifcation.vue'
import { useAuthStore } from '@/stores/auth'
import { onMounted, onBeforeUnmount } from 'vue'
import { useNotificationStore } from '@/stores/notificationStore'

const notification = useNotificationStore()

onMounted(() => {
  const auth = useAuthStore()
  auth.restoreSession()
  notification.startMonitoring()
})

onBeforeUnmount(() => {
  notification.stopMonitoring()
})
</script>

<template>
  <div class="flex flex-col bg-stone-950 h-screen text-white">
    <Notification />
    <div class="overflow-hidden grow">
      <RouterView> </RouterView>
    </div>
  </div>
</template>
