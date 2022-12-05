<template>
  <form @submit.prevent="getJwtToken">
    <label>Username:</label>
    <input type="username" required v-model="username">
    <label>Password:</label>
    <input type="password" required v-model="password">
    <!-- <div v-if="passwordError" class="error">{{ passwordError }}</div> -->
    <button>Login</button>
  </form>
</template>

<script>
export default {
  data() {
    return {
      username: '',
      password: '',
      passwordError: '',
      jwt: ''
    }
  },
  methods: {
    getJwtToken() {
      fetch(this.$url+'/login', {
        method: 'POST',
        body: JSON.stringify({username: this.username, password: this.password})
      })
        .then(res => res.json())
        .then(data => {
          this.jwt = data.data.token
          this.$emit('jwt', this.jwt)
        })
        .catch(err => console.log(err.message))
    }
  }
}
</script>

<style>
form {
  max-width: 420px;
  margin: 30px auto;
  background: white;
  text-align: left;
  padding: 40px;
  border-radius: 10px;
}
label {
  color: #aaa;
  display: inline-block;
  margin: 25px 0 15px;
  font-size: 0.6em;
  text-transform: uppercase;
  letter-spacing: 1px;
  font-weight: bold;
}
input, select {
  display: block;
  padding: 10px 6px;
  width: 100%;
  box-sizing: border-box;
  border: none;
  border-bottom: 1px solid #ddd;
  color: #555;
}
</style>