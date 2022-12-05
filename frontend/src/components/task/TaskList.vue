<template>
  <div style="top: 30.3em;bottom: 0.4em;position: absolute;background-color: white;">
    <perfect-scrollbar v-if="this.tasks" style="height: 100%;">
      <table style="width: 100%;table-layout: fixed;">
        <thead>
          <th>Start Time:</th>
          <th>End Time:</th>
          <th>Task Number:</th>
          <th v-if="this.role == 'root'">User:</th>
          <th>Status:</th>
        </thead>
        <tbody>
          <tr v-for="task in this.computedTasks" :key="task.id" v-bind:class="{'active':(task.id == this.activeTaskID)}">
            <td @click="setTask(task)">{{task.info.time.start}}</td>
            <td @click="setTask(task)">{{task.info.time.end}}</td>
            <td @click="setTask(task)">{{task.id}}</td>
            <td v-if="this.role == 'root'" @click="setTask(task)">{{task.info.user}}</td>
            <td @click="setTask(task)">{{task.info.status}}</td>
          </tr>
        </tbody>
      </table>
    </perfect-scrollbar>
  </div>
</template>

<script>
export default {
  props: ['role','jwt'],
  emits: ['task'],
  data() {
    return {
      tasks: [],
      activeTaskID: 0,
    }
  },
  created() {
    this.listAllTasksRecursive()
  },
  computed: {
    computedTasks() {
      let computed = this.tasks
      for (let i = 0; i < computed.length; i++) {
        computed[i].info.time.start = this.formatDateTime(computed[i].info.time.start*1000)
        if (typeof computed[i].info.time.end !== 'undefined') {
          computed[i].info.time.end = this.formatDateTime(computed[i].info.time.end*1000)
        }
      }
      return computed
    },
  },
  methods: {
    listAllTasksRecursive() {
      this.listAllTasks()
      setTimeout(() => {this.listAllTasksRecursive()}, this.$timeout);
    },
    listAllTasks() {
      fetch(this.$url+'/tasks', {
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
        this.tasks = data.data.tasks.sort((a,b) => b.id - a.id)
      })
      .catch(err => {
        console.log(err.message)
      })
    },
    formatDateTime(timestamp) {
      var d = new Date(timestamp)
      let month
      switch(d.getMonth()) {
        case 1:
          month = "Jan"
          break;
        case 2:
          month = "Feb"
          break;
        case 3:
          month = "Mar"
          break
        case 4:
          month = "Apr"
          break;
        case 5:
          month = "May"
          break;
        case 6:
          month = "Jun"
          break;
        case 7:
          month = "Jul"
          break;
        case 8:
          month = "Aug"
          break;
        case 9:
          month = "Sep"
          break;
        case 10:
          month = "Oct"
          break;
        case 11:
          month = "Nov"
          break;
        case 12:
          month = "Dec"
          break;
      }
      let dayNumber = d.getDay()
      let day
      if (dayNumber < 10) {
        day = '0' + dayNumber
      }
      return month + ' ' + day + ' ' + d.getHours() + ':' + d.getMinutes() + ':' + d.getSeconds()
    },
    setTask(task) {
      this.activeTaskID = task.id
      this.$emit('task', task)
    }
  }
}
</script>

<style>

</style>