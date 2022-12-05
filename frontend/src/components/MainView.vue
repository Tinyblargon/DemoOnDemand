<template>
  <TopBar :name="this.name" :role="this.role" @window="setTopWindow"/>
  <div style="height: 27em;">
    <div class="demo-items border-box" style="width: 25%;float: left;">
      <DemoList @demoSelected="setDemoStatus" @demo="setDemo" :role="this.role" :jwt="this.jwt"/>
    </div>
    <div style="float: left;width: 0.4em;height: 100%;">
    </div>
    <div class="demo-items border-box">
      <DemoStatus v-if="this.demo" @window="setTopWindow" :demoSelected="this.demoSelected" :demo="this.demo" :jwt="this.jwt"/>
    </div>
  </div>
  <TaskList @task="setTask" :role="this.role" :jwt="this.jwt"/>
  <TopMostWindows v-if="this.topWindow.name != ''" :topWindow="this.topWindow" :task="this.task" :demo="this.demo" :user="this.name" :role="this.role" :jwt="this.jwt" @close="setTopWindow"/>
</template>

<script>
import DemoList from './demo/DemoList.vue';
import DemoStatus from './demo/DemoStatus.vue';
import TaskList from './task/TaskList.vue';
import TopBar from './topBar/TopBar.vue';
import TopMostWindows from './topMostWindows/TopMostWindows.vue';
export default {
  components: {DemoList,DemoStatus,TaskList,TopBar,TopMostWindows},
  data() {
    return {
      topWindow: {
        name: '',
        text: '',
        width: 0,
        height: 0,
      },
      demo: null,
      demoSelected: null,
      task: null,
    }
  },
  props: ['name','role','jwt'],
  methods: {
      setDemo(demo) {
        this.demo = demo
      },
      setDemoStatus(demo) {
        this.demoSelected = demo
      },
      setTopWindow(window) {
        this.topWindow = window
      },
      setTask(task) {
        this.task = task
        this.setTopWindow({
          name: 'Task',
          text: 'Task '+task.id,
          height: 30,
          width: 55,
        })
      }
  }
}
</script>

<style>
.demo-items {
  height: 100%;
  overflow: hidden;
  background-color: white;
  text-align: left;
}
.border-box {
  -webkit-box-sizing: border-box;
  box-sizing: border-box;
}
</style>