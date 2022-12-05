<template>
  <div style="top: 30.3em;bottom: 0.4em;position: absolute;background-color: white;">
    <perfect-scrollbar v-if="this.tasks" style="height: 100%;">
      <table v-for="task in tasks" :key="task.id" v-bind:class="{'active':(task.id == this.activeTaskID)}">
        <tbody>
          <tr>
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
        this.tasks = data.data.tasks
      })
      .catch(err => {
        console.log(err.message)
      })
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