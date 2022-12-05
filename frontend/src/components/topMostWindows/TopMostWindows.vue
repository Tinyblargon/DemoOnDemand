<template>
  <div style="left:0;top:0;height: 100%;width: 100%;background-color: rgba(128, 128, 128, 0.5);position: fixed;margin: 0 0 0 0;">
  </div>
  <div v-bind:style="{width:this.width,height:this.height,margin:this.margin,position:'absolute',left:'50%',top:'50%','background-color':'rgba(245, 245, 245, 1)',overflow:'hidden'}">
    <div style="display: flex;">
      <h1 style="margin: 0.3em; ">{{this.topWindow.text}}</h1>
      <button @click="emitClose('')" style="margin-left: auto; float: right;">close</button>
    </div>
    <CreateDemo v-if="this.topWindow.name == 'Demo'" :user="this.user" :role="this.role" :jwt="this.jwt" @close="emitClose"/>
    <CreateTemplate v-if="this.topWindow.name == 'Template'" :jwt="this.jwt" @close="emitClose"/>
    <DestroyConfirmation v-if="this.topWindow.name == 'DestroyConfirm'" :demo="this.demo" :role="this.role" :jwt="this.jwt" @close="emitClose"/>
    <TaskStatus v-if="this.topWindow.name == 'Task'" :task="this.task" :jwt="this.jwt"/>
  </div>
</template>

<script>
import CreateDemo from './windows/CreateDemo.vue';
import CreateTemplate from './windows/createTemplate/CreateTemplate.vue';
import DestroyConfirmation from './windows/DestroyConfirmation.vue';
import TaskStatus from './windows/TaskStatus.vue';
export default {
  components: {CreateDemo,CreateTemplate,DestroyConfirmation,TaskStatus},
  props: ['topWindow','task','demo','user','role','jwt'],
  emits: ["close"],
  data() {
    return {
      width: '',
      height: '',
      margin: '',
      format: 'em'
    }
  },
  created() {
    this.calcSize()
  },
  methods: {
    calcSize() {
      this.width = this.topWindow.width+this.format
      this.height = this.topWindow.height+this.format
      this.margin = '-'+this.topWindow.height/2+this.format+' 0 0 -'+this.topWindow.width/2+this.format
    },
    emitClose(close) {
      var window = {name: close};
      this.$emit('close', window)
    }
  }
}
</script>

<style>
/* style="overflow:hidden;" */
</style>