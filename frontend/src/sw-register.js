import { registerSW } from 'virtual:pwa-register'

registerSW({
    immediate: true, // take over instantly
    onOfflineReady() {
        console.log('App is ready to work offline.')
    },
})
