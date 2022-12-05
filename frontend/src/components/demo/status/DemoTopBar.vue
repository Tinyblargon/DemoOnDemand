<template>
  <div v-if="this.demo" style="display: flex; height: 1.7em;">
    <p style="margin: 0.5em 0.4em;font-weight: bold;font-size: large">
      <span style="font-size: small;">{{this.demo.number}}</span>
      {{this.demo.demo}}</p>
    <div style="margin-left: auto;">
    <button :disabled='this.status' @click="startTask('start')">
      <span><!-- icon --></span>
      <span>start</span></button>
    <button :disabled='!this.status' @click="startTask('stop')">
      <span><!-- icon --></span>
      <span>stop  </span></button>
    <button @click="startTask('restart')">
      <span><!-- icon --></span>
      <span>restart</span></button>
    <button @click="emitWindow('DestroyConfirm','Confirm',20,12)">
      <span><!-- icon --></span>
      <span>destroy</span></button>
  </div>
</div>
</template>

<script>
export default {
  props: ['status','demo','jwt'],
  emits: ["window"],
  methods: {
    startTask(task) {
      fetch(this.$url+'/demos/'+this.demo.user+'_'+this.demo.number+'_'+this.demo.demo, {
      method: 'PUT',
      body: JSON.stringify({task: task}),
      headers: {
        'Authorization': this.jwt
      }
    })
      .then(res => res.text())
      .catch(err => {
        console.log(err.message)
      })
    },
    emitWindow(name,text,width,height) {
      var window = {
        name: name,
        text: text,
        width: width,
        height: height
      };
      this.$emit('window', window)
    } 
  }

}
</script>

<style>
button {
  margin: 0.5em;
  margin-left: auto; /* Push this element to the right */
}
</style>