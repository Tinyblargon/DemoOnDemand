<template>
  <div class="innerWindowTemplate">
    <label>DemoPath</label>
    <input type="text" required v-model="this.path.current" :disabled="this.networks.scanning" @blur="this.checkPath()"/>
    <button :disabled="(this.networks.scanning || this.ready === true)" @click="this.getNetworks(this.path.current)">Scan</button>
    <p v-if="this.error != ''" style="padding-top: 12em;" class="errorText">{{this.error}}</p>
    <perfect-scrollbar v-else style="height: 14.2em;">
      <ul v-if="(this.networks.data.length && this.path.current === this.path.scanned)">
        <li v-for="network in this.networks.data" :key="network">{{network}}</li>
      </ul>
    </perfect-scrollbar>
  </div>
</template>

<script>
export default {
  props: ['jwt','currentPath','currentNetworks'],
  emits: ["state"],
  data() {
    return {
      ready: false,
      error: '',
      path: {
        current: '',
        scanned: '',
      },
      networks: {
        data: [],
        scanning: false,
      },
    }
  },
  created() {
    this.path.current = this.currentPath
    this.networks.data = this.currentNetworks
    if (this.currentNetworks.length > 0) {
      this.path.scanned = this.currentPath
      this.ready = true
    }
  },
  methods: {
    getNetworks(path) {
      this.networks.scanning = true
      this.emit()
      fetch(this.$url+'/networks', {
        method: 'POST',
        headers: {'Authorization': this.jwt},
        body: JSON.stringify({path: path})
      })
        .then(res => res.text())
        .then(body => {
          try {
            return JSON.parse(body);
          } catch {
            throw Error(body)
          }
        })
        .then(data => {
          this.error = ''
          this.networks.data = data.data.networks.sort()
          this.path.scanned = path
          this.networks.scanning = false
          this.ready = true
          this.emit()
        })
        .catch(err => {
          this.error = err.message
          this.networks.scanning = false
          this.ready = false
          this.emit()
        })
    },
    checkPath() {
      if ((this.path.current === this.path.scanned || this.path.scanned === '')&& this.networks.data.length > 0) {
        this.ready = true
      } else {
        this.ready = false
      }
    },
    emit() {
      this.$emit('state', {
        info: {
          path: this.path.scanned,
          networks: this.networks.data,
        },
        ready: this.ready,
        working: this.networks.scanning,
      })
    }
  }
}
</script>

<style>

</style>