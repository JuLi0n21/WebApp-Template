import { defineStore } from 'pinia'

interface AuthState {
  token: string | null
  user: any | null // Replace `any` with your user type if needed
}

function isTokenExpired(token: string): boolean {
  try {
    const payloadBase64 = token.split('.')[1]
    const payloadJson = atob(payloadBase64)
    const payload = JSON.parse(payloadJson)

    if (!payload.exp) return false // No exp claim = can't check expiry

    const now = Math.floor(Date.now() / 1000) // current time in seconds
    return payload.exp < now
  } catch (e) {
    console.error('Invalid token format:', e)
    return true // Treat as expired if invalid
  }
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: null,
    user: null,
  }),
  getters: {
    isAuthenticated: (state): boolean => !!state.token && !isTokenExpired(state.token),
  },
  actions: {
    setToken(token: string | undefined) {
      if (typeof token === 'string' && !isTokenExpired(token)) {
        this.token = token
        localStorage.setItem('token', token)
      } else {
        this.token = null
        localStorage.removeItem('token')
      }
    },
    logout() {
      this.token = null
      this.user = null
      localStorage.removeItem('token')
    },
    restoreSession() {
      const token = localStorage.getItem('token')
      if (token && !isTokenExpired(token)) {
        this.token = token
      } else {
        this.token = null
        localStorage.removeItem('token')
      }
    },
  },
})
