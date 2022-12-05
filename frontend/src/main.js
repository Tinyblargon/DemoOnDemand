import { createApp } from 'vue'
import App from './App.vue'
import PerfectScrollbar from 'vue3-perfect-scrollbar'
import 'vue3-perfect-scrollbar/dist/vue3-perfect-scrollbar.css'

import '../src/assets/css/main.css';

const app = createApp(App)

// app.config.globalProperties.$jwt = ''
app.config.globalProperties.$url = process.env.VUE_APP_ROOT_API
app.config.globalProperties.$timeout = 1500


app.use(PerfectScrollbar)
app.mount('#app')

