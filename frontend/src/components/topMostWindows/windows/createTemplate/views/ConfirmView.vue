<template>
  <div class="innerWindowTemplate">
    <div style="height: 1.8em;">
      <label>Summary</label>
    </div>
    <TemplateSummary :info="this.info" style="height: 18.2em;"/>
    <p v-if="this.error != ''" class="errorText" style="margin-top: 0.2em;">{{this.error}}</p>
  </div>
</template>

<script>
import TemplateSummary from '../../../../template/TemplateSummary.vue'
export default {
  components: {TemplateSummary},
  props: ['info','error'],
}
</script>

<style>

</style>