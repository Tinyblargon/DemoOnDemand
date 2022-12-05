<template>
  <div class="innerWindowTemplate">
    <label>Name</label>
    <input type="text" required v-model="info.name" @blur="this.checkTemplateExistence(info.name)" />
    <label>Description</label>
    <input type="text" required v-model="info.description" @blur="this.emit()"/>
  </div>
</template>

<script>
import util from './../../../../../util.js'
export default {
  props: ['templates','name','description'],
  emits: ["state"],
  data() {
    return {
      info: {
        name: '',
        description: '',
      },
      templateExists: false,
      ready: false,
    }
  },
  created() {
    this.info.name = this.name
    this.info.description = this.description
  },
  methods: {
    checkTemplateExistence(name) {
      let existingTemplates = this.templates
      let localError = false
      for (let i = 0; i < existingTemplates.length; i++) {
        if (existingTemplates[i].name === name) {
          localError = true
        }
      }
      this.templateExists = localError
      this.emit()
    },
    checkReady() {
      if (this.info.name == '') {
        this.ready = false
        return
      }
      this.ready = util.BoolInvert(this.templateExists)
    }, 
    emit() {
      this.checkReady()
      this.$emit('state', {
        info: this.info,
        ready: this.ready,
      })
    }
  }
}
</script>

<style>

</style>