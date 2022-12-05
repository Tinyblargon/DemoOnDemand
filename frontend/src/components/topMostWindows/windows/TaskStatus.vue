<template>
  <div>
    <button @click="setOutput(true)" :disabled='this.output'>Output</button>
    <button @click="setOutput(false)" :disabled='!this.output'>Status</button>
  </div>
  <div>
    <div v-if="this.output">
      <perfect-scrollbar style="height: 24em;">
        <tbody style="font-size: 0.6em;">
          <tr v-for="stat in status" :key="stat.kind+stat.text">
            <td>{{stat.kind}}</td>
            <td>{{stat.text}}</td>
          </tr>
        </tbody>
      </perfect-scrollbar>
    </div>
    <div v-else>
      <tr>
        <td>Status</td>
        <td>{{this.task.info.status}}</td>
      </tr>
      <tr>
        <td>User</td>
        <td>{{this.task.info.user}}</td>
      </tr>
      <tr>
        <td>Start Time</td>
        <td>{{this.task.info.time.start}}</td>
      </tr>
      <tr>
        <td>End Time</td>
        <td>{{this.task.info.time.end}}</td>
      </tr>
    </div>
  </div>
</template>

<script>
export default {
  props: ['task','jwt'],
  data() {
    return {
      output: true,
      status: [],
      active: true,
    }
  },
  created() {
    this.recursiveGetTaskStatus()
  },
  unmounted() {
    this.active = false
  },
  methods: {
    recursiveGetTaskStatus() {
      this.getTaskStatus()
      if (this.active) {
        setTimeout(() => {this.recursiveGetTaskStatus()}, this.$timeout);
      }
    },
    getTaskStatus() {
      fetch(this.$url+'/tasks/'+this.task.id, {
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
          this.status = data.data
        })
        .catch(err => {
          console.log(err.message)
        })
    },
    setOutput(bool) {
      this.output = bool
    }
  }
}
</script>

<style>

</style>