<template>
  <div class="buttonsTemplateTop">
    <button :disabled="(this.view.current == 1 || this.working)" @click="this.jumpToView(1)" class="buttonTemplateTop">General</button>
    <button :disabled="(this.view.current == 2 || this.view.completed < 1 || this.working)" @click="this.jumpToView(2)" class="buttonTemplateTop">Path</button>
    <button :disabled="(this.view.current == 3 || this.view.completed < 2 || this.working)" @click="this.jumpToView(3)" class="buttonTemplateTop">Network</button>
    <button :disabled="(this.view.current == 4 || this.view.completed < 3 || this.working)" @click="this.jumpToView(4)" class="buttonTemplateTop">Port</button>
    <button :disabled="(this.view.current == 5 || this.view.completed < 4 || this.working)" @click="this.jumpToView(5)" class="buttonTemplateTop">Confirm</button>
  </div>
  <div class="innerWindow" style="height: 22em;">
    <GeneralView v-if="(this.view.current == 1)" :templates="this.templates" :name="this.info.name" :description="this.info.description" @state="stateGeneral" />
    <PathView v-else-if="(this.view.current == 2)" :jwt="this.jwt" :currentPath="this.info.path" :currentNetworks="this.networks" @state="statePath"/>
    <NetworkView v-else-if="(this.view.current == 3)" :newNetworks="this.networks" :currentNetworks="this.tmp.networks" @state="stateNetwork"/>
    <PortView v-else-if="(this.view.current == 4)" :currentPortForwards="this.tmp.portForwards" @state="statePort"/>
    <ConfirmView v-else :info="this.info" :error="this.error"/>
  </div>
  <div>
    <button v-if="(this.view.current != 5)" :disabled="(this.view.completed < this.view.current) || this.working" @click="this.incrementView()" class="buttonTemplateBottom">Next</button>
    <button v-else @click="this.createTemplate()" class="buttonTemplateBottom">Finish</button>
    <button :disabled="(this.view.current == 1 || this.working)" @click="this.decrementView()" class="buttonTemplateBottom">Back</button>
  </div>
</template>

<script>
import GeneralView from './views/GeneralView.vue'
import PathView from './views/PathView.vue'
import NetworkView from "./views/NetworkView.vue"
import PortView from "./views/PortView.vue"
import ConfirmView from "./views/ConfirmView.vue"
import api from '../../../../api.js'
export default {
  components: {GeneralView,PathView,NetworkView,PortView,ConfirmView},
  props: ['jwt'],
  emits: ["close"],
  data() {
    return {
      info: {
        name: '',
        description: '',
        path: '',
        networks: [],
        portforwards: [],
      },
      view: {
        current: 1,
        completed: 0,
      },
      templates: [],
      networks: [],
      tmp: {
        networks: [],
        portForwards: [],
      },
      ready: {
        general: false,
        path: false,
        network: false,
        port: false,
      },
      working: false,
      error: '',
    }
  },
  created() {
    api.GetTemplates(this.$url,this.jwt).then(obj => {
      this.templates = obj
    })
  },
  methods: {
    stateGeneral(state) {
      this.info.name = state.info.name
      this.info.description = state.info.description
      this.ready.general = state.ready
      this.setReadyView()
    },
    statePath(state) {
      this.info.path = state.info.path
      this.networks = state.info.networks
      this.working = state.working
      this.ready.path = state.ready
      this.setReadyView()
    },
    stateNetwork(state) {
      this.tmp.networks = state.info.networks
      this.ready.network = state.ready
      this.setReadyView()
    },
    statePort(state) {
      this.tmp.portForwards = state.info.portForwards
      this.ready.port = state.ready
      this.setReadyView()
    },
    formatInfo() {
      this.info.networks = new Array(this.tmp.networks.length)
      for (let i = 0; i < this.tmp.networks.length; i++) {
        this.info.networks[i] = {
          name: this.tmp.networks[i].name,
          cidr: this.tmp.networks[i].cidr,
        }
      }
      this.info.portforwards = new Array(this.tmp.portForwards.length)
      for (let i = 0; i < this.tmp.portForwards.length; i++) {
        this.info.portforwards[i] = {
          sourceport: parseInt(this.tmp.portForwards[i].sourcePort.value),
          destinationport: this.tmp.portForwards[i].destinationPort.value,
          destinationip: this.tmp.portForwards[i].ip.value,
          protocol: this.tmp.portForwards[i].protocol.value,
          comment: this.tmp.portForwards[i].comment,
        }
        if (this.info.portforwards[i].destinationport == '') {
          this.info.portforwards[i].destinationport = 0
        } else {
          this.info.portforwards[i].destinationport = parseInt(this.info.portforwards[i].destinationport)
        }
      }
    },
    createTemplate() {
      fetch(this.$url+'/templates', {
          method: 'POST',
          body: JSON.stringify(this.info),
          headers: {
            'Authorization': this.jwt
          }
        })
        .then((response) => {
          if (response.ok) {
            this.error = ''
            this.emitClose()
          }
          this.error = 'Error while making request'
        })
    },
    incrementView() {
      this.jumpToView(this.view.current+1)
    },
    decrementView() {
      let currentView = this.view.current
      if (currentView > 0) {
        this.view.current = currentView - 1
      }
    },
    jumpToView(view) {
      if (view === 5) {
        this.formatInfo()
      }
      this.view.current = view
    },
    setReadyView() {
      if (this.ready.port) {
        this.readyView(4)
      } else if (this.ready.network) {
        this.readyView(3)
      } else if (this.ready.path) {
        this.readyView(2)
      } else if (this.ready.general) {
        this.readyView(1)
      } else {
        this.readyView(0)
      }
    },
    readyView(view){
      this.view.completed = view
    },
    emitClose() {
      this.$emit('close', '')
    },
  }
}
</script>

<style>
.buttonTemplateTop {
  width: 14%;
  margin-left: 3%;
  margin-right: 3%;
}
.buttonTemplateBottom {
  position: relative;
  float: right;
  width: 5em;
}
.innerWindowTemplate {
  margin-left: 1%;
  margin-right: 1%;
}
</style>