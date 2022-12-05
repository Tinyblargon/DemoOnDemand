<template>
  <TopBar :status="this.demoSelected.active" :demo="this.demo" :jwt="this.jwt" @window="emitWindow"/>
  <div style="display: flex; margin-top: 1em;">
    <div>
      <ul id="sideBar">
        <li @click="setMode('')"><span></span><span>Summary</span></li>
        <li @click="setMode('portForward')"><span></span><span>PortForwards</span></li>
      </ul>
    </div>
    <div style="margin-left: 1em;">
      <div v-if="this.mode == 'portForward'">
        <div v-if="demo.portforwards">
          <div v-for="portForward in demo.portforwards" :key="portForward.sourceport+portForward.protocol">
            <tbody>
              <tr>
                <td>
                  {{portForward.sourceport}}
                </td>
                <td>
                  {{portForward.protocol}}
                </td>
                <td>
                  {{portForward.comment}}
                </td>
              </tr>
            </tbody>
          </div>
        </div>
      </div>
      <div v-else>
        <p v-if="this.demoSelected.active" class="infoLine">Status: running</p>
        <p v-else class="infoLine">Status: stopped</p>
        <p class="infoLine">IP: {{this.demo.ip}}</p>
        <p class="infoLine">Description: {{this.demo.description}}</p>
        <!-- <p class="infoLine">Comment: {{this.demo.description}}</p> -->
    </div>
    </div>
  </div>
</template>

<script>
import TopBar from './status/DemoTopBar.vue';
export default {
  components: {
    TopBar,
  },
  emits: ["window"],
  data() {
    return {mode: ''}
  },
  props: ['demoSelected','demo','jwt'],
  methods: {
    setMode(mode) {
      this.mode = mode
    },
    emitWindow(window) {
      this.$emit('window', window)
    }
  }
}
</script>

<style>
#sideBar {
  padding-left: 0.4em;
  margin: 0;
}
.infoLine {
  padding: 0;
  margin: 0;
}
</style>