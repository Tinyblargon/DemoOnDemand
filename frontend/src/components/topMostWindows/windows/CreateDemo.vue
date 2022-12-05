<template>
  <div class="innerWindow" style="height: 26em;">
    <div style="overflow: hidden;height: 91%;padding-left: 1em;">
      <div style="float: left;height:auto;width: 15em;">
        <h2>Templates</h2>
        <perfect-scrollbar style="height: 20em;">
          <ul id="templateList">
            <li v-for="template in templates" :key="template.name" @click="setTemplate(template)" v-bind:class="{'active':(this.template===template.name)}">{{template.name}}</li>
          </ul>
        </perfect-scrollbar>
      </div>
      <div style="float: left;margin-left: 1em;">
        <h2>Description</h2>
        <p>{{templateDescription}}</p>
      </div>
    </div>
    <div style="display: flex;">
      <input v-if="this.role == 'root'" type="text" required v-model="userName" style="margin-left: auto;width: 5em;background-color: yellow;">
      <button :disabled="this.template == ''" @click="this.createDemo(this.userName,this.template,this.number)" style="margin-left: auto;">Create</button>
    </div>
  </div>
</template>

<script>
import api from '../../../api.js'
export default {
  props: ['user','role','jwt'],
  emits: ["close"],
  data() {
    return {
      templates: [],
      template: '',
      templateDescription: '',
      userName: ''
    }
  },
  created() {
    api.GetTemplates(this.$url,this.jwt).then(obj => {
      this.templates = obj
    })
  },
  methods: {
    setTemplate(template) {
      this.template = template.name
      this.templateDescription = template.description
    },
    createDemo(name,template) {
      if (name == '') {
        name = this.user
      }
      fetch(this.$url+'/demos', {
        method: 'POST',
        body: JSON.stringify({template: template, username: name}),
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
  }
}
</script>

<style>
#templateList{
  padding-left: 0.4em;
  margin: 0;
}
</style>