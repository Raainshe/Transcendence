import '@fontsource/press-start-2p/index.css'
import '@/assets/styles/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { watch } from 'vue'

import App from './App.vue'
import { i18n } from './i18n'
import router from './router'
import { useAuthStore } from './stores/auth'
import { useGameSettingsStore } from './stores/gameSettings'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)

void useAuthStore(pinia).hydrate()
app.use(i18n as never)
app.use(router)

const settings = useGameSettingsStore(pinia)
i18n.global.locale.value = settings.locale

watch(
  () => settings.locale,
  (locale) => {
    i18n.global.locale.value = locale
    document.documentElement.lang = locale
    const routeName = router.currentRoute.value.name
    const routeKey = typeof routeName === 'string' ? routeName : 'home'
    document.title = i18n.global.t(`routes.${routeKey}`)
  },
  { immediate: true },
)

router.afterEach((to) => {
  const routeKey = typeof to.name === 'string' ? to.name : 'home'
  document.title = i18n.global.t(`routes.${routeKey}`)
})

app.mount('#app')
