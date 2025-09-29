<template>
  <h2 class="p-4 text-2xl text-center">Login</h2>
  <form @submit.prevent="handleLogin" class="flex flex-col space-y-4 m-4 p-4 rounded-sm text-white">
    <input
      class="bg-green-400/40 p-4 border-3 rounded-lg"
      v-model="username"
      placeholder="Username"
      required
    />
    <input
      class="bg-green-400/40 p-4 border-3 rounded-lg"
      v-model="password"
      type="password"
      placeholder="Password"
      required
    />
    <button class="bg-green-400/40 p-4 border-3 rounded-lg" type="submit">Anmelden</button>
  </form>
  <p v-if="error" class="text-red">{{ error }}</p>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/services/ApiService'

const username = ref('')
const password = ref('')
const error = ref<string | null>(null)

const router = useRouter()
const auth = useAuthStore()
const api = useApi()

const handleLogin = async () => {
  try {
    const res = await api.serviceLogin({
      username: username.value,
      password: password.value,
    })

    auth.setToken(res.data.token)
    router.push({ name: 'home' })
  } catch (err) {
    console.error('Login failed:', err)
    error.value = 'Falsche Anmeldedaten. Bitte versuche es erneut.'
  }
}
</script>
