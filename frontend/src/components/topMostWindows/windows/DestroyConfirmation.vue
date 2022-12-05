<template>
  <p style="text-align: center;">Are you sure you want to destroy:<br>({{this.demoName}})?</p>
  <button @click="destroyDemo(this.demo.user,this.demo.demo,this.demo.number)">Destroy</button>
</template>

<script>
export default {
  props: ['demo','role','jwt'],
  emits: ['close'],
  data() {
    return {
      demoName: ''
    }
  },
  created() {
    this.setDemoName()
  },
  methods: {
    destroyDemo(user,demo,number) {
      fetch(this.$url+'/demos/'+user+'_'+number+'_'+demo, {
        method: 'DELETE',
        headers: {
          'Authorization': this.jwt
        }
      })
      .then(res => {
        res.text()
        this.$emit('close', '')
      })
      .catch(err => {
        console.log(err.message)
      })
    },
    setDemoName() {
      this.demoName = this.demo.demo+' '+this.demo.number
      if (this.role == 'root') {
        this.demoName = this.demo.name+' '+this.demoName
      }
    }
  }
}
</script>

<style>

</style>