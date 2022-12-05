<template>
  <div v-if="!jwt">
    <LoginWindow @jwt="setJwtToken"/>
  </div>
  <div v-if="jwt">
    <MainView :name="this.name" :role="this.role" :jwt="this.jwt"/>
  </div>
</template>

<script>
import LoginWindow from './components/LoginWindow.vue'
import MainView from './components/MainView.vue'

export default {
  data() {
    return {
      jwt: '',
      name: '',
      role: '',
    }
  },
  name: 'App',
  components: {
    LoginWindow,
    MainView
  },
  methods: {
    setJwtToken(token) {
      this.jwt = token
      this.setUserProperties(token)
    },
    setUserProperties(token) {
      var jwt = this.decodeJWT(token)
      this.name = jwt.name
      this.role = jwt.role
    },
    decodeJWT(token) {
      var base64Url = token.split('.')[1];
      var base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
      var jsonPayload = decodeURIComponent(window.atob(base64).split('').map(function(c) {
          return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
      }).join(''));
      return JSON.parse(jsonPayload);
    }
  }
}
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: left;
  color: #2c3e50;
  /* margin-top: 60px; */
}
body {
  background-color: #cfcfcf;
}
ul {
  list-style: none;
}
</style>
