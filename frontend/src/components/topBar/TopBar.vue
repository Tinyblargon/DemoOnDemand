<template>
  <div style="display: flex; height: 2.5em;">
    <h2 style="  margin: 0.1em 0.4em 0 0">User: {{this.name}}</h2>
    <div style="margin-left: auto;">
      <button v-if="(this.role === 'root'||this.role === 'admin')" @click="emitWindow('Template','Create: Template',60,30)">
        <span><!-- icon --></span>
        <span>Create Template</span></button>
      <button @click="emitWindow('Demo','Create: Demo',50,30)">
        <span><!-- icon --></span>
        <span>Create Demo</span></button>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      topWindow: {
        name: '',
        text: '',
        width: 0,
        height: 0,
      },
    }
  },
  props: ['name','role'],
  methods: {
    emitWindow(name,text,width,height) {
      this.topWindow.name = name
      this.topWindow.text = text
      this.topWindow.width = width
      this.topWindow.height = height
      this.$emit('window', this.topWindow)
    } 
  }
}
</script>

<style>

</style>