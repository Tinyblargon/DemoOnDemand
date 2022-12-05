<template>
  <div class="innerWindowTemplate">
    <perfect-scrollbar style="height: 20em;">
      <table>
        <tbody>
          <tr v-for="network in this.networks" :key="network.name">
            <td>{{network.name}}</td>
            <td>
              <input type="text" required v-model="network.cidr" @blur="this.checkCIDR(network)"/>
            </td>
          </tr>
        </tbody>
      </table>
    </perfect-scrollbar>
    <p v-if="this.error" class="errorText">Invalid CIDR</p>
  </div>
</template>

<script>
import cidrRegex from "cidr-regex"
export default {
  props: ['jwt','newNetworks','currentNetworks'],
  emits: ["state"],
  data() {
    return {
      ready: false,
      networks: [
        {
          name: '',
          cidr: '',
          error: false,
        }
      ],
      error: false,
    }
  },
  created() {
    if (this.newNetworks.length == this.currentNetworks.length) {
      let same = true
      for (let i = 0; i < this.newNetworks; i++) {
        for (let ii = 0; ii < this.currentNetworks; ii++) {
          if (this.networks[i] == this.currentNetworks.name) {
            same = true
            break
          }
          same = false
        }
      }
      if (same) {
        this.loadNetworks()
      } else {
        this.initializeNetworks()
      }
    } else {
      this.initializeNetworks()
    }
  },
  methods: {
    initializeNetworks() {
      let networks = new Array(this.newNetworks.length);
      for (let i = 0; i < networks.length; i++) {
        networks[i] = {
          name: this.newNetworks[i],
          cidr: '',
          error: false,
        }
      }
      this.networks = networks
      return networks
    },
    loadNetworks() {
      let networks = new Array(this.currentNetworks.length);
      for (let i = 0; i < networks.length; i++) {
        networks[i] = this.currentNetworks[i]
      }
      this.networks = networks
      this.globalError()
    },
    checkCIDR(network) {
      if (cidrRegex({exact: true}).test(network.cidr) || cidrRegex.v6({exact: true}).test(network.cidr) || network.cidr === '') {
        network.error = false
      } else {
        network.error = true
      }
      this.emit()
    },
    globalError() {
      for (let i = 0; i < this.networks.length; i++) {
        if (this.networks[i].error) {
          this.error = true
          return
        }
      }
      this.error = false
    },
    setReady() {
      if (this.error) {
        this.ready = false
      } else {
        this.ready = this.cidrFilledOut()
      }
    },
    cidrFilledOut() {
      for (let i = 0; i < this.networks.length; i++) {
        if (this.networks[i].cidr === '') {
          return false
        }
      }
      return true
    },
    emit() {
      this.globalError()
      this.setReady()
      this.$emit('state', {
        info: {
          networks: this.networks,
        },
        ready: this.ready,
      })
    }
  }
}
</script>

<style>

</style>