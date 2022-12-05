<template>
  <div v-if="demos.length">
    <perfect-scrollbar style="height: 27em;">
    
    <ul v-if="this.role == 'root'" id="demoList">
      <li v-for="demo in demos" :key="demo.user+'_'+demo.demo+'_'+demo.number" @click="setDemo(demo)" v-bind:class="{'active':(demo.user === this.demoSelected.user && demo.demo === this.demoSelected.demo && demo.number === this.demoSelected.number)}">
        <img v-if="demo.active" :src="this.image.running" alt="running icon"/>
        <img v-else :src="this.image.paused" alt="stopped icon"/>
        {{ demo.user }} {{ demo.demo }} {{ demo.number }}
      </li>
    </ul>
    <ul v-else id="demoList">
      <li v-for="demo in demos" :key="demo.user+'_'+demo.demo+'_'+demo.number" @click="setDemo(demo)" v-bind:class="{'active':(demo.user === this.demoSelected.user && demo.demo === this.demoSelected.demo && demo.number === this.demoSelected.number)}">
        <img v-if="demo.active" :src="this.image.running" alt="running icon"/>
        <img v-else :src="this.image.paused" alt="stopped icon"/>
        {{ demo.demo }} {{ demo.number }}
      </li>
    </ul>
    </perfect-scrollbar>
  </div>
</template>

<script>
import paused from "./../../assets/images/paused.png"
import running from "./../../assets/images/running.png"
export default {
  props: ['role','jwt'],
  emits: ['demo','demoSelected'],
  data() {
    return {
      demos: [],
      demo: '',
      demoSelected: '',
      image: {
        paused: paused,
        running: running,
      }
    }
  },
  created() {
    this.getDemoListRecursive()
  },
  mounted() {
  },
  methods: {
    getDemoListRecursive(){
      this.getDemoList()
      this.updateDemoStatus()
      setTimeout(() => {this.getDemoListRecursive()}, this.$timeout);
    },
    updateDemoStatus() {
      for (let i = 0; i < this.demos.length; i++) {
        var demo = this.demos[i]
        if (this.sameDemo(demo,this.demoSelected) && this.sameDemo(this.demo,this.demoSelected)) {
          this.$emit('demoSelected',demo)
        }
      }
    },
    getDemoList() {
      fetch(this.$url+'/demos', {
        method: 'GET',
        headers: {
          'Authorization': this.jwt
        }
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
          this.demos = data.data.demos
        })
        .catch(err => {
          console.log(err.message)
        })
    },
    setDemo(demo) {
      this.demoSelected = demo
      this.getDemoInfo(demo.user,demo.demo,demo.number)
    },
    getDemoInfo(user,demo,number) {
      fetch(this.$url+'/demos/'+user+'_'+number+'_'+demo, {
        method: 'GET',
        headers: {
          'Authorization': this.jwt
        }
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
        var demo = data.data.demo
        if (this.sameDemo(demo,this.demoSelected)) {
          this.demo = demo
          this.$emit('demo', this.demo)
          this.$emit('demoSelected',this.demoSelected)
        }
      })
      .catch(err => {
        console.log(err.message)
      })
    },
    sameDemo(demoA, demoB) {
      if (demoA.user === demoB.user && demoA.demo === demoB.demo && demoA.number === demoB.number) {
        return true
      }
      return false
    }
  }
}
</script>

<style>
.active{
  background-color: aqua;
}
#demoList {
  padding-left: 0.4em;
  margin: 0;
}
</style>
